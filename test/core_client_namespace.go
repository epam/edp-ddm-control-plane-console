package test

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
)

type MockNamespaceInterface struct {
	CreateResult *v1.Namespace
	CreateError  error
	GetResult    *v1.Namespace
	GetError     error
	DeleteError  error
	ListResult   *v1.NamespaceList
	ListError    error
}

func (m MockNamespaceInterface) Create(*v1.Namespace) (*v1.Namespace, error) {
	return m.CreateResult, m.CreateError
}

func (MockNamespaceInterface) Update(*v1.Namespace) (*v1.Namespace, error) {
	return nil, nil
}

func (MockNamespaceInterface) UpdateStatus(*v1.Namespace) (*v1.Namespace, error) {
	return nil, nil
}

func (m MockNamespaceInterface) Delete(name string, options *metav1.DeleteOptions) error {
	return m.DeleteError
}

func (m MockNamespaceInterface) Get(name string, options metav1.GetOptions) (*v1.Namespace, error) {
	return m.GetResult, m.GetError
}

func (m MockNamespaceInterface) List(opts metav1.ListOptions) (*v1.NamespaceList, error) {
	return m.ListResult, m.ListError
}

func (MockNamespaceInterface) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return nil, nil
}

func (MockNamespaceInterface) Patch(name string, pt types.PatchType,
	data []byte, subresources ...string) (result *v1.Namespace, err error) {
	return nil, nil
}

func (MockNamespaceInterface) Finalize(item *v1.Namespace) (*v1.Namespace, error) {
	return nil, nil
}
