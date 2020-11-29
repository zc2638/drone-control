/**
 * Created by zc on 2020/9/5.
 */
package store

import (
	"github.com/drone/drone/core"
	"github.com/drone/drone/store/build"
	"github.com/drone/drone/store/repos"
	"github.com/drone/drone/store/stage"
	"github.com/drone/drone/store/step"
	"github.com/zc2638/drone-control/global"
)

func RepoStore() core.RepositoryStore {
	return repos.New(global.DroneDB())
}

func BuildStore() core.BuildStore {
	return build.New(global.DroneDB())
}

func StageStore() core.StageStore {
	return stage.New(global.DroneDB())
}

func StepStore() core.StepStore {
	return step.New(global.DroneDB())
}
