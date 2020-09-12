/**
 * Created by zc on 2020/9/6.
 */
package api

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/drone/drone-yaml/yaml"
	"github.com/drone/drone/core"
	"github.com/drone/drone/trigger/dag"
	"github.com/go-chi/chi"
	"github.com/pkgms/go/ctr"
	"github.com/zc2638/drone-control/global"
	"github.com/zc2638/drone-control/store"
	"net/http"
	"strconv"
	"time"
)

type Req struct {
	RepoID int64          `json:"repo_id"`
	Spec   *yaml.Pipeline `json:"spec"`
}

func Trigger() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.ParseInt(
			chi.URLParam(r, "repo"), 10, 64)
		if id < 1 {
			ctr.BadRequest(w, errors.New("cannot find repo, repo_id is invalid"))
			return
		}

		var repo store.ReposData
		db := global.GormDB().Where(&store.ReposData{ID: id}).First(&repo)
		if db.Error != nil {
			ctr.BadRequest(w, db.Error)
			return
		}
		last, err := store.BuildStore().FindRef(context.Background(), repo.ID, "")
		if err != nil && err != sql.ErrNoRows {
			ctr.BadRequest(w, errors.New("trigger: cannot increment build sequence"))
			return
		}
		var number int64 = 1
		if last != nil {
			number = last.Number + 1
		}
		dd := dag.New()
		matched, err := parsePipeline([]byte(repo.Data), dd)
		if err != nil {
			ctr.BadRequest(w, err)
			return
		}
		now := time.Now().Local()
		build := &core.Build{
			RepoID:       repo.ID,
			Number:       number,
			Status:       core.StatusPending,
			Created:      now.Unix(),
			Updated:      now.Unix(),
			Title:        "build",
			Message:      "build #" + strconv.FormatInt(number, 10),
			Author:       repo.Username,
			Sender:       repo.Username,
			Source:       "source",
			Target:       "target",
			Before:       "before",
			After:        "after",
			Trigger:      "trigger",
			Event:        "event",
			Action:       "action",
			AuthorAvatar: "http://localhost:2639/static/image/avatar.jpg",
		}

		stages := make([]*core.Stage, len(matched))
		for i, match := range matched {
			onSuccess := match.Trigger.Status.Match(core.StatusPassing)
			onFailure := match.Trigger.Status.Match(core.StatusFailing)
			if len(match.Trigger.Status.Include)+len(match.Trigger.Status.Exclude) == 0 {
				onFailure = false
			}

			stage := &core.Stage{
				RepoID:    repo.ID,
				Number:    i + 1,
				Name:      match.Name,
				Kind:      match.Kind,
				Type:      match.Type,
				OS:        match.Platform.OS,
				Arch:      match.Platform.Arch,
				Variant:   match.Platform.Variant,
				Kernel:    match.Platform.Version,
				Limit:     match.Concurrency.Limit,
				Status:    core.StatusWaiting,
				DependsOn: match.DependsOn,
				OnSuccess: onSuccess,
				OnFailure: onFailure,
				Labels:    match.Node,
				Created:   now.Unix(),
				Updated:   now.Unix(),
			}
			completionStage(stage)
			if len(stage.DependsOn) == 0 {
				stage.Status = core.StatusPending
			}
			stages[i] = stage
		}

		for _, stage := range stages {
			// here we re-work the dependencies for the stage to
			// account for the fact that some steps may be skipped
			// and may otherwise break the dependency chain.
			stage.DependsOn = dd.Dependencies(stage.Name)

			// if the stage is pending dependencies, but those
			// dependencies are skipped, the stage can be executed
			// immediately.
			if stage.Status == core.StatusWaiting && len(stage.DependsOn) == 0 {
				stage.Status = core.StatusPending
			}
		}

		ctx := r.Context()
		if err = store.BuildStore().Create(ctx, build, stages); err != nil {
			fmt.Println("trigger: cannot create build")
			ctr.BadRequest(w, err)
			return
		}

		if err = global.Scheduler().Schedule(ctx, nil); err != nil {
			fmt.Println("trigger: cannot enqueue build")
			ctr.BadRequest(w, errors.New("trigger: cannot enqueue build"))
			return
		}
		ctr.Success(w)
	}
}
