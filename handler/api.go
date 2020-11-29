/**
 * Created by zc on 2020/9/5.
 */
package handler

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/zc2638/drone-control/global"
	"github.com/zc2638/drone-control/handler/api"
	"github.com/zc2638/drone-control/handler/public"
	"github.com/zc2638/drone-control/handler/rpc"
	"net/http"
)

const (
	PathRepo         = "/repo"
	PathRepoSlug     = "/repo/{namespace}/{name}"
	PathBuildSlug    = "/build/{namespace}/{name}"
	PathBuildSlugNum = "/build/{namespace}/{name}/{build}"
	PathBuildLogSlug = "/build/{namespace}/{name}/{build}/log/{stage}/{step}"
	PathStreamLog    = "/stream/{repo}/{build}/{stage}/{step}"
)

func API() http.Handler {
	r := chi.NewRouter()
	r.Use(
		middleware.Recoverer,
		middleware.Logger,
	)

	r.Get(PathRepo, api.RepoList())
	r.Post(PathRepo, api.RepoApply())
	r.Get(PathRepoSlug, api.RepoInfoBySlug())
	r.Delete(PathRepoSlug, api.RepoDelete())

	r.Get(PathBuildSlug, api.BuildList())
	r.Post(PathBuildSlug, api.Trigger())
	r.Get(PathBuildSlugNum, api.BuildInfo())
	r.Get(PathBuildLogSlug, api.BuildLog())

	r.Get(PathStreamLog, api.BuildLogStream())
	return r
}

func Static() http.Handler {
	r := chi.NewRouter()
	r.Handle("/image/avatar.jpg", public.Image())
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
			//if token == "" {
			//	w.WriteHeader(403)
			//}
			if token == r.Header.Get("X-Drone-Token") {
				next.ServeHTTP(w, r)
			} else {
				w.WriteHeader(401)
			}
		})
	}
}
