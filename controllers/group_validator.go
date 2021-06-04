package controllers

import "github.com/astaxie/beego/context"

type GroupValidator interface {
	IsAllowedToCreateRegistry(ctx *context.Context, creatorGroup string) bool
}

type groupValidator struct{}

func (g *groupValidator) IsAllowedToCreateRegistry(ctx *context.Context, creatorGroup string) bool {
	data := ctx.Input.Session(sessionGroupsKey)
	if data == nil {
		return false
	}

	groups, ok := data.([]string)
	if !ok {
		return false
	}

	for _, gr := range groups {
		if gr == creatorGroup {
			return true
		}
	}

	return false
}
