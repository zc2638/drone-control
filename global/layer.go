/**
 * Created by zc on 2020/9/5.
 */
package global

import (
	"github.com/drone/drone/core"
	"github.com/drone/drone/livelog"
	"github.com/drone/drone/scheduler/queue"
	dronedb "github.com/drone/drone/store/shared/db"
	"github.com/drone/drone/store/stage"
)

var config *Config

func InitCfg(cfg *Config) error {
	config = cfg
	if err := initDatabase(&cfg.Database); err != nil {
		return err
	}
	InitScheduler(DroneDB())
	return nil
}

func Cfg() *Config {
	return config
}

var logStream = livelog.New()

func LogStream() core.LogStream {
	return logStream
}

var queues core.Scheduler

func InitScheduler(db *dronedb.DB) {
	queues = queue.New(stage.New(db))
}

func Scheduler() core.Scheduler {
	return queues
}
