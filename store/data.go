/**
 * Created by zc on 2020/9/5.
 */
package store

const (
	StatusSkipped  = "skipped"
	StatusBlocked  = "blocked"
	StatusDeclined = "declined"
	StatusWaiting  = "waiting_on_dependencies"
	StatusPending  = "pending"
	StatusRunning  = "running"
	StatusPassing  = "success"
	StatusFailing  = "failure"
	StatusKilled   = "killed"
	StatusError    = "error"
)

type Repos struct {
	ID          int64  `json:"id" gorm:"primaryKey;uniqueIndex;autoIncrement"`
	UID         string `json:"uid"`
	UserID      int64  `json:"user_id"`
	Namespace   string `json:"namespace"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	SCM         string `json:"scm"`
	HTTPURL     string `json:"git_http_url"`
	SSHURL      string `json:"git_ssh_url"`
	Link        string `json:"link"`
	Branch      string `json:"default_branch"`
	Private     bool   `json:"private"`
	Visibility  string `json:"visibility"`
	Active      bool   `json:"active"`
	Config      string `json:"config_path"`
	Trusted     bool   `json:"trusted"`
	Protected   bool   `json:"protected"`
	IgnoreForks bool   `json:"ignore_forks"`
	IgnorePulls bool   `json:"ignore_pull_requests"`
	CancelPulls bool   `json:"auto_cancel_pull_requests"`
	CancelPush  bool   `json:"auto_cancel_pushes"`
	Timeout     int64  `json:"timeout"`
	Counter     int64  `json:"counter"`
	Synced      int64  `json:"synced"`
	Created     int64  `json:"created"`
	Updated     int64  `json:"updated"`
	Version     int64  `json:"version"`
	Signer      string `json:"signer"`
	Secret      string `json:"secret"`
}

type ReposData struct {
	ID        int64  `json:"id" gorm:"primaryKey;uniqueIndex;autoIncrement"`
	Username  string `json:"username"`
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Timeout   int64  `json:"timeout"`
	Data      string `json:"data" gorm:"type:text"`
	Created   int64  `json:"created" gorm:"autoCreateTime:milli"`
	Updated   int64  `json:"updated" gorm:"autoUpdateTime:milli"`
}
