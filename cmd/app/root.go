/*
Copyright © 2020 zc

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package app

import (
	"github.com/go-chi/chi"
	"github.com/pkgms/go/ctr"
	"github.com/pkgms/go/server"
	"github.com/spf13/cobra"
	"github.com/zc2638/drone-control/global"
	"github.com/zc2638/drone-control/handler"
	"github.com/zc2638/drone-control/store"
	"net/http"
	"os"
)

var cfgFile string

func NewServerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "control",
		Short: "drone control service",
		Long:  `Drone Control Service.`,
		RunE:  Run,
	}
	cfgFilePath := os.Getenv("LUBAN_CONFIG")
	if cfgFilePath == "" {
		cfgFilePath = "config.yaml"
	}
	cmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", cfgFilePath, "config file (default is $HOME/config.yaml)")
	return cmd
}

func Run(cmd *cobra.Command, args []string) error {
	cfg, err := global.ParseConfig(cfgFile)
	if err != nil {
		return err
	}
	if err := global.InitCfg(cfg); err != nil {
		return err
	}
	if err := Migrate(); err != nil {
		return err
	}
	s := server.New(&cfg.Server)
	s.Handler = Route()
	return s.Run()
}

func Route() http.Handler {
	r := chi.NewRouter()
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		ctr.OK(w, "Hello World!")
	})
	r.Mount("/rpc/v2", handler.RPC())
	r.Mount("/api", handler.API())
	r.Mount("/static", handler.Static())
	return r
}

func Migrate() error {
	if ok := global.GormDB().Migrator().HasTable(&store.ReposData{}); !ok {
		return global.GormDB().Migrator().CreateTable(&store.ReposData{})
	}
	return nil
}
