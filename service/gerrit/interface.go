package gerrit

import "context"

type ServiceInterface interface {
	GetProjects(ctx context.Context) ([]GerritProject, error)
}
