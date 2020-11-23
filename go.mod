module github.com/zc2638/drone-control

go 1.15

require (
	github.com/PuerkitoBio/goquery v1.5.1 // indirect
	github.com/drone/drone v1.9.0
	github.com/drone/drone-go v1.3.2 // indirect
	github.com/drone/drone-yaml v1.2.4-0.20200326192514-6f4d6dfb39e4
	github.com/go-chi/chi v4.1.2+incompatible
	github.com/go-resty/resty/v2 v2.3.0
	github.com/go-sql-driver/mysql v1.5.0
	github.com/google/uuid v1.1.2 // indirect
	github.com/hashicorp/go-multierror v1.0.0
	github.com/jackc/pgx/v4 v4.9.2 // indirect
	github.com/jmoiron/sqlx v0.0.0-20180614180643-0dae4fefe7c0
	github.com/lib/pq v1.3.0
	github.com/mattn/go-sqlite3 v1.14.5
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/mapstructure v1.1.2
	github.com/pkgms/go v0.0.0-20200907134721-398896e623cf
	github.com/sirupsen/logrus v1.6.0
	github.com/spf13/cobra v1.0.0
	github.com/spf13/viper v1.4.0
	github.com/vinzenz/yaml v0.0.0-20170920082545-91409cdd725d
	golang.org/x/crypto v0.0.0-20201117144127-c1f2f97bffc9 // indirect
	golang.org/x/text v0.3.4 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	gorm.io/driver/mysql v1.0.3
	gorm.io/driver/postgres v1.0.5
	gorm.io/driver/sqlite v1.1.3
	gorm.io/gorm v1.20.7
)

replace github.com/h2non/gock => gopkg.in/h2non/gock.v1 v1.0.15
