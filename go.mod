module github.com/zc2638/drone-control

go 1.15

require (
	github.com/drone/drone v1.9.0
	github.com/drone/drone-go v1.3.2 // indirect
	github.com/drone/drone-yaml v1.2.4-0.20200326192514-6f4d6dfb39e4
	github.com/go-chi/chi v4.1.2+incompatible
	github.com/google/uuid v1.1.2 // indirect
	github.com/hashicorp/go-multierror v1.0.0
	github.com/jmoiron/sqlx v0.0.0-20180614180643-0dae4fefe7c0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/mapstructure v1.1.2
	github.com/pkgms/go v0.0.0-20200907134721-398896e623cf
	github.com/sirupsen/logrus v1.6.0
	github.com/spf13/cobra v1.0.0
	github.com/spf13/viper v1.4.0
	github.com/vinzenz/yaml v0.0.0-20170920082545-91409cdd725d
	gorm.io/driver/sqlite v1.1.1
	gorm.io/gorm v1.20.0
)

replace github.com/h2non/gock => gopkg.in/h2non/gock.v1 v1.0.15
