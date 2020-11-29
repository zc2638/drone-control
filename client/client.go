/**
 * Created by zc on 2020/9/29.
 */
package client

import (
	"github.com/drone/drone/core"
	"github.com/go-resty/resty/v2"
	"github.com/zc2638/drone-control/handler/api"
	"github.com/zc2638/drone-control/store"
	"io"
)

type Interface interface {
	Repo(namespace, name string) RepoInterface
	Build(namespace, name string) BuildInterface
	Stream(namespace, name string) StreamInterface
}

type (
	RepoInterface interface {
		List() ([]store.ReposData, error)
		Info() (*store.ReposData, error)
		Apply(repo *api.RepoParams) error
		Delete() error
	}

	BuildInterface interface {
		List() ([]core.Build, error)
		Info(build int) (*core.Build, error)
		Run() error
		Log(build, stage, step int) ([]byte, error)
	}

	StreamInterface interface {
		Log(build, stage, step int) (io.ReadCloser, error)
	}
)

type Config struct {
	Host string `json:"host"`
}

type Client struct {
	config Config
}

func New(cfg Config) Interface {
	return &Client{config: cfg}
}

func (c *Client) cli() *resty.Request {
	return resty.New().
		SetHostURL(c.config.Host).
		SetHeader("Content-Type", "application/json").
		R()
}

func (c *Client) Repo(namespace, name string) RepoInterface {
	return &repoClient{
		client:    c,
		namespace: namespace,
		name:      name,
	}
}

func (c *Client) Build(namespace, name string) BuildInterface {
	return &buildClient{
		client:    c,
		namespace: namespace,
		name:      name,
	}
}

func (c *Client) Stream(namespace, name string) StreamInterface {
	return &streamClient{
		client:    c,
		namespace: namespace,
		name:      name,
	}
}
