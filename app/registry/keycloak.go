package registry

import (
	"context"
	"fmt"
	"net/url"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"

	edpComponent "ddm-admin-console/service/edp_component"
)

func (a *App) prepareKeycloakCustomHostname(ctx *gin.Context, r *registry, values *Values,
	_ map[string]map[string]interface{}, _ *[]string,
) (bool, error) {
	keycloakDefaultHostname, err := LoadKeycloakDefaultHostname(ctx, a.KeycloakDefaultHostname, a.EDPComponent)
	if err != nil {
		return false, fmt.Errorf("unable to load keycloak default hostname")
	}

	valuesChanged := false

	if keycloakDefaultHostname == r.KeycloakCustomHostname {
		valuesChanged = values.Keycloak.CustomHost != ""
		values.Keycloak.CustomHost = ""
	} else if r.KeycloakCustomHostname != "" && keycloakDefaultHostname != r.KeycloakCustomHostname {
		valuesChanged = values.Keycloak.CustomHost != r.KeycloakCustomHostname
		values.Keycloak.CustomHost = r.KeycloakCustomHostname
	}

	if !valuesChanged {
		return false, nil
	}

	values.OriginalYaml["keycloak"] = values.Keycloak
	return true, nil
}

func (a *App) loadKeycloakHostnames() ([]string, error) {
	values, err := a.getClusterValues()
	if err != nil {
		return nil, fmt.Errorf("unable to get cluster values, %w", err)
	}

	hostnames := make([]string, 0, len(values.Keycloak.CustomHosts))

	for _, v := range values.Keycloak.CustomHosts {
		hostnames = append(hostnames, v.Host)
	}

	return hostnames, nil
}

func (a *App) getClusterValues() (*ClusterValues, error) {
	data, err := a.Gerrit.GetFileFromBranch(a.ClusterCodebaseName, MasterBranch, url.PathEscape(ValuesLocation))
	if err != nil {
		return nil, fmt.Errorf("unable to get cluster values")
	}

	var values ClusterValues
	if err := yaml.Unmarshal([]byte(data), &values); err != nil {
		return nil, fmt.Errorf("unable to decode cluster values")
	}

	return &values, nil
}

func LoadKeycloakDefaultHostname(ctx context.Context, envValue string, edpService edpComponent.ServiceInterface) (string, error) {
	if envValue != "" {
		return envValue, nil
	}

	comp, err := edpService.Get(ctx, "main-keycloak")
	if err != nil {
		return "", fmt.Errorf("unable to get edp component, %w", err)
	}

	urlData, err := url.Parse(comp.Spec.Url)
	if err != nil {
		return "", fmt.Errorf("unabe to parse url, %w", err)
	}

	return urlData.Host, nil
}
