/**
 * Created by zc on 2020/9/5.
 */
package handler

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/zc2638/drone-control/handler/api"
	"github.com/zc2638/drone-control/handler/public"
	"github.com/zc2638/drone-control/handler/rpc"
	"net/http"
)

func Static() http.Handler {
	r := chi.NewRouter()
	r.Handle("/image/avatar.jpg", public.Image())
	return r
}

func API() http.Handler {
	r := chi.NewRouter()
	r.Use(
		middleware.Recoverer,
		middleware.Logger,
	)
	r.Route("/repo", func(cr chi.Router) {
		cr.Get("/", api.RepoList())
		cr.Post("/", api.RepoCreate())
		cr.Get("/info", api.RepoInfo())
		cr.Put("/{repo}", api.RepoUpdate())
		cr.Delete("/{repo}", api.RepoDelete())
		cr.Post("/{repo}/build", api.Trigger())
		cr.Get("/{repo}/build", api.BuildList())
		cr.Get("/{repo}/build/{build}", api.BuildInfo())
		cr.Get("/{repo}/build/{build}/log/{stage}/{step}", api.BuildLog())
	})
	r.Route("/stream", func(cr chi.Router) {
		cr.Get("/{repo}/{build}/{stage}/{step}", api.BuildLogStream())
	})
	return r
}

func RPC(secret string) http.Handler {
	r := chi.NewRouter()
	r.Use(
		middleware.Recoverer,
		middleware.Logger,
		middleware.NoCache,
		authorization(secret),
	)
	r.Post("/nodes/:machine", rpc.Join())
	r.Delete("/nodes/:machine", rpc.Leave())
	r.Post("/ping", rpc.Ping())
	r.Post("/stage", rpc.Request())
	r.Post("/stage/{stage}", rpc.Accept())
	r.Get("/stage/{stage}", rpc.Info())
	r.Put("/stage/{stage}", rpc.UpdateStage())
	r.Put("/step/{step}", rpc.UpdateStep())
	r.Post("/build/{build}/watch", rpc.Watch())
	r.Post("/step/{step}/logs/batch", rpc.LogBatch())
	r.Post("/step/{step}/logs/upload", rpc.LogUpload())
	return r
}

func authorization(token string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// prevents system administrators from accidentally
			// exposing drone without credentials.
			if token == "" {
				w.WriteHeader(403)
			} else if token == r.Header.Get("X-Drone-Token") {
				next.ServeHTTP(w, r)
			} else {
				w.WriteHeader(401)
			}
		})
	}
}
