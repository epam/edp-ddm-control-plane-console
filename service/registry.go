package service

import (
	"ddm-admin-console/k8s"
	"ddm-admin-console/models"
	"fmt"
	"time"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	registryConfigMapName           = "registry-config"
	registryNamespaceLabelKey       = "registry-env"
	registryUpdatedAtTimeFormat     = "2006-01-02 15:04:05"
	registryUpdatedAtConfigMapKey   = "updated_at"
	registryDescriptionConfigMapKey = "description"
)

type Registry struct {
	k8sClient k8s.CoreClient
	// env is needed to split namespaces that is created for development and for test cases
	env string
}

func MakeRegistry(k8sClient k8s.CoreClient, env string) *Registry {
	return &Registry{
		k8sClient: k8sClient,
		env:       env,
	}
}

type RegistryExistsError struct {
	Err error
}

func (r RegistryExistsError) Error() string {
	return r.Err.Error()
}

func (r *Registry) Create(name, description string) (*models.Registry, error) {
	_, err := r.k8sClient.Namespaces().Get(name, metav1.GetOptions{IncludeUninitialized: true})
	if err == nil {
		return nil, RegistryExistsError{
			Err: errors.New("unable to create registry, namespace with such name exists")}
	}

	ns, err := r.k8sClient.Namespaces().Create(&v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				registryNamespaceLabelKey: r.env,
			},
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to create registry, something went wrong")
	}

	if _, err := r.k8sClient.ConfigMaps(name).Create(&v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: registryConfigMapName},
		Data: map[string]string{
			registryDescriptionConfigMapKey: description,
			registryUpdatedAtConfigMapKey:   ns.ObjectMeta.CreationTimestamp.Time.Format(registryUpdatedAtTimeFormat),
		},
	}); err != nil {
		return nil, errors.Wrap(err, "unable to create namespace config map")
	}

	return &models.Registry{
		Name:        name,
		Description: description,
		CreatedAt:   ns.ObjectMeta.CreationTimestamp.Time,
		UpdatedAt:   ns.ObjectMeta.CreationTimestamp.Time,
	}, nil
}

func (r *Registry) Delete(name string) error {
	ns, err := r.k8sClient.Namespaces().Get(name, metav1.GetOptions{IncludeUninitialized: true})
	if err != nil {
		return errors.Wrap(err, "unable to get registry, namespace with such name does not exists")
	}

	if err := r.k8sClient.ConfigMaps(ns.Name).Delete(registryConfigMapName, &metav1.DeleteOptions{}); err != nil {
		return errors.Wrap(err, "unable to delete config map, something went wrong")
	}

	if err := r.k8sClient.Namespaces().Delete(ns.Name, &metav1.DeleteOptions{}); err != nil {
		return errors.Wrap(err, "unable to delete registry namespace, something went wrong")
	}

	return nil
}

func (r *Registry) Get(name string) (*models.Registry, error) {
	ns, err := r.k8sClient.Namespaces().Get(name, metav1.GetOptions{IncludeUninitialized: true})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get registry, namespace with such name does not exists")
	}

	return r.registryFromNamespace(ns)
}

func (r *Registry) registryFromNamespace(ns *v1.Namespace) (*models.Registry, error) {
	cm, err := r.k8sClient.ConfigMaps(ns.Name).Get(registryConfigMapName, metav1.GetOptions{IncludeUninitialized: true})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get registry namespace config map")
	}

	updatedAt, err := time.Parse(registryUpdatedAtTimeFormat, cm.Data[registryUpdatedAtConfigMapKey])
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse updated_at time from config map")
	}

	return &models.Registry{
		Name:        ns.Name,
		Description: cm.Data[registryDescriptionConfigMapKey],
		UpdatedAt:   updatedAt,
		CreatedAt:   ns.CreationTimestamp.Time,
		NS:          ns,
	}, nil
}

func (r *Registry) List() ([]*models.Registry, error) {
	nss, err := r.k8sClient.Namespaces().List(metav1.ListOptions{
		LabelSelector:        fmt.Sprintf("%s=%s", registryNamespaceLabelKey, r.env),
		IncludeUninitialized: true,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get registry list")
	}

	rgs := make([]*models.Registry, 0, len(nss.Items))
	for i := range nss.Items {
		rg, err := r.registryFromNamespace(&nss.Items[i])
		if err != nil {
			return nil, errors.Wrap(err, "unable to convert namespace to registry")
		}

		rgs = append(rgs, rg)
	}

	return rgs, nil
}

func (r *Registry) EditDescription(name, description string) error {
	ns, err := r.k8sClient.Namespaces().Get(name, metav1.GetOptions{IncludeUninitialized: true})
	if err != nil {
		return errors.Wrap(err, "unable to get registry, namespace with such name does not exists")
	}

	cm, err := r.k8sClient.ConfigMaps(ns.Name).Get(registryConfigMapName, metav1.GetOptions{IncludeUninitialized: true})
	if err != nil {
		return errors.Wrap(err, "unable to get registry config map")
	}

	cm.Data[registryDescriptionConfigMapKey] = description
	cm.Data[registryUpdatedAtConfigMapKey] = time.Now().Format(registryUpdatedAtTimeFormat)

	if _, err := r.k8sClient.ConfigMaps(ns.Name).Update(cm); err != nil {
		return errors.Wrap(err, "unable to update registry config map")
	}

	return nil
}
