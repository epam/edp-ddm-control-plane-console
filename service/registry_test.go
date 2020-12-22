package service

import (
	"ddm-admin-console/test"
	"testing"
	"time"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	namespaceNotFoundMessage = `namespaces "unknown-registry" not found`
	configMapNotFoundMessage = `configmaps "registry-config" not found`
)

func TestRegistry_Create(t *testing.T) {
	coreClientMock := test.MockCoreClient{
		MockNamespaceInterface: test.MockNamespaceInterface{
			GetError:     k8sErrors.NewNotFound(schema.GroupResource{}, "namespace"),
			CreateResult: &v1.Namespace{},
		},
		MockConfigMapInterface: test.MockConfigMapInterface{
			CreateResult: &v1.ConfigMap{},
		},
	}
	regService := MakeRegistry(coreClientMock, "test")

	if _, err := regService.Create("foo14", "bar"); err != nil {
		t.Fatal(err)
	}
}

func TestRegistry_CreateFailure_NamespaceExists(t *testing.T) {
	coreClientMock := test.MockCoreClient{
		MockNamespaceInterface: test.MockNamespaceInterface{},
	}
	regService := MakeRegistry(coreClientMock, "test")

	_, err := regService.Create("foo31", "bar")
	if err == nil {
		t.Fatal("no error on duplicate name")
	}

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
	mockCoreClient := test.MockCoreClient{
		MockNamespaceInterface: test.MockNamespaceInterface{
			ListResult: &v1.NamespaceList{
				Items: []v1.Namespace{
					{},
					{},
				},
			},
		},
		MockConfigMapInterface: test.MockConfigMapInterface{
			GetResult: &v1.ConfigMap{
				Data: map[string]string{
					registryUpdatedAtConfigMapKey: time.Now().Format(registryUpdatedAtTimeFormat),
				},
			},
		},
	}
	regService := MakeRegistry(mockCoreClient, "test")

	rgs, err := regService.List()
	if err != nil {
		t.Fatal(err)
	}

	if len(rgs) != 2 {
		t.Fatal("wrong namespaces count")
	}
}

func TestRegistry_EditDescription(t *testing.T) {
	coreClientMock := test.MockCoreClient{
		MockNamespaceInterface: test.MockNamespaceInterface{
			GetResult: &v1.Namespace{},
		},
		MockConfigMapInterface: test.MockConfigMapInterface{
			GetResult: &v1.ConfigMap{
				Data: map[string]string{},
			},
		},
	}
	regService := MakeRegistry(coreClientMock, "test")

	if err := regService.EditDescription("foo95", "new description"); err != nil {
		t.Fatal(err)
	}
}

func TestRegistry_DeleteFailure_NamespaceNotExists(t *testing.T) {
	getError := k8sErrors.NewNotFound(schema.GroupResource{}, "namespace")
	getError.ErrStatus.Message = namespaceNotFoundMessage

	coreClientMock := test.MockCoreClient{
		MockNamespaceInterface: test.MockNamespaceInterface{
			GetError: getError,
		},
	}
	regService := MakeRegistry(coreClientMock, "test")
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
	deleteError := k8sErrors.NewNotFound(schema.GroupResource{}, "configMap")
	deleteError.ErrStatus.Message = configMapNotFoundMessage

	coreClientMock := test.MockCoreClient{
		MockNamespaceInterface: test.MockNamespaceInterface{
			GetResult: &v1.Namespace{},
		},
		MockConfigMapInterface: test.MockConfigMapInterface{
			DeleteError: deleteError,
		},
	}
	regService := MakeRegistry(coreClientMock, "test-delete")

	err := regService.Delete("test")
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
	getError := k8sErrors.NewNotFound(schema.GroupResource{}, "namespace")
	getError.ErrStatus.Message = namespaceNotFoundMessage

	mockCoreClient := test.MockCoreClient{
		MockNamespaceInterface: test.MockNamespaceInterface{
			GetError: getError,
		},
	}
	regService := MakeRegistry(mockCoreClient, "test")
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
	getError := k8sErrors.NewNotFound(schema.GroupResource{}, "configMap")
	getError.ErrStatus.Message = configMapNotFoundMessage

	mockCoreClient := test.MockCoreClient{
		MockNamespaceInterface: test.MockNamespaceInterface{
			GetResult: &v1.Namespace{},
		},
		MockConfigMapInterface: test.MockConfigMapInterface{
			GetError: getError,
		},
	}

	regService := MakeRegistry(mockCoreClient, "test-delete")

	_, err := regService.Get("test")
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
	mockCoreClient := test.MockCoreClient{
		MockNamespaceInterface: test.MockNamespaceInterface{
			GetResult: &v1.Namespace{},
		},
		MockConfigMapInterface: test.MockConfigMapInterface{
			GetResult: &v1.ConfigMap{
				Data: map[string]string{
					registryUpdatedAtConfigMapKey: "wrong time format",
				},
			},
		},
	}
	regService := MakeRegistry(mockCoreClient, "test-delete")

	_, err := regService.Get("test")
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
	getError := k8sErrors.NewNotFound(schema.GroupResource{}, "namespace")
	getError.ErrStatus.Message = namespaceNotFoundMessage

	mockCoreClient := test.MockCoreClient{
		MockNamespaceInterface: test.MockNamespaceInterface{
			GetError: getError,
		},
	}
	regService := MakeRegistry(mockCoreClient, "test")
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

func TestRegistry_EditDescriptionFailure_ConfigMap(t *testing.T) {
	getError := k8sErrors.NewNotFound(schema.GroupResource{}, "configMap")
	getError.ErrStatus.Message = configMapNotFoundMessage

	mockCoreClient := test.MockCoreClient{
		MockNamespaceInterface: test.MockNamespaceInterface{
			GetResult: &v1.Namespace{},
		},
		MockConfigMapInterface: test.MockConfigMapInterface{
			GetError: getError,
		},
	}
	regService := MakeRegistry(mockCoreClient, "test")

	err := regService.EditDescription("test", "desc")
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

func TestRegistry_Delete(t *testing.T) {
	mockCoreClient := test.MockCoreClient{
		MockNamespaceInterface: test.MockNamespaceInterface{
			GetResult: &v1.Namespace{},
		},
		MockConfigMapInterface: test.MockConfigMapInterface{},
	}
	regService := MakeRegistry(mockCoreClient, "test")

	err := regService.Delete("test")
	if err != nil {
		t.Fatal(err)
	}
}
