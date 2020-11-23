/**
 * Created by zc on 2020/9/5.
 */
package global

import (
	"database/sql"
	dronedb "github.com/drone/drone/store/shared/db"
	"github.com/zc2638/drone-control/global/database"
	"gorm.io/gorm"
)

func initDatabase(cfg *database.Config) error {
	var err error
	if sqlDB, err = database.New(cfg); err != nil {
		return err
	}
	if gormDB, err = database.NewGormDB(cfg, DB()); err != nil {
		return err
	}
	droneDB, err = database.NewDroneDB(cfg, DB())
	return err
}

var sqlDB *sql.DB

func DB() *sql.DB {
	return sqlDB
}

var gormDB *gorm.DB

func GormDB() *gorm.DB {
	return gormDB
}

var droneDB *dronedb.DB

func DroneDB() *dronedb.DB {
	return droneDB
}
