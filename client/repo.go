/**
 * Created by zc on 2020/11/27.
 */
package client

import (
	"errors"
	"github.com/zc2638/drone-control/handler"
	"github.com/zc2638/drone-control/handler/api"
	"github.com/zc2638/drone-control/store"
)

type repoClient struct{ *slug }

func (c *repoClient) List() ([]store.ReposData, error) {
	res, err := c.client.R().
		SetResult([]store.ReposData{}).
		Get(handler.PathRepo)
	if err != nil {
		return nil, err
	}
	return *res.Result().(*[]store.ReposData), nil
}

func (c *repoClient) Info() (*store.ReposData, error) {
	res, err := c.client.R().
		SetPathParams(map[string]string{
			"namespace": c.namespace,
			"name":      c.name,
		}).
		SetResult(store.ReposData{}).
		Get(handler.PathRepoSlug)
	if err != nil {
		return nil, err
	}
	result := res.Result().(*store.ReposData)
	if result.ID == 0 {
		return nil, nil
	}
	return result, nil
}

func (c *repoClient) Apply(repo *api.RepoParams) error {
	res, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(repo).
		Post(handler.PathRepo)
	if err != nil {
		return err
	}
	if res.StatusCode() == 200 {
		return nil
	}
	return errors.New(res.String())
}

func (c *repoClient) Delete() error {
	res, err := c.client.R().
		SetPathParams(map[string]string{
			"namespace": c.namespace,
			"name":      c.name,
		}).
		Delete(handler.PathRepoSlug)
	if err != nil {
		return err
	}
	if res.StatusCode() == 200 {
		return nil
	}
	return errors.New(res.String())
}
