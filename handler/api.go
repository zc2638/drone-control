/**
 * Created by zc on 2020/9/5.
 */
package handler

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/zc2638/drone-control/global"
	"github.com/zc2638/drone-control/global/route"
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

	r.Get(route.APIRouteRepo, api.RepoList())
	r.Post(route.APIRouteRepo, api.RepoCreate())
	r.Get(route.APIRouteRepoInfoPath, api.RepoInfoBySlug())
	r.Get(route.APIRouteRepoPath, api.RepoInfo())
	r.Put(route.APIRouteRepoPath, api.RepoUpdate())
	r.Delete(route.APIRouteRepoPath, api.RepoDelete())

	r.Get(route.APIRouteBuild, api.BuildList())
	r.Post(route.APIRouteBuild, api.Trigger())
	r.Get(route.APIRouteBuildPath, api.BuildInfo())
	r.Get(route.APIRouteStepLogPath, api.BuildLog())

	r.Route(route.APIRouteStream, func(cr chi.Router) {
		cr.Get("/{repo}/{build}/{stage}/{step}", api.BuildLogStream())
	})
	return r
}

func RPC() http.Handler {
	r := chi.NewRouter()
	r.Use(
		middleware.Recoverer,
		middleware.Logger,
		middleware.NoCache,
		authorization(),
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

func authorization() func(http.Handler) http.Handler {
	token := global.Cfg().RPC.Secret
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
