/**
 * Created by zc on 2020/9/5.
 */
package global

import (
	"fmt"
	droneconfig "github.com/drone/drone/cmd/drone-server/config"
	"github.com/mitchellh/go-homedir"
	"github.com/mitchellh/mapstructure"
	"github.com/pkgms/go/server"
	"github.com/spf13/viper"
	"github.com/zc2638/drone-control/global/database"
	"net"
	"os"
	"strconv"
	"strings"
)

const EnvPrefix = "DRONE_CONTROL"
const PathStageLog = "slogs"
const ServerPort = 2639
const DefaultRepoTimeout = 30

type Config struct {
	Server   server.Config       `json:"server"`
	Cleanup  droneconfig.Cleanup `json:"cleanup"`
	RPC      droneconfig.RPC     `json:"rpc"`
	Database database.Config     `json:"database"`
}

func Environ() *Config {
	cfg := &Config{}
	cfg.Server.Port = ServerPort
	cfg.RPC.Proto = "http"
	cfg.RPC.Host = net.JoinHostPort("127.0.0.1", strconv.Itoa(cfg.Server.Port))
	return cfg
}

func ParseConfig(cfgPath string) (*Config, error) {
	if cfgPath != "" {
		viper.SetConfigFile(cfgPath)
	} else {
		home, err := homedir.Dir()
		if err != nil {
			return nil, err
		}
		viper.AddConfigPath(home)
		viper.SetConfigName("config.yaml")
	}
	viper.SetEnvPrefix(EnvPrefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	cfg := Environ()
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(*os.PathError); ok {
			fmt.Println("Warning: not find config file, use default.")
			return cfg, nil
		}
		return nil, err
	}
	fmt.Println("Using config file:", viper.ConfigFileUsed())
	err := viper.Unmarshal(cfg, func(dc *mapstructure.DecoderConfig) {
		dc.TagName = "json"
	})
	return cfg, err
}
