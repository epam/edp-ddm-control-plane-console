package query

import "time"

type ActionLog struct {
	ID             int       `json:"id" orm:"column(id)"`
	LastTimeUpdate time.Time `json:"last_time_update" orm:"column(updated_at)"`
	UserName       string    `json:"user_name" orm:"column(username)"`
	Message        string    `json:"message" orm:"column(action_message)"`
	Action         string    `json:"action" orm:"column(action)"`
	Result         string    `json:"result" orm:"column(result)"`
}

func (a *ActionLog) TableName() string {
	return "action_log"
}

func (a ActionLog) FormattedLastTimeUpdate() string {
	return a.LastTimeUpdate.Format(viewTimeFormat)
}
