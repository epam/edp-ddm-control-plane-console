package k8s

import v1 "k8s.io/client-go/kubernetes/typed/core/v1"

type CoreClient interface {
	v1.SecretsGetter
	v1.NamespacesGetter
	v1.ConfigMapsGetter
}
