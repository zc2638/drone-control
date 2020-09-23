/**
 * Created by zc on 2020/9/5.
 */
package global

import (
	"database/sql"
	dronedb "github.com/drone/drone/store/shared/db"
	"github.com/drone/drone/store/shared/migrate/mysql"
	"github.com/drone/drone/store/shared/migrate/postgres"
	"github.com/drone/drone/store/shared/migrate/sqlite"
	"github.com/jmoiron/sqlx"
	gormsqlite "gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"sync"
	"time"
	"unsafe"
)

var sqlDB *sql.DB

func DB() *sql.DB {
	return sqlDB
}

func NewDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "control.db")
	if err != nil {
		return nil, err
	}
	db.SetMaxIdleConns(10)               // 最大空闲连接数
	db.SetMaxOpenConns(100)              // 最大打开连接数
	db.SetConnMaxLifetime(time.Hour * 6) // 连接最长存活时间
	return db, nil
}

var gormDB *gorm.DB

func GormDB() *gorm.DB {
	return gormDB
}

func NewGormDB(db *sql.DB) (*gorm.DB, error) {
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
	dialector := gormsqlite.Dialector{}
	if err := dialector.Initialize(dbConnect); err != nil {
		return nil, err
	}
	dbConnect.Dialector = dialector
	return dbConnect.Debug(), nil
}

var droneDB *dronedb.DB

func DroneDB() *dronedb.DB {
	return droneDB
}

func NewDroneDB(driver string, db *sql.DB) (*dronedb.DB, error) {
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
