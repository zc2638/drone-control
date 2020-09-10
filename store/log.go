/**
 * Created by zc on 2020/9/5.
 */
package store

import (
	"context"
	"github.com/drone/drone/core"
	"io"
)

type logStore struct{ store }

func LogStore() core.LogStore {
	s := &logStore{}
	s.Init()
	return s
}

func (s *logStore) Find(ctx context.Context, stage int64) (io.ReadCloser, error) {
	panic("implement me")
}

func (s *logStore) Create(ctx context.Context, stage int64, r io.Reader) error {
	panic("implement me")
}

func (s *logStore) Update(ctx context.Context, stage int64, r io.Reader) error {
	panic("implement me")
}

func (s *logStore) Delete(ctx context.Context, stage int64) error {
	panic("implement me")
}
