package codebase

import (
	"context"
	"encoding/base64"
	"reflect"
	"strings"

	"github.com/pkg/errors"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"ddm-admin-console/controller"
	"ddm-admin-console/registry"
	codebaseService "ddm-admin-console/service/codebase"
)

const (
	adminsAnnotation = "registry-parameters/administrators"
)

type Controller struct {
	logger      controller.Logger
	mgr         ctrl.Manager
	k8sClient   client.Client
	adminSyncer AdminSyncer
}

type AdminSyncer interface {
	SyncAdmins(ctx context.Context, registryName string, admins []registry.Admin) error
}

func Make(mgr ctrl.Manager, logger controller.Logger, adminSyncer AdminSyncer) error {
	c := Controller{
		mgr:         mgr,
		logger:      logger,
		k8sClient:   mgr.GetClient(),
		adminSyncer: adminSyncer,
	}

	if err := ctrl.NewControllerManagedBy(mgr).
		For(&codebaseService.Codebase{}, builder.WithPredicates(predicate.Funcs{
			UpdateFunc: isSpecUpdated})).
		Complete(&c); err != nil {
		return errors.Wrap(err, "unable to create controller")
	}
	return nil
}

func isSpecUpdated(e event.UpdateEvent) bool {
	oo := e.ObjectOld.(*codebaseService.Codebase)
	no := e.ObjectNew.(*codebaseService.Codebase)

	return !reflect.DeepEqual(oo.Spec, no.Spec) ||
		(oo.GetDeletionTimestamp().IsZero() && !no.GetDeletionTimestamp().IsZero())
}

func (c *Controller) Reconcile(ctx context.Context, request reconcile.Request) (result reconcile.Result,
	resultErr error) {
	c.logger.Infow("reconciling codebase", "Request.Namespace", request.Namespace,
		"Request.Name", request.Name)

	var instance codebaseService.Codebase
	if err := c.k8sClient.Get(ctx, request.NamespacedName, &instance); err != nil {
		if k8sErrors.IsNotFound(err) {
			c.logger.Infow("instance not found", "Request.Namespace", request.Namespace, "Request.Name", request.Name)
			return
		}

		resultErr = errors.Wrap(err, "unable to get codebase from k8s")
		return
	}

	if instance.Spec.Type == codebaseService.RegistryCodebaseType {
		if err := c.reconcile(ctx, &instance); err != nil {
			resultErr = errors.Wrap(err, "unable to reconcile codebase")
			return
		}
	}

	c.logger.Infow("reconciling done", "Request.Namespace", request.Namespace,
		"Request.Name", request.Name)

	return
}

func (c *Controller) reconcile(ctx context.Context, instance *codebaseService.Codebase) error {
	c.logger.Info(instance.Name)

	//TODO: remove in future release
	if err := c.migrateAdminAnnotations(ctx, instance); err != nil {
		return errors.Wrap(err, "unable to migrate admin annotations")
	}

	return nil
}

//deprecated
func (c *Controller) migrateAdminAnnotations(ctx context.Context, instance *codebaseService.Codebase) error {
	adminsEncoded, ok := instance.Annotations[adminsAnnotation]
	if !ok {
		return nil
	}

	adminsBuffer, err := base64.StdEncoding.DecodeString(adminsEncoded)
	if err != nil {
		return errors.Wrap(err, "unable to decode admins annotation")
	}

	var syncAdmins []registry.Admin
	admins := strings.Split(string(adminsBuffer), ",")
	for _, admin := range admins {
		c.logger.Infow("converting admin", "email", admin)
		syncAdmins = append(syncAdmins, registry.Admin{
			Username:  admin,
			Email:     admin,
			LastName:  "-",
			FirstName: "-",
		})
	}

	if err := c.adminSyncer.SyncAdmins(ctx, instance.Name, syncAdmins); err != nil {
		return errors.Wrap(err, "unable to sync admins")
	}

	annotations := instance.Annotations
	delete(annotations, adminsAnnotation)

	if err := c.k8sClient.Update(ctx, instance); err != nil {
		return errors.Wrap(err, "unable to update codebase")
	}

	return nil
}
