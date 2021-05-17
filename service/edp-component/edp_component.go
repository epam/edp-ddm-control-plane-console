package edpcomponent

import (
	"ddm-admin-console/models/query"
	ec "ddm-admin-console/repository/edp-component"
	"ddm-admin-console/service/logger"
	"ddm-admin-console/util"
	"ddm-admin-console/util/consts"
	dberrors "ddm-admin-console/util/error/db-errors"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

var log = logger.GetLogger()

type Service struct {
	IEDPComponent ec.IEDPComponentRepository
	Namespace     string
}

func MakeService(iEDPComponent ec.IEDPComponentRepository, namespace string) *Service {
	return &Service{
		IEDPComponent: iEDPComponent,
		Namespace:     namespace,
	}
}

// Service gets EDP component by type from DB
func (s Service) GetEDPComponent(componentType string) (*query.EDPComponent, error) {
	log.Debug("start fetching EDP Component", zap.String("type", componentType))
	c, err := s.IEDPComponent.GetEDPComponent(componentType)
	if err != nil {
		if dberrors.IsNotFound(err) {
			log.Debug("edp component wasn't found in DB", zap.String("name", componentType))
			return nil, nil
		}
		return nil, errors.Wrapf(err, "an error has occurred while fetching EDP Component by %v type from DB",
			componentType)
	}
	log.Info("edp component has been fetched from DB",
		zap.String("type", c.Type), zap.String("url", c.URL))
	return c, nil
}

// GetEDPComponents gets all EDP components from DB
func (s Service) GetEDPComponents() ([]*query.EDPComponent, error) {
	log.Debug("start fetching EDP Components...")
	c, err := s.IEDPComponent.GetEDPComponents()
	if err != nil {
		return nil, errors.Wrap(err, "an error has occurred while fetching EDP Components from DB")
	}
	log.Info("edp components have been fetched", zap.Any("length", len(c)))

	for i, v := range c {
		modifyPlatformLinks(v.URL, v.Type, s.Namespace, c[i])
	}

	return c, nil
}

func modifyPlatformLinks(url, componentType, namespace string, c *query.EDPComponent) {
	if componentType == consts.Openshift || componentType == consts.Kubernetes {
		c.URL = util.CreateNativeProjectLink(url, namespace)
	}
}
