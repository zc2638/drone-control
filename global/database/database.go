/**
 * Created by zc on 2020/11/23.
 */
package database

import (
	"database/sql"
	dronedb "github.com/drone/drone/store/shared/db"
	"github.com/drone/drone/store/shared/migrate/mysql"
	"github.com/drone/drone/store/shared/migrate/postgres"
	"github.com/drone/drone/store/shared/migrate/sqlite"
	"github.com/jmoiron/sqlx"
	"sync"
	"time"
	"unsafe"
)

func NewDroneDB(cfg *Config, db *sql.DB) (*dronedb.DB, error) {
	driver := dealDriver(cfg.Driver)
	switch driver {
	case "mysql":
		db.SetMaxIdleConns(0)
	}
	if err := pingDatabase(db); err != nil {
		return nil, err
	}
	if err := setupDatabase(db, driver); err != nil {
		return nil, err
	}

	var engine dronedb.Driver
	var locker dronedb.Locker
	switch driver {
	case "mysql":
		engine = dronedb.Mysql
		locker = &nopLocker{}
	case "postgres":
		engine = dronedb.Postgres
		locker = &nopLocker{}
	default:
		engine = dronedb.Sqlite
		locker = &sync.RWMutex{}
	}
	return (*dronedb.DB)(unsafe.Pointer(&DBCopy{
		conn:   sqlx.NewDb(db, driver),
		lock:   locker,
		driver: engine,
	})), nil
}

// helper function to ping the database with backoff to ensure
// a connection can be established before we proceed with the
// database setup and migration.
func pingDatabase(db *sql.DB) (err error) {
	for i := 0; i < 30; i++ {
		err = db.Ping()
		if err == nil {
			return
		}
		time.Sleep(time.Second)
	}
	return
}

// helper function to setup the databsae by performing automated
// database migration steps.
func setupDatabase(db *sql.DB, driver string) error {
	switch driver {
	case "mysql":
		return mysql.Migrate(db)
	case "postgres":
		return postgres.Migrate(db)
	default:
		return sqlite.Migrate(db)
	}
}

// DB is a pool of zero or more underlying connections to
// the drone database.
type DBCopy struct {
	conn   *sqlx.DB
	lock   dronedb.Locker
	driver dronedb.Driver
}

type nopLocker struct{}

func (nopLocker) Lock()    {}
func (nopLocker) Unlock()  {}
func (nopLocker) RLock()   {}
func (nopLocker) RUnlock() {}
