package test

import v1 "k8s.io/client-go/kubernetes/typed/core/v1"

type MockCoreClient struct {
	MockNamespaceInterface v1.NamespaceInterface
	MockConfigMapInterface v1.ConfigMapInterface
	MockSecretInterface    v1.SecretInterface
}

func (m MockCoreClient) Secrets(namespace string) v1.SecretInterface {
	return m.MockSecretInterface
}

func (m MockCoreClient) Namespaces() v1.NamespaceInterface {
	return m.MockNamespaceInterface
}

func (m MockCoreClient) ConfigMaps(namespace string) v1.ConfigMapInterface {
	return m.MockConfigMapInterface
}
