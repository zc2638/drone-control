/**
 * Created by zc on 2020/11/23.
 */
package database

import (
	"database/sql"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func NewGormDB(cfg *Config, db *sql.DB) (*gorm.DB, error) {
	driver := dealDriver(cfg.Driver)
	var dbConnect *gorm.DB
	var err error
	switch driver {
	case "mysql":
		dbConnect, err = mysqlLink(db)
	case "postgres":
		dbConnect, err = postgresLink(db)
	default:
		dbConnect, err = sqliteLink(db)
	}
	if err != nil {
		return nil, err
	}
	if cfg.Debug {
		dbConnect = dbConnect.Debug()
	}
	return dbConnect, nil
}

func mysqlLink(db *sql.DB) (*gorm.DB, error) {
	dialector := mysql.New(mysql.Config{
		Conn: db,
	})
	return gorm.Open(
		dialector,
		&gorm.Config{
			SkipDefaultTransaction: true,
			NamingStrategy: schema.NamingStrategy{
				SingularTable: true,
			},
		},
	)
}

func postgresLink(db *sql.DB) (*gorm.DB, error) {
	dialector := postgres.New(postgres.Config{
		Conn: db,
	})
	return gorm.Open(
		dialector,
		&gorm.Config{
			SkipDefaultTransaction: true,
			NamingStrategy: schema.NamingStrategy{
				SingularTable: true,
			},
		},
	)
}

func sqliteLink(db *sql.DB) (*gorm.DB, error) {
	dbConnect, err := gorm.Open(
		nil,
		&gorm.Config{
			SkipDefaultTransaction: true,
			NamingStrategy: schema.NamingStrategy{
				SingularTable: true,
			},
			ConnPool: db,
		},
	)
	if err != nil {
		return nil, err
	}
	dialector := sqlite.Dialector{}
	if err := dialector.Initialize(dbConnect); err != nil {
		return nil, err
	}
	dbConnect.Dialector = dialector
	return dbConnect, nil
}
