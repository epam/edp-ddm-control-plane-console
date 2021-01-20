package edpcomponent

import (
	"ddm-admin-console/models/query"
	"fmt"
)

type Local struct {
	components map[string]*query.EDPComponent
}

type NotFound string

func (e NotFound) Error() string {
	return string(e)
}

func (s *Local) GetEDPComponent(componentType string) (*query.EDPComponent, error) {
	cmp, ok := s.components[componentType]
	if !ok {
		return nil, NotFound(fmt.Sprintf("%s not found", componentType))
	}

	return cmp, nil
}

func MakeLocal(components map[string]*query.EDPComponent) *Local {
	return &Local{
		components: components,
	}
}

func MakeLocalLinks(components map[string]string) *Local {
	edpComponents := make(map[string]*query.EDPComponent)
	for k, v := range components {
		edpComponents[k] = &query.EDPComponent{
			Type:    k,
			URL:     v,
			Visible: true,
		}
	}

	return MakeLocal(edpComponents)
}
