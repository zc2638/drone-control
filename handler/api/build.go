/**
 * Created by zc on 2020/9/10.
 */
package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/pkgms/go/ctr"
	"github.com/zc2638/drone-control/global"
	"github.com/zc2638/drone-control/store"
	"io"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strconv"
	"time"
)

func BuildList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.ParseInt(
			chi.URLParam(r, "repo"), 10, 64)
		if id < 1 {
			ctr.BadRequest(w, errors.New("cannot find repo, repo_id is invalid"))
			return
		}
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		size, _ := strconv.Atoi(r.URL.Query().Get("size"))
		if page < 1 {
			page = 1
		}
		if size < 1 {
			size = 10
		}

		list, err := store.BuildStore().List(context.Background(), id, size, (page-1)*size)
		if err != nil {
			ctr.BadRequest(w, err)
			return
		}
		ctr.OK(w, list)
	}
}

func BuildInfo() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.ParseInt(
			chi.URLParam(r, "repo"), 10, 64)
		if id < 1 {
			ctr.BadRequest(w, errors.New("cannot find repo, repo_id is invalid"))
			return
		}
		buildNumber, _ := strconv.ParseInt(
			chi.URLParam(r, "build"), 10, 64)

		build, err := store.BuildStore().FindNumber(context.Background(), id, buildNumber)
		if err != nil {
			ctr.BadRequest(w, err)
			return
		}
		stages, err := store.StageStore().ListSteps(context.Background(), build.ID)
		if err != nil {
			ctr.BadRequest(w, err)
			return
		}
		build.Stages = stages
		ctr.OK(w, build)
	}
}

func BuildLog() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.ParseInt(
			chi.URLParam(r, "repo"), 10, 64)
		if id < 1 {
			ctr.BadRequest(w, errors.New("cannot find repo, repo_id is invalid"))
			return
		}
		buildNumber, _ := strconv.ParseInt(
			chi.URLParam(r, "build"), 10, 64)
		stageNumber, _ := strconv.Atoi(
			chi.URLParam(r, "stage"))
		stepNumber, _ := strconv.Atoi(
			chi.URLParam(r, "step"))

		build, err := store.BuildStore().FindNumber(context.Background(), id, buildNumber)
		if err != nil {
			ctr.BadRequest(w, err)
			return
		}
		stage, err := store.StageStore().FindNumber(context.Background(), build.ID, stageNumber)
		if err != nil {
			ctr.BadRequest(w, err)
			return
		}
		step, err := store.StepStore().FindNumber(context.Background(), stage.ID, stepNumber)
		if err != nil {
			ctr.BadRequest(w, err)
			return
		}
		fp := filepath.Join(global.PathStageLog, strconv.FormatInt(step.ID, 10))
		data, err := ioutil.ReadFile(fp)
		if err != nil {
			fmt.Println("read log error: ", err)
			ctr.BadRequest(w, errors.New("read log error"))
			return
		}
		ctr.Bytes(w, data)
	}
}

func BuildLogStream() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.ParseInt(
			chi.URLParam(r, "repo"), 10, 64)
		if id < 1 {
			ctr.BadRequest(w, errors.New("cannot find repo, repo_id is invalid"))
			return
		}
		buildNumber, _ := strconv.ParseInt(
			chi.URLParam(r, "build"), 10, 64)
		stageNumber, _ := strconv.Atoi(
			chi.URLParam(r, "stage"))
		stepNumber, _ := strconv.Atoi(
			chi.URLParam(r, "step"))

		build, err := store.BuildStore().FindNumber(r.Context(), id, buildNumber)
		if err != nil {
			ctr.NotFound(w, err)
			return
		}
		stage, err := store.StageStore().FindNumber(r.Context(), build.ID, stageNumber)
		if err != nil {
			ctr.NotFound(w, err)
			return
		}
		step, err := store.StepStore().FindNumber(r.Context(), stage.ID, stepNumber)
		if err != nil {
			ctr.NotFound(w, err)
			return
		}

		h := w.Header()
		h.Set("Content-Type", "text/event-stream")
		h.Set("Cache-Control", "no-cache")
		h.Set("Connection", "keep-alive")
		h.Set("X-Accel-Buffering", "no")

		f, ok := w.(http.Flusher)
		if !ok {
			return
		}

		io.WriteString(w, ": ping\n\n")
		f.Flush()

		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()

		enc := json.NewEncoder(w)
		linec, errc := global.LogStream().Tail(ctx, step.ID)
		if errc == nil {
			io.WriteString(w, "event: error\ndata: eof\n\n")
			return
		}

	L:
		for {
			select {
			case <-ctx.Done():
				break L
			case <-errc:
				break L
			case <-time.After(time.Hour):
				break L
			case <-time.After(time.Second * 30):
				io.WriteString(w, ": ping\n\n")
			case line := <-linec:
				io.WriteString(w, "data: ")
				enc.Encode(line)
				io.WriteString(w, "\n\n")
				f.Flush()
			}
		}

		io.WriteString(w, "event: error\ndata: eof\n\n")
		f.Flush()
	}
}
