package models

import (
	"time"

	v1 "k8s.io/api/core/v1"
)

const viewTimeFormat = "02.01.2006 15:04"

type Registry struct {
	Name        string `form:"name" valid:"Required;Match(/^[a-z0-9]([-a-z0-9]*[a-z0-9])?([a-z0-9]([-a-z0-9]*[a-z0-9])?)*$/);MinSize(3);MaxSize(12)"`
	Description string `form:"description" valid:"MaxSize(250)"`
	Admins      string `form:"admins" valid:"Match(/(^[a-zA-Z0-9@._,-]+$)|(^$)/)"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	NS          *v1.Namespace
}

func (r Registry) FormattedCreatedAt() string {
	return r.CreatedAt.Format(viewTimeFormat)
}
