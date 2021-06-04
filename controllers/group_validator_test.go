package controllers

import (
	"github.com/astaxie/beego/context"
	"github.com/stretchr/testify/mock"
)

type mockGroupValidator struct {
	mock.Mock
}

func (m *mockGroupValidator) IsAllowedToCreateRegistry(ctx *context.Context, creatorGroup string) bool {
	return m.Called(creatorGroup).Bool(0)
}
