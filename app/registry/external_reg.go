package registry

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"ddm-admin-console/router"
)

func (a *App) addExternalReg(ctx *gin.Context) (*router.Response, error) {
	userCtx := a.router.ContextWithUserAccessToken(ctx)

	registryName := ctx.Param("name")

	values, err := a.gerritService.GetFileContents(userCtx, registryName, "master", "deploy-templates/values.yaml")
	if err != nil {
		return nil, errors.Wrap(err, "unable to get values yaml")
	}

	var valuesDict map[string]interface{}
	if err := yaml.Unmarshal([]byte(values), &valuesDict); err != nil {
		return nil, errors.Wrap(err, "unable to decode values yaml")
	}

	eRegs := make([]ExternalRegistration, 0)
	externalReg, ok := valuesDict["nontrembita-external-registration"]
	if ok {
		eRegs, ok = externalReg.([]ExternalRegistration)
		if !ok {
			return nil, errors.New("wrong nontrembita-external-registration structure")
		}
	}

	eRegs = append(eRegs, ExternalRegistration{
		Name:     ctx.PostForm("reg-name"),
		Enabled:  true,
		External: ctx.PostForm("external-system-type") == "external-system",
	})
	valuesDict["nontrembita-external-registration"] = eRegs

	newValues, err := yaml.Marshal(valuesDict)
	if err != nil {
		return nil, errors.Wrap(err, "unable to encode new values yaml")
	}

	//CREATE MERGE REQUEST

	return nil, nil
}
