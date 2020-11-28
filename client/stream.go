/**
 * Created by zc on 2020/11/28.
 */
package client

import (
	"github.com/zc2638/drone-control/handler"
	"io"
	"strconv"
)

type streamClient struct {
	client    *Client
	namespace string
	name      string
}

func (c *streamClient) Log(build, stage, step int) (io.ReadCloser, error) {
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
	return res.RawBody(), nil
}
