package registry

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"ddm-admin-console/router"
	"ddm-admin-console/service/gerrit"
)

const (
	valuesLocation             = "deploy-templates/values.yaml"
	mrLabelTarget              = "console/target"
	mrAnnotationRegName        = "ext-reg/name"
	mrAnnotationRegType        = "ext-reg/type"
	externalSystemTypeExternal = "external-system"
	erValuesIndex              = "nontrembita-external-registration"
	erStatusInactive           = "inactive"
	erStatusActive             = "active"
	erStatusDisabled           = "disabled"
)

type ExternalRegistration struct {
	Name     string `yaml:"name"`
	Enabled  bool   `yaml:"enabled"`
	External bool   `yaml:"external"`
	status   string
	KeyValue string
}

func (e ExternalRegistration) Status() string {
	s := e.status
	if s == "" {
		s = erStatusActive
	}

	if !e.Enabled {
		s = erStatusDisabled
	}

	return fmt.Sprintf("status-%s", s)
}

func (a *App) addExternalReg(ctx *gin.Context) (*router.Response, error) {
	userCtx := a.router.ContextWithUserAccessToken(ctx)

	registryName := ctx.Param("name")
	er := ExternalRegistration{
		Name:     ctx.PostForm("reg-name"),
		External: ctx.PostForm("external-system-type") == externalSystemTypeExternal,
		Enabled:  true,
	}
	values, err := a.prepareRegistryValues(userCtx, registryName, &er)
	if err != nil {
		return nil, errors.Wrap(err, "unable to prepare registry values")
	}

	if err := a.createErMergeRequest(userCtx, ctx, registryName, er.Name, values); err != nil {
		return nil, errors.Wrap(err, "unable to create MR")
	}

	return router.MakeRedirectResponse(http.StatusFound,
		fmt.Sprintf("/admin/registry/view/%s", registryName)), nil
}

func (a *App) getValuesFromGit(ctx context.Context, registryName string) (map[string]interface{}, []ExternalRegistration, error) {
	values, err := a.gerritService.GetFileContents(ctx, registryName, "master", valuesLocation)
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to get values yaml")
	}

	var valuesDict map[string]interface{}
	if err := yaml.Unmarshal([]byte(values), &valuesDict); err != nil {
		return nil, nil, errors.Wrap(err, "unable to decode values yaml")
	}
	if valuesDict == nil {
		valuesDict = make(map[string]interface{})
	}

	eRegs := make([]ExternalRegistration, 0)
	externalReg, ok := valuesDict[erValuesIndex]
	if ok {
		eRegs, err = convertExternalRegFromInterface(externalReg)
		if err != nil {
			return nil, nil, errors.Wrap(err, "unable to convert external regs")
		}
	}

	return valuesDict, eRegs, nil
}

func (a *App) prepareRegistryValues(ctx context.Context, registryName string, er *ExternalRegistration) (string, error) {
	valuesDict, eRegs, err := a.getValuesFromGit(ctx, registryName)
	if err != nil {
		return "", errors.Wrap(err, "unable to get values from git")
	}

	for _, _er := range eRegs {
		if er.Name == _er.Name {
			return "", errors.New("external reg system already exists")
		}
	}

	eRegs = append(eRegs, ExternalRegistration{
		Name:     er.Name,
		Enabled:  er.Enabled,
		External: er.External,
	})
	valuesDict[erValuesIndex] = eRegs

	newValues, err := yaml.Marshal(valuesDict)
	if err != nil {
		return "", errors.Wrap(err, "unable to encode new values yaml")
	}

	return string(newValues), nil
}

func (a *App) createErMergeRequest(userCtx context.Context, ctx *gin.Context, registryName, erName, values string) error {
	if err := a.gerritService.CreateMergeRequestWithContents(userCtx, &gerrit.MergeRequest{
		ProjectName:   registryName,
		Name:          fmt.Sprintf("ers-mr-%s-%s-%d", registryName, erName, time.Now().Unix()),
		AuthorEmail:   ctx.GetString(router.UserEmailSessionKey),
		AuthorName:    ctx.GetString(router.UserNameSessionKey),
		CommitMessage: fmt.Sprintf("update registry external reg systems"),
		TargetBranch:  "master",
		Labels: map[string]string{
			mrLabelTarget: "external-reg",
		},
		Annotations: map[string]string{
			mrAnnotationRegName: erName,
			mrAnnotationRegType: ctx.PostForm("external-system-type"),
		},
	}, map[string]string{
		valuesLocation: values,
	}); err != nil {
		return errors.Wrap(err, "unable to create MR with new values")
	}

	return nil
}

func (a *App) disableExternalReg(ctx *gin.Context) (*router.Response, error) {
	userCtx := a.router.ContextWithUserAccessToken(ctx)

	registryName := ctx.Param("name")
	systemName := ctx.PostForm("reg-name")
	if systemName == "" {
		return nil, errors.New("reg-name is required")
	}

	values, eRegs, err := a.getValuesFromGit(ctx, registryName)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get values from git")
	}

	found := false
	for i, v := range eRegs {
		if v.Name == systemName {
			eRegs[i].Enabled = !eRegs[i].Enabled
			found = true
		}
	}
	if !found {
		return nil, errors.New("reg-name not found")
	}

	values[erValuesIndex] = eRegs
	newValues, err := yaml.Marshal(values)
	if err != nil {
		return nil, errors.Wrap(err, "unable to encode new values yaml")
	}

	if err := a.createErMergeRequest(userCtx, ctx, registryName, systemName, string(newValues)); err != nil {
		return nil, errors.Wrap(err, "unable to crete MR")
	}

	return router.MakeRedirectResponse(http.StatusFound,
		fmt.Sprintf("/admin/registry/view/%s", registryName)), nil
}

func (a *App) removeExternalReg(ctx *gin.Context) (*router.Response, error) {
	userCtx := a.router.ContextWithUserAccessToken(ctx)

	registryName := ctx.Param("name")
	systemName := ctx.PostForm("reg-name")
	if systemName == "" {
		return nil, errors.New("reg-name is required")
	}

	values, eRegs, err := a.getValuesFromGit(ctx, registryName)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get values from git")
	}

	found := false
	for i, v := range eRegs {
		if v.Name == systemName {
			eRegs = append(eRegs[:i], eRegs[i+1:]...)
			found = true
		}
	}
	if !found {
		return nil, errors.New("reg-name not found")
	}

	values[erValuesIndex] = eRegs
	newValues, err := yaml.Marshal(values)
	if err != nil {
		return nil, errors.Wrap(err, "unable to encode new values yaml")
	}

	if err := a.createErMergeRequest(userCtx, ctx, registryName, systemName, string(newValues)); err != nil {
		return nil, errors.Wrap(err, "unable to crete MR")
	}

	return router.MakeRedirectResponse(http.StatusFound,
		fmt.Sprintf("/admin/registry/view/%s", registryName)), nil
}
