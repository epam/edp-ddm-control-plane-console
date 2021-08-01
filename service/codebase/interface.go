package codebase

import "context"

type ServiceInterface interface {
	GetAllByType(tp string) ([]Codebase, error)
	Get(name string) (*Codebase, error)
	GetBranchesByCodebase(codebaseName string) ([]CodebaseBranch, error)
	Create(cb *Codebase) error
	CreateBranch(branch *CodebaseBranch) error
	Update(cb *Codebase) error
	Delete(name string) error
	CreateDefaultBranch(cb *Codebase) error
	ServiceForContext(ctx context.Context) (ServiceInterface, error)
}
