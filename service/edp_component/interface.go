package edpcomponent

type ServiceInterface interface {
	GetAll() ([]EDPComponent, error)
	GetAllNamespace(ns string, onlyVisible bool) ([]EDPComponent, error)
	Get(name string) (*EDPComponent, error)
}
