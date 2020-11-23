/**
 * Created by zc on 2020/9/6.
 */
package rpc

import (
	"context"
	"github.com/drone/drone/core"
	"github.com/drone/drone/operator/manager"
	"github.com/drone/drone/store/shared/db"
	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"
	"github.com/zc2638/drone-control/global"
	"github.com/zc2638/drone-control/handler/api"
	"github.com/zc2638/drone-control/store"
	"path"
	"time"
)

func Detail(ctx context.Context, id int64) (*manager.Context, error) {
	logger := logrus.WithField("step-id", id)
	logger.Debugln("detail: fetching stage details")

	buildStore := store.BuildStore()
	stageStore := store.StageStore()

	stage, err := stageStore.Find(context.Background(), id)
	if err != nil {
		logger = logger.WithError(err)
		logger.Warnln("detail: cannot find stage")
		return nil, err
	}
	build, err := buildStore.Find(context.Background(), stage.BuildID)
	if err != nil {
		logger = logger.WithError(err)
		logger.Warnln("detail: cannot find build")
		return nil, err
	}
	stages, err := stageStore.List(ctx, stage.BuildID)
	if err != nil {
		logger = logger.WithError(err)
		logger.Warnln("detail: cannot list stages")
		return nil, err
	}
	build.Stages = stages

	var repo store.ReposData
	if err := global.GormDB().Where(&store.ReposData{
		ID: build.RepoID,
	}).First(&repo).Error; err != nil {
		logger = logger.WithError(err)
		logger.Warnln("detail: cannot find repo")
		return nil, err
	}
	config, err := api.CheckClone(repo.Data)
	if err != nil {
		logger = logger.WithError(err)
		logger.Warnln("detail: cannot compile config spec")
	}
	return &manager.Context{
		Build: build,
		Stage: stage,
		Repo: &core.Repository{
			Namespace: repo.Namespace,
			Name:      repo.Name,
			Slug:      path.Join(repo.Namespace, repo.Name),
			Signer:    repo.Username,
			Branch:    "default",
			Timeout:   repo.Timeout,
			Created:   repo.Created,
			Updated:   repo.Updated,
			Active:    true,
			Build:     build,
		},
		Config: &core.File{
			Data: config,
		},
	}, nil
}

func BeforeAll(ctx context.Context, stage *core.Stage) error {
	logger := logrus.WithField("stage.id", stage.ID)

	buildStore := store.BuildStore()
	stageStore := store.StageStore()
	stepStore := store.StepStore()

	build, err := buildStore.Find(context.Background(), stage.BuildID)
	if err != nil {
		logger.WithError(err).Warnln("manager: cannot find the build")
		return err
	}
	logger = logger.WithFields(
		logrus.Fields{
			"build.number": build.Number,
			"build.id":     build.ID,
			"stage.id":     stage.ID,
			"repo.id":      build.RepoID,
		},
	)
	// // note that if multiple stages run concurrently it will attempt
	// // to create the watcher multiple times. The watcher is responsible
	// // for handling multiple concurrent requests and preventing duplication.
	// err = s.Watcher.Register(noContext, build.ID)
	// if err != nil {
	// 	logger.WithError(err).Warnln("manager: cannot create the watcher")
	// 	return err
	// }

	if len(stage.Error) > 500 {
		stage.Error = stage.Error[:500]
	}
	stage.Updated = time.Now().Unix()
	err = stageStore.Update(context.Background(), stage)
	if err != nil {
		logger.WithError(err).
			WithField("stage.status", stage.Status).
			Warnln("manager: cannot update the stage")
		return err
	}

	for _, step := range stage.Steps {
		if len(step.Error) > 500 {
			step.Error = step.Error[:500]
		}
		err := stepStore.Create(context.Background(), step)
		if err != nil {
			logger.WithError(err).
				WithField("stage.status", stage.Status).
				WithField("step.name", step.Name).
				WithField("step.id", step.ID).
				Warnln("manager: cannot persist the step")
			return err
		}
	}

	if build.Status != core.StatusPending {
		return nil
	}
	build.Started = time.Now().Unix()
	build.Updated = time.Now().Unix()
	build.Status = core.StatusRunning
	err = buildStore.Update(context.Background(), build)
	if err == db.ErrOptimisticLock {
		return nil
	}
	if err != nil {
		logger.WithError(err).Warnln("manager: cannot update the build")
		return err
	}
	return nil
}

func AfterAll(ctx context.Context, stage *core.Stage) error {
	noContext := context.Background()
	buildStore := store.BuildStore()
	stageStore := store.StageStore()
	stepStore := store.StepStore()

	logger := logrus.WithField("stage.id", stage.ID)
	logger.Debugln("manager: stage is complete. teardown")

	build, err := buildStore.Find(noContext, stage.BuildID)
	if err != nil {
		logger.WithError(err).Warnln("manager: cannot find the build")
		return err
	}

	logger = logger.WithFields(
		logrus.Fields{
			"build.number": build.Number,
			"build.id":     build.ID,
			"repo.id":      build.RepoID,
		},
	)

	for _, step := range stage.Steps {
		if len(step.Error) > 500 {
			step.Error = step.Error[:500]
		}
		err := stepStore.Update(noContext, step)
		if err != nil {
			logger.WithError(err).
				WithField("stage.status", stage.Status).
				WithField("step.name", step.Name).
				WithField("step.id", step.ID).
				Warnln("manager: cannot persist the step")
			return err
		}
	}

	if len(stage.Error) > 500 {
		stage.Error = stage.Error[:500]
	}

	stage.Updated = time.Now().Unix()
	err = stageStore.Update(noContext, stage)
	if err != nil {
		logger.WithError(err).
			Warnln("manager: cannot update the stage")
		return err
	}

	stages, err := stageStore.ListSteps(noContext, build.ID)
	if err != nil {
		logger.WithError(err).Warnln("manager: cannot get stages")
		return err
	}

	//
	//
	//

	err = cancelDownstream(ctx, stages)
	if err != nil {
		logger.WithError(err).
			Errorln("manager: cannot cancel downstream builds")
		return err
	}

	err = scheduleDownstream(ctx, stage, stages)
	if err != nil {
		logger.WithError(err).
			Errorln("manager: cannot schedule downstream builds")
		return err
	}

	//
	//
	//

	if !isBuildComplete(stages) {
		logger.Debugln("manager: build pending completion of additional stages")
		return nil
	}

	logger.Debugln("manager: build is finished, teardown")

	build.Status = core.StatusPassing
	build.Finished = time.Now().Unix()
	for _, sibling := range stages {
		if sibling.Status == core.StatusKilled {
			build.Status = core.StatusKilled
			break
		}
		if sibling.Status == core.StatusFailing {
			build.Status = core.StatusFailing
			break
		}
		if sibling.Status == core.StatusError {
			build.Status = core.StatusError
			break
		}
	}
	if build.Started == 0 {
		build.Started = build.Finished
	}

	err = buildStore.Update(noContext, build)
	if err == db.ErrOptimisticLock {
		logger.WithError(err).
			Warnln("manager: build updated by another goroutine")
		return nil
	}
	if err != nil {
		logger.WithError(err).
			Warnln("manager: cannot update the build")
		return err
	}
	return nil
}

func isBuildComplete(stages []*core.Stage) bool {
	for _, stage := range stages {
		switch stage.Status {
		case core.StatusPending,
			core.StatusRunning,
			core.StatusWaiting,
			core.StatusDeclined,
			core.StatusBlocked:
			return false
		}
	}
	return true
}

func cancelDownstream(
	ctx context.Context,
	stages []*core.Stage,
) error {
	failed := false
	for _, s := range stages {
		if s.IsFailed() {
			failed = true
		}
	}

	var errs error
	for _, s := range stages {
		if s.Status != core.StatusWaiting {
			continue
		}

		var skip bool
		if failed && !s.OnFailure {
			skip = true
		}
		if !failed && !s.OnSuccess {
			skip = true
		}
		if !skip {
			continue
		}

		if !areDepsComplete(s, stages) {
			continue
		}

		logger := logrus.WithFields(
			logrus.Fields{
				"stage.id":         s.ID,
				"stage.on_success": s.OnSuccess,
				"stage.on_failure": s.OnFailure,
				"stage.is_failure": failed,
				"stage.depends_on": s.DependsOn,
			},
		)
		logger.Debugln("manager: skipping step")

		s.Status = core.StatusSkipped
		s.Started = time.Now().Unix()
		s.Stopped = time.Now().Unix()
		err := store.StageStore().Update(context.Background(), s)
		if err == db.ErrOptimisticLock {
			resync(ctx, s)
			continue
		}
		if err != nil {
			logger.WithError(err).
				Warnln("manager: cannot update stage status")
			errs = multierror.Append(errs, err)
		}
	}
	return errs
}

func scheduleDownstream(
	ctx context.Context,
	stage *core.Stage,
	stages []*core.Stage,
) error {

	var errs error
	for _, sibling := range stages {
		if sibling.Status == core.StatusWaiting {
			if len(sibling.DependsOn) == 0 {
				continue
			}

			// PROBLEM: isDep only checks the direct parent
			// i think ....
			// if isDep(stage, sibling) == false {
			// 	continue
			// }
			if !areDepsComplete(sibling, stages) {
				continue
			}
			// if isLastDep(stage, sibling, stages) == false {
			// 	continue
			// }

			logger := logrus.WithFields(
				logrus.Fields{
					"stage.id":         sibling.ID,
					"stage.name":       sibling.Name,
					"stage.depends_on": sibling.DependsOn,
				},
			)
			logger.Debugln("manager: schedule next stage")

			sibling.Status = core.StatusPending
			sibling.Updated = time.Now().Unix()
			err := store.StageStore().Update(context.Background(), sibling)
			if err == db.ErrOptimisticLock {
				resync(ctx, sibling)
				continue
			}
			if err != nil {
				logger.WithError(err).
					Warnln("manager: cannot update stage status")
				errs = multierror.Append(errs, err)
			}

			err = global.Scheduler().Schedule(context.Background(), sibling)
			if err != nil {
				logger.WithError(err).
					Warnln("manager: cannot schedule stage")
				errs = multierror.Append(errs, err)
			}
		}
	}
	return errs
}

func resync(ctx context.Context, stage *core.Stage) error {
	updated, err := store.StageStore().Find(ctx, stage.ID)
	if err != nil {
		return err
	}
	stage.Status = updated.Status
	stage.Error = updated.Error
	stage.ExitCode = updated.ExitCode
	stage.Machine = updated.Machine
	stage.Started = updated.Started
	stage.Stopped = updated.Stopped
	stage.Created = updated.Created
	stage.Updated = updated.Updated
	return nil
}

func areDepsComplete(stage *core.Stage, stages []*core.Stage) bool {
	deps := map[string]struct{}{}
	for _, dep := range stage.DependsOn {
		deps[dep] = struct{}{}
	}
	for _, sibling := range stages {
		if _, ok := deps[sibling.Name]; !ok {
			continue
		}
		if !sibling.IsDone() {
			return false
		}
	}
	return true
}
