/**
 * Created by zc on 2020/9/7.
 */
package api

import (
	"encoding/json"
	"errors"
	"github.com/go-chi/chi"
	"github.com/pkgms/go/ctr"
	"github.com/zc2638/drone-control/global"
	"github.com/zc2638/drone-control/store"
	"gorm.io/gorm"
	"net/http"
	"strconv"
)

type RepoParams struct {
	UserID    int64  `json:"user_id"`
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Timeout   int64  `json:"timeout"`
	Data      string `json:"data"`
}

// RepoList returns an http.HandlerFunc that processes an
// http.Request to get the repo list.
func RepoList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pageStr := r.URL.Query().Get("page")
		sizeStr := r.URL.Query().Get("size")
		size, _ := strconv.Atoi(sizeStr)
		page, _ := strconv.Atoi(pageStr)
		if page < 1 {
			page = 1
		}
		if size < 1 {
			size = 10
		}
		var repos []store.ReposData
		db := global.GormDB().Limit(size).Offset((page - 1) * size).Find(&repos)
		if db.Error != nil {
			ctr.BadRequest(w, db.Error)
			return
		}
		ctr.OK(w, repos)
	}
}

// RepoInfo returns an http.HandlerFunc that processes an
// http.Request to get the repo details.
func RepoInfo() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.ParseInt(
			chi.URLParam(r, "repo"), 10, 64)
		if id < 1 {
			ctr.BadRequest(w, errors.New("cannot find repo, repo_id is invalid"))
			return
		}

		var repo store.ReposData
		db := global.GormDB().Where(&store.ReposData{
			ID: id,
		}).First(&repo)
		if db.Error != nil {
			ctr.BadRequest(w, db.Error)
			return
		}
		ctr.OK(w, repo)
	}
}

// RepoInfoBySlug returns an http.HandlerFunc that processes an
// http.Request to get the repo details by namespace and name.
func RepoInfoBySlug() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		namespace := chi.URLParam(r, "namespace")
		name := chi.URLParam(r, "name")
		if namespace == "" || name == "" {
			ctr.BadRequest(w, errors.New("space or name is empty"))
			return
		}
		var repo store.ReposData
		db := global.GormDB().Where(&store.ReposData{
			Namespace: namespace,
			Name:      name,
		}).First(&repo)
		if db.Error != nil {
			ctr.BadRequest(w, db.Error)
			return
		}
		if db.RowsAffected == 0 {
			ctr.OK(w, nil)
			return
		}
		ctr.OK(w, repo)
	}
}

// RepoApply returns an http.HandlerFunc that processes an
// http.Request to create a repo.
func RepoApply() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var data RepoParams
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			ctr.BadRequest(w, err)
			return
		}
		defer r.Body.Close()
		fullData, err := CheckClone(data.Data)
		if err != nil {
			ctr.BadRequest(w, errors.New("exec yaml failed"))
			return
		}
		if _, err := parsePipeline(fullData, nil); err != nil {
			ctr.BadRequest(w, err)
			return
		}
		var timeout int64 = global.DefaultRepoTimeout // Minute
		if data.Timeout > 0 {
			timeout = data.Timeout
		}
		repo := &store.ReposData{
			UserID:    data.UserID,
			Namespace: data.Namespace,
			Name:      data.Name,
			Timeout:   timeout,
			Data:      data.Data,
		}

		var current store.ReposData
		db := global.GormDB().Where(&store.ReposData{
			Namespace: data.Namespace,
			Name:      data.Name,
		}).First(&current)
		if db.Error != nil && !errors.Is(gorm.ErrRecordNotFound, db.Error) {
			ctr.BadRequest(w, db.Error)
			return
		}
		if db.RowsAffected == 0 {
			err = global.GormDB().Create(repo).Error
		} else {
			err = global.GormDB().Model(&store.ReposData{}).Where(&store.ReposData{
				Namespace: data.Namespace,
				Name:      data.Name,
			}).Updates(repo).Error
		}
		if err != nil {
			ctr.BadRequest(w, err)
			return
		}
		ctr.OK(w, repo)
	}
}

// RepoDelete returns an http.HandlerFunc that processes an
// http.Request to delete the repo.
func RepoDelete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		namespace := chi.URLParam(r, "namespace")
		name := chi.URLParam(r, "name")
		if namespace == "" || name == "" {
			ctr.BadRequest(w, errors.New("namespace or name is empty"))
			return
		}
		if err := global.GormDB().Where(&store.ReposData{
			Namespace: namespace,
			Name:      name,
		}).Delete(&store.ReposData{}).Error; err != nil {
			ctr.BadRequest(w, err)
			return
		}
		ctr.Success(w)
	}
}
