package service

import (
	"ddm-admin-console/k8s"
	"ddm-admin-console/test"
	"testing"
	"time"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	namespaceNotFoundMessage = `namespaces "unknown-registry" not found`
	configMapNotFoundMessage = `configmaps "registry-config" not found`
)

func TestRegistry_Create(t *testing.T) {
	clients := k8s.CreateOpenShiftClients()
	regService := MakeRegistry(clients.CoreClient, "test")

	if _, err := regService.Create("foo14", "bar"); err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := regService.Delete("foo14"); err != nil {
			t.Fatal(err)
		}
	}()

	if _, err := regService.Get("foo14"); err != nil {
		t.Fatal(err)
	}
}

func TestRegistry_CreateFailure_NamespaceExists(t *testing.T) {
	clients := k8s.CreateOpenShiftClients()
	regService := MakeRegistry(clients.CoreClient, "test")

	if _, err := regService.Create("foo31", "bar"); err != nil {
		t.Fatal(err)
	}

	_, err := regService.Create("foo31", "bar")
	if err == nil {
		t.Fatal("no error on duplicate name")
	}

	defer func() {
		if err := regService.Delete("foo31"); err != nil {
			t.Fatal(err)
		}
	}()

	switch errors.Cause(err).(type) {
	case RegistryExistsError:
		if errors.Cause(err).(RegistryExistsError).Err == nil {
			t.Fatal("registry error not set")
		}
	default:
		t.Fatal("type of error is not RegistryExistsError")
	}
}

func TestRegistryExistsError_Error(t *testing.T) {
	testErr := "test"
	err := errors.New(testErr)
	ree := RegistryExistsError{
		Err: err,
	}

	if ree.Error() != testErr {
		t.Fatal("wrong error returned")
	}
}

func TestRegistry_CreateFailure_CreationError(t *testing.T) {
	mockError := errors.New("error during namespace creation")
	mockCoreClient := test.MockCoreClient{
		MockNamespaceInterface: test.MockNamespaceInterface{
			CreateError: mockError,
			GetError:    mockError,
		},
	}
	regService := MakeRegistry(mockCoreClient, "test")

	_, err := regService.Create("foo31", "bar")
	if err == nil {
		t.Fatal("no error on namespace creation failure")
	}

	if errors.Cause(err) != mockError {
		t.Fatal("wrong type of error returned")
	}
}

func TestRegistry_CreateFailure_ConfigMapError(t *testing.T) {
	mockError := errors.New("error during config map creation")

	mockCoreClient := test.MockCoreClient{
		MockNamespaceInterface: test.MockNamespaceInterface{
			CreateResult: &v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					CreationTimestamp: metav1.Time{
						Time: time.Now(),
					},
				},
			},
			GetError: mockError,
		},
		MockConfigMapInterface: test.MockConfigMapInterface{
			CreateError: mockError,
		},
	}
	regService := MakeRegistry(mockCoreClient, "test")

	_, err := regService.Create("foo31", "bar")
	if err == nil {
		t.Fatal("no error on namespace creation failure")
	}

	if errors.Cause(err) != mockError {
		t.Fatal("wrong type of error returned")
	}
}

func TestRegistry_List(t *testing.T) {
	clients := k8s.CreateOpenShiftClients()
	pseudoProdRegService := MakeRegistry(clients.CoreClient, "pseudo-test")

	if _, err := pseudoProdRegService.Create("foo55", "bar"); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := pseudoProdRegService.Delete("foo55"); err != nil {
			t.Fatal(err)
		}
	}()

	testRegService := MakeRegistry(clients.CoreClient, "test2")

	if _, err := testRegService.Create("foo61", "bar"); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := testRegService.Delete("foo61"); err != nil {
			t.Fatal(err)
		}
	}()

	if _, err := testRegService.Create("foo65", "bar"); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := pseudoProdRegService.Delete("foo65"); err != nil {
			t.Fatal(err)
		}
	}()

	rgs, err := testRegService.List()
	if err != nil {
		t.Fatal(err)
	}

	if len(rgs) != 2 {
		t.Fatal("wrong namespaces count")
	}
}

func TestRegistry_EditDescription(t *testing.T) {
	clients := k8s.CreateOpenShiftClients()
	regService := MakeRegistry(clients.CoreClient, "test")

	if _, err := regService.Create("foo95", "bar"); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := regService.Delete("foo95"); err != nil {
			t.Fatal(err)
		}
	}()

	newDescription := "new description"

	if err := regService.EditDescription("foo95", newDescription); err != nil {
		t.Fatal(err)
	}

	rg, err := regService.Get("foo95")
	if err != nil {
		t.Fatal(err)
	}

	if rg.Description != newDescription {
		t.Fatal("description is not updated")
	}
}

func TestRegistry_DeleteFailure_NamespaceNotExists(t *testing.T) {
	clients := k8s.CreateOpenShiftClients()
	regService := MakeRegistry(clients.CoreClient, "test")
	err := regService.Delete("unknown-registry")
	if err == nil {
		t.Fatal("no error on trying to delete not existing registry")
	}

	switch errors.Cause(err).(type) {
	case *k8sErrors.StatusError:
		if errors.Cause(err).(*k8sErrors.StatusError).ErrStatus.Message != namespaceNotFoundMessage {
			t.Fatal("wrong k8s status error")
		}
	default:
		t.Fatal("wrong error type")
	}
}

func TestRegistry_DeleteFailure_ConfigMapMissing(t *testing.T) {
	clients := k8s.CreateOpenShiftClients()
	regService := MakeRegistry(clients.CoreClient, "test-delete")

	ns, err := regService.Create("foo208", "bar")
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := clients.CoreClient.Namespaces().Delete(ns.Name, &metav1.DeleteOptions{}); err != nil {
			t.Fatal(err)
		}
	}()

	if err := clients.CoreClient.ConfigMaps(ns.Name).Delete(registryConfigMapName,
		&metav1.DeleteOptions{}); err != nil {
		t.Fatal(err)
	}

	err = regService.Delete(ns.Name)
	if err == nil {
		t.Fatal("no error on deleted config map")
	}

	switch errors.Cause(err).(type) {
	case *k8sErrors.StatusError:
		if errors.Cause(err).(*k8sErrors.StatusError).ErrStatus.Message != configMapNotFoundMessage {
			t.Fatal("wrong k8s status error")
		}
	default:
		t.Fatal("wrong error type")
	}
}

func TestRegistry_DeleteFailure_DeletionError(t *testing.T) {
	mockError := errors.New("error during namespace deletion")

	mockCoreClient := test.MockCoreClient{
		MockNamespaceInterface: test.MockNamespaceInterface{
			GetResult:   &v1.Namespace{},
			DeleteError: mockError,
		},
		MockConfigMapInterface: test.MockConfigMapInterface{},
	}
	regService := MakeRegistry(mockCoreClient, "test")

	err := regService.Delete("test")
	if err == nil {
		t.Fatal("no error on namespace deletion failure")
	}

	if errors.Cause(err) != mockError {
		t.Fatal("wrong type of error returned")
	}
}

func TestRegistry_GetFailure_NotExists(t *testing.T) {
	clients := k8s.CreateOpenShiftClients()
	regService := MakeRegistry(clients.CoreClient, "test")
	_, err := regService.Get("unknown-registry")
	if err == nil {
		t.Fatal("no error on trying to get not existing registry")
	}

	switch errors.Cause(err).(type) {
	case *k8sErrors.StatusError:
		if errors.Cause(err).(*k8sErrors.StatusError).ErrStatus.Message != namespaceNotFoundMessage {
			t.Fatal("wrong k8s status error")
		}
	default:
		t.Fatal("wrong error type")
	}
}

func TestRegistry_GetFailure_ConfigMap(t *testing.T) {
	clients := k8s.CreateOpenShiftClients()
	regService := MakeRegistry(clients.CoreClient, "test-delete")

	ns, err := regService.Create("foo282", "bar")
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := clients.CoreClient.Namespaces().Delete(ns.Name, &metav1.DeleteOptions{}); err != nil {
			t.Fatal(err)
		}
	}()

	if err := clients.CoreClient.ConfigMaps(ns.Name).Delete(registryConfigMapName,
		&metav1.DeleteOptions{}); err != nil {
		t.Fatal(err)
	}

	_, err = regService.Get(ns.Name)
	if err == nil {
		t.Fatal("no error on deleted config map")
	}

	switch errors.Cause(err).(type) {
	case *k8sErrors.StatusError:
		if errors.Cause(err).(*k8sErrors.StatusError).ErrStatus.Message != configMapNotFoundMessage {
			t.Fatal("wrong k8s status error")
		}
	default:
		t.Fatal("wrong error type")
	}
}

func TestRegistry_GetFailure_ConfigMapUpdatedAt(t *testing.T) {
	clients := k8s.CreateOpenShiftClients()
	regService := MakeRegistry(clients.CoreClient, "test-delete")

	ns, err := regService.Create("317", "bar")
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := regService.Delete(ns.Name); err != nil {
			t.Fatal(err)
		}
	}()

	cm, err := clients.CoreClient.ConfigMaps(ns.Name).Get(registryConfigMapName,
		metav1.GetOptions{IncludeUninitialized: true})
	if err != nil {
		t.Fatal(err)
	}
	cm.Data[registryUpdatedAtConfigMapKey] = "wrong time format"

	if _, err := clients.CoreClient.ConfigMaps(ns.Name).Update(cm); err != nil {
		t.Fatal(err)
	}

	_, err = regService.Get(ns.Name)
	if err == nil {
		t.Fatal("no error on deleted config map")
	}

	switch errors.Cause(err).(type) {
	case *time.ParseError:
		if errors.Cause(err).(*time.ParseError).Layout != registryUpdatedAtTimeFormat {
			t.Fatal("wrong time format error")
		}
	default:
		t.Fatal("wrong error type")
	}
}

func TestRegistry_ListFailure(t *testing.T) {
	mockError := errors.New("error during namespace list")

	mockCoreClient := test.MockCoreClient{
		MockNamespaceInterface: test.MockNamespaceInterface{
			ListError: mockError,
		},
	}
	regService := MakeRegistry(mockCoreClient, "test")

	_, err := regService.List()
	if err == nil {
		t.Fatal("no error on namespace list failure")
	}

	if errors.Cause(err) != mockError {
		t.Fatal("wrong type of error returned")
	}
}

func TestRegistry_ListFailure_ConfigMap(t *testing.T) {
	mockError := errors.New("error during get config map")

	mockCoreClient := test.MockCoreClient{
		MockNamespaceInterface: test.MockNamespaceInterface{
			ListResult: &v1.NamespaceList{
				Items: []v1.Namespace{
					{},
				},
			},
		},
		MockConfigMapInterface: test.MockConfigMapInterface{
			GetError: mockError,
		},
	}
	regService := MakeRegistry(mockCoreClient, "test")

	_, err := regService.List()
	if err == nil {
		t.Fatal("no error on namespace list failure")
	}

	if errors.Cause(err) != mockError {
		t.Fatal("wrong type of error returned")
	}
}

func TestRegistry_EditDescriptionFailure_NotExists(t *testing.T) {
	clients := k8s.CreateOpenShiftClients()
	regService := MakeRegistry(clients.CoreClient, "test")
	err := regService.EditDescription("unknown-registry", "description")
	if err == nil {
		t.Fatal("no error on trying to get not existing registry")
	}

	switch errors.Cause(err).(type) {
	case *k8sErrors.StatusError:
		if errors.Cause(err).(*k8sErrors.StatusError).ErrStatus.Message != namespaceNotFoundMessage {
			t.Fatal("wrong k8s status error")
		}
	default:
		t.Fatal("wrong error type")
	}
}

func TestRegistry_EditDescription_ConfigMap(t *testing.T) {
	clients := k8s.CreateOpenShiftClients()
	regService := MakeRegistry(clients.CoreClient, "test-delete")

	ns, err := regService.Create("foo423", "bar")
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := clients.CoreClient.Namespaces().Delete(ns.Name, &metav1.DeleteOptions{}); err != nil {
			t.Fatal(err)
		}
	}()

	if err := clients.CoreClient.ConfigMaps(ns.Name).Delete(registryConfigMapName,
		&metav1.DeleteOptions{}); err != nil {
		t.Fatal(err)
	}

	err = regService.EditDescription(ns.Name, "desc")
	if err == nil {
		t.Fatal("no error on deleted config map")
	}

	switch errors.Cause(err).(type) {
	case *k8sErrors.StatusError:
		if errors.Cause(err).(*k8sErrors.StatusError).ErrStatus.Message != configMapNotFoundMessage {
			t.Fatal("wrong k8s status error")
		}
	default:
		t.Fatal("wrong error type")
	}
}

func TestRegistry_EditDescription_Updating(t *testing.T) {
	mockError := errors.New("error during config map updating")

	mockCoreClient := test.MockCoreClient{
		MockNamespaceInterface: test.MockNamespaceInterface{
			GetResult: &v1.Namespace{},
		},
		MockConfigMapInterface: test.MockConfigMapInterface{
			GetResult: &v1.ConfigMap{
				Data: make(map[string]string),
			},
			UpdateError: mockError,
		},
	}

	regService := MakeRegistry(mockCoreClient, "test")

	err := regService.EditDescription("foo", "bar")
	if err == nil {
		t.Fatal("no error on edit description failure")
	}

	if errors.Cause(err) != mockError {
		t.Fatal("wrong type of error returned")
	}
}
