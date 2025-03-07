package cluster

import (
	"ddm-admin-console/app/registry"
	"ddm-admin-console/router"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type clusterKey struct {
	Scenario               string
	SignKeyIssuer          string   `form:"sign-key-issuer" binding:"required_if=KeyDeviceType file Scenario key-required"`
	SignKeyPwd             string   `form:"sign-key-pwd" binding:"required_if=KeyDeviceType file Scenario key-required"`
	KeyDeviceType          string   `form:"key-device-type"`
	RemoteType             string   `form:"remote-type" binding:"required_if=KeyDeviceType hardware Scenario key-required"`
	RemoteKeyPassword      string   `form:"remote-key-pwd" binding:"required_if=KeyDeviceType hardware Scenario key-required"`
	RemoteCAName           string   `form:"remote-ca-name" binding:"required_if=KeyDeviceType hardware Scenario key-required"`
	RemoteCAHost           string   `form:"remote-ca-host" binding:"required_if=KeyDeviceType hardware Scenario key-required"`
	RemoteCAPort           string   `form:"remote-ca-port" binding:"required_if=KeyDeviceType hardware Scenario key-required"`
	RemoteSerialNumber     string   `form:"remote-serial-number" binding:"required_if=KeyDeviceType hardware Scenario key-required"`
	RemoteKeyPort          string   `form:"remote-key-port" binding:"required_if=KeyDeviceType hardware Scenario key-required"`
	RemoteKeyHost          string   `form:"remote-key-host" binding:"required_if=KeyDeviceType hardware Scenario key-required"`
	RemoteKeyMask          string   `form:"remote-key-mask" binding:"required_if=KeyDeviceType hardware Scenario key-required"`
	INIConfig              string   `form:"remote-ini-config" binding:"required_if=KeyDeviceType hardware Scenario key-required"`
	AllowedKeysSerial      []string `form:"allowed-keys-serial[]" binding:"required_if=Scenario key-required"`
	AllowedKeysIssuer      []string `form:"allowed-keys-issuer[]" binding:"required_if=Scenario key-required"`
	KeyDataChanged         string   `form:"key-data-changed"`
	KeyVerificationChanged string   `form:"key-verification-changed"`
}

type keyManagement struct {
	r               *clusterKey
	vaultSecretPath string
}

func (k keyManagement) VaultSecretPath() string {
	return k.vaultSecretPath
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

func (k keyManagement) KeyDataChanged() bool {
	return k.r.KeyDataChanged == "on"
}

func (k keyManagement) KeyVerificationChanged() bool {
	return k.r.KeyVerificationChanged == "on"
}

func (a *App) vaultPlatformPathKey(key string) string {
	return fmt.Sprintf("%s/%s",
		strings.ReplaceAll(a.VaultClusterPathTemplate, "{engine}", a.Config.VaultKVEngineName), key)
}

func (a *App) updateKeyView(ctx *gin.Context) (router.Response, error) {
	if err := a.updateKey(ctx); err != nil {
		return nil, errors.Wrap(err, "unable to update keys")
	}

	return router.MakeRedirectResponse(http.StatusFound, "/admin/cluster/management"), nil
}

func (a *App) updateKey(ctx *gin.Context) error {
	ck := clusterKey{}

	if err := ctx.ShouldBind(&ck); err != nil {
		_, ok := err.(validator.ValidationErrors)
		if !ok {
			return errors.Wrap(err, "unable to parse registry form")
		}

		return err
	}

	values, err := registry.GetValuesFromGit(a.Config.CodebaseName, registry.MasterBranch, a.Gerrit)
	if err != nil {
		return errors.Wrap(err, "unable to get values from git")
	}
	vaultSecretData := make(map[string]map[string]interface{})

	vaultPath := a.vaultPlatformPathKey(fmt.Sprintf("%s-%s", registry.KeyManagementVaultPath, time.Now().Format("20060201T150405Z")))

	repoFiles := make(map[string]string)
	keysChanged, err := registry.PrepareRegistryKeys(
		keyManagement{r: &ck, vaultSecretPath: vaultPath},
		ctx.Request,
		vaultSecretData,
		values,
		repoFiles,
		a.Vault,
	)
	if err != nil {
		return errors.Wrap(err, "unable to create registry keys")
	}

	if err := registry.CacheRepoFiles(a.TempFolder, a.ClusterRepo, repoFiles, a.appCache); err != nil {
		return fmt.Errorf("unable to cache repo files")
	}

	if keysChanged || len(repoFiles) > 0 {
		if err := registry.CreateEditMergeRequest(ctx, a.Config.CodebaseName, values.OriginalYaml, a.Gerrit, []string{}); err != nil {
			return errors.Wrap(err, "unable to create edit merge request")
		}
	}

	if len(vaultSecretData) > 0 {
		if err := registry.CreateVaultSecrets(a.Vault, vaultSecretData, false); err != nil {
			return errors.Wrap(err, "unable to create vault secrets")
		}
	}

	return nil
}
