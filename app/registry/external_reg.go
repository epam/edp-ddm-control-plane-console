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
	valuesLocation      = "deploy-templates/values.yaml"
	mrLabelTarget       = "console/target"
	mrAnnotationRegName = "ext-reg/name"
	mrAnnotationRegType = "ext-reg/type"
)

func (a *App) addExternalReg(ctx *gin.Context) (*router.Response, error) {
	userCtx := a.router.ContextWithUserAccessToken(ctx)

	registryName := ctx.Param("name")
	er := ExternalRegistration{
		Name:     ctx.PostForm("reg-name"),
		External: ctx.PostForm("external-system-type") == "external-system",
		Enabled:  true,
	}
	values, err := a.prepareRegistryValues(userCtx, registryName, &er)

	if err != nil {
		return nil, errors.Wrap(err, "unable to prepare registry values")
	}

	if err := a.gerritService.CreateMergeRequestWithContents(userCtx, &gerrit.MergeRequest{
		ProjectName:   registryName,
		Name:          fmt.Sprintf("external-reg-system-mr-%s-%d", registryName, time.Now().Unix()),
		AuthorEmail:   ctx.GetString(router.UserEmailSessionKey),
		AuthorName:    ctx.GetString(router.UserNameSessionKey),
		CommitMessage: fmt.Sprintf("add new external reg system to registry %s", registryName),
		TargetBranch:  "master",
		Labels: map[string]string{
			mrLabelTarget: "external-reg",
		},
		Annotations: map[string]string{
			mrAnnotationRegName: er.Name,
			mrAnnotationRegType: ctx.PostForm("external-system-type"),
		},
	}, map[string]string{
		valuesLocation: values,
	}); err != nil {
		return nil, errors.Wrap(err, "unable to create MR with new values")
	}

	return router.MakeRedirectResponse(http.StatusFound,
		fmt.Sprintf("/admin/registry/view/%s", registryName)), nil
}

func (a *App) prepareRegistryValues(ctx context.Context, registryName string, er *ExternalRegistration) (string, error) {
	values, err := a.gerritService.GetFileContents(ctx, registryName, "master", valuesLocation)
	if err != nil {
		return "", errors.Wrap(err, "unable to get values yaml")
	}

	var valuesDict map[string]interface{}
	if err := yaml.Unmarshal([]byte(values), &valuesDict); err != nil {
		return "", errors.Wrap(err, "unable to decode values yaml")
	}
	if valuesDict == nil {
		valuesDict = make(map[string]interface{})
	}

	eRegs := make([]ExternalRegistration, 0)
	externalReg, ok := valuesDict["nontrembita-external-registration"]
	if ok {
		eRegs, ok = externalReg.([]ExternalRegistration)
		if !ok {
			return "", errors.New("wrong nontrembita-external-registration structure")
		}
	}

	eRegs = append(eRegs, ExternalRegistration{
		Name:     er.Name,
		Enabled:  er.Enabled,
		External: er.External,
	})
	valuesDict["nontrembita-external-registration"] = eRegs

	newValues, err := yaml.Marshal(valuesDict)
	if err != nil {
		return "", errors.Wrap(err, "unable to encode new values yaml")
	}

	return string(newValues), nil
}
