/**
 * Created by zc on 2020/9/5.
 */
package rpc

import (
	"context"
	"encoding/json"
	"github.com/drone/drone/core"
	"github.com/drone/drone/operator/manager"
	"github.com/drone/drone/store/shared/db"
	"github.com/go-chi/chi"
	"github.com/pkgms/go/ctr"
	"github.com/sirupsen/logrus"
	"github.com/zc2638/drone-control/global"
	"github.com/zc2638/drone-control/store"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// Join returns an http.HandlerFunc that makes an
// http.Request to join the cluster.
func Join() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
}

// Leave returns an http.HandlerFunc that makes an
// http.Request to leave the cluster.
func Leave() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
}

// Ping returns an http.HandlerFunc that makes an
// http.Request to ping the server and confirm connectivity.
func Ping() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
}

// Request returns an http.HandlerFunc that processes an
// http.Request to reqeust a stage from the queue for execution.
func Request() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx, cancel := context.WithTimeout(ctx, time.Second*30)
		defer cancel()

		args := new(manager.Request)
		err := json.NewDecoder(r.Body).Decode(args)
		if err != nil {
			ctr.BadRequest(w, err)
			return
		}
		logger := logrus.WithFields(
			logrus.Fields{
				"kind":    args.Kind,
				"type":    args.Type,
				"os":      args.OS,
				"arch":    args.Arch,
				"kernel":  args.Kernel,
				"variant": args.Variant,
			},
		)
		logger.Debugln("manager: request queue item")

		stage, err := global.Scheduler().Request(ctx, core.Filter{
			Kind:    args.Kind,
			Type:    args.Type,
			OS:      args.OS,
			Arch:    args.Arch,
			Kernel:  args.Kernel,
			Variant: args.Variant,
			Labels:  args.Labels,
		})
		if err != nil && ctx.Err() != nil {
			logger.Debugln("manager: context canceled")
			ctr.BadRequest(w, err)
			return
		}
		if err != nil {
			logger = logrus.WithError(err)
			logger.Warnln("manager: request queue item error")
			ctr.BadRequest(w, err)
			return
		}
		ctr.OK(w, stage)
	}
}

// Accept returns an http.HandlerFunc that processes an
// http.Request to accept ownership of the stage.
func Accept() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.ParseInt(
			chi.URLParam(r, "stage"), 10, 64)
		machine := r.FormValue("machine")
		logger := logrus.WithFields(
			logrus.Fields{
				"stage-id": id,
				"machine":  machine,
			},
		)
		logger.Debugln("manager: accept stage")

		stageStore := store.StageStore()
		stage, err := stageStore.Find(context.Background(), id)
		if err != nil {
			logger = logger.WithError(err)
			logger.Warnln("manager: cannot find stage")
			ctr.BadRequest(w, err)
			return
		}
		if stage.Machine != "" {
			logger.Debugln("manager: stage already assigned. abort.")
			ctr.BadRequest(w, db.ErrOptimisticLock)
			return
		}

		stage.Machine = machine
		stage.Status = core.StatusPending
		stage.Updated = time.Now().Unix()

		err = stageStore.Update(context.Background(), stage)
		if err == db.ErrOptimisticLock {
			logger = logger.WithError(err)
			logger.Debugln("manager: stage processed by another agent")
			ctr.BadRequest(w, err)
			return
		}
		if err != nil {
			logger = logger.WithError(err)
			logger.Debugln("manager: cannot update stage")
			ctr.BadRequest(w, err)
			return
		}
		logger.Debugln("manager: stage accepted")
		ctr.OK(w, stage)
	}
}

type details struct {
	*manager.Context
	Netrc *core.Netrc `json:"netrc"`
	Repo  *repositroy `json:"repository"`
}
type repositroy struct {
	*core.Repository
	Secret string `json:"secret"`
}

// Info returns an http.HandlerFunc that processes an
// http.Request to get the build details.
func Info() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.ParseInt(
			chi.URLParam(r, "stage"), 10, 64)

		res, err := Detail(context.Background(), id)
		if err != nil {
			ctr.BadRequest(w, err)
			return
		}
		// TODO config add
		res.System = &core.System{
			Proto: "http",
			Host:  "127.0.0.1",
		}
		ctr.OK(w, &details{
			Context: res,
			Repo: &repositroy{
				Repository: res.Repo,
			},
		})
	}
}

// UpdateStage returns an http.HandlerFunc that processes
// an http.Request to update a stage.
func UpdateStage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dst := new(core.Stage)
		err := json.NewDecoder(r.Body).Decode(dst)
		if err != nil {
			ctr.BadRequest(w, err)
			return
		}
		if dst.Status == core.StatusPending || dst.Status == core.StatusRunning {
			err = BeforeAll(context.Background(), dst)
		} else {
			err = AfterAll(context.Background(), dst)
		}
		if err != nil {
			ctr.BadRequest(w, err)
			return
		}
		ctr.OK(w, dst)
	}
}

// UpdateStep returns an http.HandlerFunc that processes
// an http.Request to update a step.
func UpdateStep() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dst := new(core.Step)
		err := json.NewDecoder(r.Body).Decode(dst)
		if err != nil {
			ctr.BadRequest(w, err)
			return
		}
		if len(dst.Error) > 500 {
			dst.Error = dst.Error[:500]
		}
		if err := store.StepStore().Update(context.Background(), dst); err != nil {
			ctr.BadRequest(w, err)
			return
		}
		ctr.OK(w, dst)
	}
}

// Watch returns an http.HandlerFunc that accepts a
// blocking http.Request that watches a build for cancellation
// events.
func Watch() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx, cancel := context.WithTimeout(ctx, time.Second*30)
		defer cancel()

		id, _ := strconv.ParseInt(
			chi.URLParamFromCtx(ctx, "build"), 10, 64)

		// we expect a context cancel error here which
		// indicates a polling timeout. The subscribing
		// client should look for the context cancel error
		// and resume polling.
		if _, err := global.Scheduler().Cancelled(ctx, id); err != nil {
			ctr.BadRequest(w, err)
			return
		}
		ctr.Success(w)
	}
}

// LogBatch returns an http.HandlerFunc that accepts an
// http.Request to submit a stream of logs to the system.
func LogBatch() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		step, _ := strconv.ParseInt(
			chi.URLParam(r, "step"), 10, 64)

		var lines []*core.Line
		err := json.NewDecoder(r.Body).Decode(&lines)
		if err != nil {
			ctr.BadRequest(w, err)
			return
		}
		for _, line := range lines {
			_ = global.LogStream().Write(r.Context(), step, line)
		}
		ctr.Success(w)
	}
}

// LogUpload returns an http.HandlerFunc that accepts an
// http.Request to upload and persist logs for a pipeline stage.
func LogUpload() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.ParseInt(
			chi.URLParam(r, "step"), 10, 64)

		dir := filepath.Join(global.PathStageLog)
		if _, err := os.Stat(dir); err != nil {
			if !os.IsNotExist(err) {
				ctr.BadRequest(w, err)
				return
			}
			if err := os.MkdirAll(dir, os.ModePerm); err != nil {
				ctr.BadRequest(w, err)
				return
			}
		}
		fp := filepath.Join(dir, strconv.FormatInt(id, 10))
		f, err := os.OpenFile(fp, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0766)
		if err != nil {
			ctr.BadRequest(w, err)
			return
		}
		defer f.Close()
		if _, err := io.Copy(f, r.Body); err != nil {
			ctr.BadRequest(w, err)
			return
		}
		ctr.Success(w)
	}
}
