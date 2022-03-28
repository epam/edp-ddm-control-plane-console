package config

import (
	"go.uber.org/zap"

	"ddm-admin-console/service/keycloak"

	"ddm-admin-console/service/codebase"
	edpcomponent "ddm-admin-console/service/edp_component"
	"ddm-admin-console/service/gerrit"
	"ddm-admin-console/service/jenkins"
	"ddm-admin-console/service/k8s"
	"ddm-admin-console/service/openshift"
)

type Settings struct {
	HTTPPort                           string `envconfig:"HTTP_PORT" default:"8080"`
	LogLevel                           string `envconfig:"LOG_LEVEL" default:"INFO"`
	LogEncoding                        string `envconfig:"LOG_ENCODING" default:"json"`
	Namespace                          string `envconfig:"NAMESPACE" default:"default"`
	SessionSecret                      string `envconfig:"SESSION_SECRET" default:"UdWaTEfULunPTkRC9sFLG26APz9W5gEC8x"`
	OCClientID                         string `envconfig:"OC_CLIENT_ID"`
	OCClientSecret                     string `envconfig:"OC_CLIENT_SECRET"`
	Host                               string `envconfig:"HOST"`
	GerritCreatorSecretName            string `envconfig:"GERRIT_CREATOR_SECRET_NAME"`
	ClusterCodebaseName                string `envconfig:"CLUSTER_CODEBASE_NAME"`
	ClusterRepo                        string `envconfig:"CLUSTER_REPO"`
	BackupSecretName                   string `envconfig:"BACKUP_SECRET_NAME" default:"backup-credential"`
	GinMode                            string `envconfig:"GIN_MODE"`
	Timezone                           string `envconfig:"TIMEZONE" default:"Europe/Kiev"`
	RegistryRepoPrefix                 string `envconfig:"REGISTRY_REPO_PREFIX" default:"registry-tenant-template-"`
	RegistryRepoHost                   string `envconfig:"REGISTRY_REPO_HOST"`
	RegistryHardwareKeyINITemplatePath string `envconfig:"REGISTRY_HW_KEY_INI_TPL_PATH" default:"osplm.ini"`
	RootGerritName                     string `envconfig:"ROOT_GERRIT_NAME" default:"gerrit"`
	GroupGitRepo                       string `envconfig:"GROUP_GIT_REPO"`
	UsersNamespace                     string `envconfig:"USERS_NAMESPACE" default:"user-management"`
	UsersRealm                         string `envconfig:"USERS_REALM" default:"openshift"`
	EnableBranchProvisioners           bool   `envconfig:"ENABLE_BRANCH_PROVISIONERS"`
}

type Services struct {
	Codebase     codebase.ServiceInterface
	EDPComponent edpcomponent.ServiceInterface
	K8S          k8s.ServiceInterface
	OpenShift    openshift.ServiceInterface
	Gerrit       gerrit.ServiceInterface
	Jenkins      jenkins.ServiceInterface
	Keycloak     keycloak.ServiceInterface
}

type Logger interface {
	Error(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
}
