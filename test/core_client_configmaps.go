package test

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
)

type MockConfigMapInterface struct {
	CreateResult *v1.ConfigMap
	CreateError  error
	GetResult    *v1.ConfigMap
	GetError     error
	UpdateResult *v1.ConfigMap
	UpdateError  error
}

func (m MockConfigMapInterface) Create(*v1.ConfigMap) (*v1.ConfigMap, error) {
	return m.CreateResult, m.CreateError
}

func (m MockConfigMapInterface) Update(*v1.ConfigMap) (*v1.ConfigMap, error) {
	return m.UpdateResult, m.UpdateError
}

func (MockConfigMapInterface) Delete(name string, options *metav1.DeleteOptions) error {
	return nil
}

func (MockConfigMapInterface) DeleteCollection(options *metav1.DeleteOptions, listOptions metav1.ListOptions) error {
	return nil
}

func (m MockConfigMapInterface) Get(name string, options metav1.GetOptions) (*v1.ConfigMap, error) {
	return m.GetResult, m.GetError
}

func (MockConfigMapInterface) List(opts metav1.ListOptions) (*v1.ConfigMapList, error) {
	return nil, nil
}

func (MockConfigMapInterface) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return nil, nil
}

func (MockConfigMapInterface) Patch(name string, pt types.PatchType,
	data []byte, subresources ...string) (result *v1.ConfigMap, err error) {
	return nil, nil
}
