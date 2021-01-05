package filters

import (
	"github.com/astaxie/beego"
	"regexp"
)

var roles map[string][]string

func init() {
	administrator := beego.AppConfig.String("adminRole")
	developer := beego.AppConfig.String("developerRole")
	roles = map[string][]string{
		"GET /admin/overview$":                    {administrator, developer},
		"GET /admin/application/overview":         {administrator, developer},
		"GET /admin/application/create$":          {administrator},
		"GET /admin/cd-pipeline/overview":         {administrator, developer},
		"GET /admin/cd-pipeline/create$":          {administrator, developer},
		"GET /admin/cd-pipeline/([^/]*)/overview": {administrator, developer},
		"POST /admin/cd-pipeline$":                {administrator},
		"POST /admin/application$":                {administrator},
		"POST /admin/codebase/([^/]*)/branch$":    {administrator},
		"GET /admin/autotest/overview":            {administrator, developer},
		"GET /admin/autotest/create$":             {administrator, developer},
		"GET /admin/codebase/([^/]*)/overview":    {administrator, developer},
		"POST /admin/autotest$":                   {administrator},
		"GET /admin/library/overview":             {administrator, developer},
		"GET /admin/library/create$":              {administrator, developer},
		"POST /admin/library$":                    {administrator},
		"GET /admin/service/overview":             {administrator, developer},
		"GET /admin/cd-pipeline/([^/]*)/update":   {administrator, developer},
		"POST /admin/cd-pipeline/([^/]*)/update":  {administrator},
		"POST /admin/codebase$":                   {administrator},
		"POST /admin/stage$":                      {administrator},
		"POST /admin/cd-pipeline/delete":          {administrator},
		"GET /admin/diagram/overview":             {administrator, developer},

		"GET /api/v1/edp/vcs$":                               {administrator, developer},
		"GET /api/v1/edp/codebase":                           {administrator, developer},
		"GET /api/v1/edp/codebase/([^/]*)$":                  {administrator, developer},
		"GET /api/v1/edp/cd-pipeline/([^/]*)$":               {administrator, developer},
		"GET /api/v1/edp/cd-pipeline/([^/]*)/stage/([^/]*)$": {administrator, developer},
		"POST /api/v1/edp/codebase$":                         {administrator},
		"POST /api/v1/edp/cd-pipeline$":                      {administrator},
		"PUT /api/v1/edp/cd-pipeline/([^/]*)$":               {administrator},
		"DELETE /api/v1/edp/codebase$":                       {administrator},
		"DELETE /api/v1/edp/stage$":                          {administrator},
	}
}

func IsPageAvailable(key string, contextRoles []string) bool {
	pageRoles := getValue(key)
	if pageRoles == nil {
		return true
	}

	if getIntersectionOfRoles(contextRoles, pageRoles) == nil {
		return false
	}
	return true
}

func getValue(key string) []string {
	for k, v := range roles {
		match, _ := regexp.MatchString(k, key)
		if match {
			return v
		}
	}
	return nil
}

func getIntersectionOfRoles(a, b []string) (c []string) {
	m := make(map[string]bool)

	for _, item := range a {
		m[item] = true
	}

	for _, item := range b {
		if _, ok := m[item]; ok {
			c = append(c, item)
		}
	}
	return
}
