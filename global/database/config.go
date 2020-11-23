/**
 * Created by zc on 2020/11/23.
 */
package database

import (
	"database/sql"
	"strings"
	"time"
)

type Config struct {
	Driver     string `json:"driver"`
	Datasource string `json:"datasource"`
	Debug      bool   `json:"debug"`
}

func dealDriver(driver string) string {
	driver = strings.TrimSpace(driver)
	switch driver {
	case "mysql":
		return "mysql"
	case "postgres":
		return "postgres"
	default:
		return "sqlite3"
	}
}

func New(cfg *Config) (*sql.DB, error) {
	if strings.TrimSpace(cfg.Datasource) == "" {
		cfg.Datasource = "control.db"
	}
	db, err := sql.Open(dealDriver(cfg.Driver), cfg.Datasource)
	if err != nil {
		return nil, err
	}
	db.SetMaxIdleConns(10)               // 最大空闲连接数
	db.SetMaxOpenConns(100)              // 最大打开连接数
	db.SetConnMaxLifetime(time.Hour * 6) // 连接最长存活时间
	return db, nil
}
