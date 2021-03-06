package cluster

import (
	"ddm-admin-console/app/registry"
	"ddm-admin-console/router"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
)

type clusterKey struct {
	Scenario           string
	SignKeyIssuer      string   `form:"sign-key-issuer" binding:"required_if=KeyDeviceType file Scenario key-required"`
	SignKeyPwd         string   `form:"sign-key-pwd" binding:"required_if=KeyDeviceType file Scenario key-required"`
	KeyDeviceType      string   `form:"key-device-type" binding:"oneof=file hardware"`
	RemoteType         string   `form:"remote-type" binding:"required_if=KeyDeviceType hardware Scenario key-required"`
	RemoteKeyPassword  string   `form:"remote-key-pwd" binding:"required_if=KeyDeviceType hardware Scenario key-required"`
	RemoteCAName       string   `form:"remote-ca-name" binding:"required_if=KeyDeviceType hardware Scenario key-required"`
	RemoteCAHost       string   `form:"remote-ca-host" binding:"required_if=KeyDeviceType hardware Scenario key-required"`
	RemoteCAPort       string   `form:"remote-ca-port" binding:"required_if=KeyDeviceType hardware Scenario key-required"`
	RemoteSerialNumber string   `form:"remote-serial-number" binding:"required_if=KeyDeviceType hardware Scenario key-required"`
	RemoteKeyPort      string   `form:"remote-key-port" binding:"required_if=KeyDeviceType hardware Scenario key-required"`
	RemoteKeyHost      string   `form:"remote-key-host" binding:"required_if=KeyDeviceType hardware Scenario key-required"`
	RemoteKeyMask      string   `form:"remote-key-mask" binding:"required_if=KeyDeviceType hardware Scenario key-required"`
	INIConfig          string   `form:"remote-ini-config" binding:"required_if=KeyDeviceType hardware Scenario key-required"`
	AllowedKeysSerial  []string `form:"allowed-keys-serial[]" binding:"required_if=Scenario key-required"`
	AllowedKeysIssuer  []string `form:"allowed-keys-issuer[]" binding:"required_if=Scenario key-required"`
}

type keyManagement struct {
	r *clusterKey
}

func (k keyManagement) KeyDeviceType() string {
	return k.r.KeyDeviceType
}

func (k keyManagement) AllowedKeysIssuer() []string {
	return k.r.AllowedKeysIssuer
}

func (k keyManagement) AllowedKeysSerial() []string {
	return k.r.AllowedKeysSerial
}

func (k keyManagement) SignKeyIssuer() string {
	return k.r.SignKeyIssuer
}

func (k keyManagement) SignKeyPwd() string {
	return k.r.SignKeyPwd
}

func (k keyManagement) RemoteType() string {
	return k.r.RemoteType
}

func (k keyManagement) RemoteSerialNumber() string {
	return k.r.RemoteSerialNumber
}

func (k keyManagement) RemoteKeyPort() string {
	return k.r.RemoteKeyPort
}

func (k keyManagement) RemoteKeyHost() string {
	return k.r.RemoteKeyHost
}

func (k keyManagement) RemoteKeyPassword() string {
	return k.r.RemoteKeyPassword
}

func (k keyManagement) INIConfig() string {
	return k.r.INIConfig
}

func (k keyManagement) KeysRequired() bool {
	return true
}

func (k keyManagement) FilesSecretName() string {
	return "digital-signature-ops-data"
}

func (k keyManagement) EnvVarsSecretName() string {
	return "digital-signature-ops-env-vars"
}

func (a *App) updateKeyView(ctx *gin.Context) (*router.Response, error) {
	if err := a.updateKey(ctx); err != nil {
		return nil, errors.Wrap(err, "unable to update keys")
	}

	return router.MakeRedirectResponse(http.StatusFound, "/admin/cluster/management"), nil
}

func (a *App) updateKey(ctx *gin.Context) error {
	ck := clusterKey{Scenario: registry.ScenarioKeyRequired}

	if err := ctx.ShouldBind(&ck); err != nil {
		_, ok := err.(validator.ValidationErrors)
		if !ok {
			return errors.Wrap(err, "unable to parse registry form")
		}

		return err
	}

	userCtx := a.router.ContextWithUserAccessToken(ctx)
	k8sService, err := a.Services.K8S.ServiceForContext(userCtx)
	if err != nil {
		return errors.Wrap(err, "unable to init service for user context")
	}

	if err := registry.CreateRegistryKeys(keyManagement{r: &ck}, ctx.Request, k8sService); err != nil {
		return errors.Wrap(err, "unable to create registry keys")
	}

	return nil
}
