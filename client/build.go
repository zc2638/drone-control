/**
 * Created by zc on 2020/11/27.
 */
package client

import (
	"errors"
	"github.com/drone/drone/core"
	"github.com/zc2638/drone-control/handler"
	"strconv"
)

type buildClient struct {
	client    *Client
	namespace string
	name      string
}

func (c *buildClient) List() ([]core.Build, error) {
	res, err := c.client.cli().
		SetPathParams(map[string]string{
			"namespace": c.namespace,
			"name":      c.name,
		}).
		SetResult([]core.Build{}).
		Get(handler.PathBuildSlug)
	if err != nil {
		return nil, err
	}
	return *res.Result().(*[]core.Build), nil
}

func (c *buildClient) Info(build int) (*core.Build, error) {
	res, err := c.client.cli().
		SetPathParams(map[string]string{
			"namespace": c.namespace,
			"name":      c.name,
			"build":     strconv.Itoa(build),
		}).
		SetResult(core.Build{}).
		Get(handler.PathBuildSlugNum)
	if err != nil {
		return nil, err
	}
	result := res.Result().(*core.Build)
	if result.ID == 0 {
		return nil, nil
	}
	return result, nil
}

func (c *buildClient) Run() error {
	res, err := c.client.cli().
		SetPathParams(map[string]string{
			"namespace": c.namespace,
			"name":      c.name,
		}).
		Post(handler.PathBuildSlug)
	if err != nil {
		return err
	}
	if res.StatusCode() == 200 {
		return nil
	}
	return errors.New(res.String())
}

func (c *buildClient) Log(build, stage, step int) ([]byte, error) {
	res, err := c.client.cli().
		SetPathParams(map[string]string{
			"namespace": c.namespace,
			"name":      c.name,
			"build":     strconv.Itoa(build),
			"stage":     strconv.Itoa(stage),
			"step":      strconv.Itoa(step),
		}).
		Get(handler.PathBuildLogSlug)
	if err != nil {
		return nil, err
	}
	return res.Body(), nil
}
