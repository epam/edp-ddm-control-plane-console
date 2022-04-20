package codebase

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path"
	"reflect"
	"strings"
	"time"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"ddm-admin-console/config"
	"ddm-admin-console/controller"
	"ddm-admin-console/registry"
	"ddm-admin-console/service"
	codebaseService "ddm-admin-console/service/codebase"
	gerritService "ddm-admin-console/service/gerrit"
	"ddm-admin-console/service/git"
)

const (
	adminsAnnotation    = "registry-parameters/administrators"
	defaultTempDir      = "/tmp"
	gitUserSecretName   = "gerrit-project-creator"
	gitUsername         = "project-creator"
	rootGitServerCRName = "gerrit"
)

type Controller struct {
	logger              controller.Logger
	mgr                 ctrl.Manager
	k8sClient           client.Client
	adminSyncer         AdminSyncer
	TempDir             string
	GitUserSecretName   string
	GitUsername         string
	_gerritSSHURL       string
	RootGitServerCRName string
	namespace           string
}

type AdminSyncer interface {
	SyncAdmins(ctx context.Context, registryName string, admins []registry.Admin) error
}

func Make(mgr ctrl.Manager, logger controller.Logger, adminSyncer AdminSyncer, cnf *config.Settings) error {
	c := Controller{
		mgr:                 mgr,
		logger:              logger,
		k8sClient:           mgr.GetClient(),
		adminSyncer:         adminSyncer,
		TempDir:             defaultTempDir,
		GitUserSecretName:   gitUserSecretName,
		GitUsername:         gitUsername,
		RootGitServerCRName: rootGitServerCRName,
		namespace:           cnf.Namespace,
		//_gerritSSHURL:     "ssh://localhost:37861", //for local development
	}

	if err := ctrl.NewControllerManagedBy(mgr).
		For(&codebaseService.Codebase{}, builder.WithPredicates(predicate.Funcs{
			UpdateFunc: isSpecUpdated})).
		Complete(&c); err != nil {
		return errors.Wrap(err, "unable to create controller")
	}

	return nil
}

func (c *Controller) gerritSSHURL() (string, error) {
	if c._gerritSSHURL != "" {
		return c._gerritSSHURL, nil
	}

	var rootGerrit codebaseService.GitServer
	if err := c.k8sClient.Get(context.Background(), types.NamespacedName{Namespace: c.namespace,
		Name: c.RootGitServerCRName}, &rootGerrit); err != nil {
		return "", errors.Wrap(err, "unable to get root gerrit")
	}

	c._gerritSSHURL = fmt.Sprintf("ssh://%s@%s:%d", c.GitUsername, rootGerrit.Spec.GitHost,
		rootGerrit.Spec.SshPort)
	return c._gerritSSHURL, nil
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
			c.logger.Error(err)
			result = reconcile.Result{RequeueAfter: time.Second * 10}
			//resultErr = errors.Wrap(err, "unable to reconcile codebase")
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

	if err := c.updateImportRepo(ctx, instance); err != nil {
		return errors.Wrap(err, "unable to update import repo")
	}

	return nil
}

func (c *Controller) updateImportRepo(ctx context.Context, instance *codebaseService.Codebase) error {
	if instance.Spec.GitUrlPath == nil || *instance.Spec.GitUrlPath != codebaseService.RepoNotReady {
		return nil
	}

	prj, err := c.getGerritProject(ctx, instance.Name)
	if service.IsErrNotFound(err) {
		return ErrPostpone(time.Second * 5)
	} else if err != nil {
		return errors.Wrap(err, "unknown error")
	}

	if prj.Status.Value != "OK" {
		return ErrPostpone(time.Second * 5)
	}

	if err := c.pushRegistryTemplate(ctx, instance); err != nil {
		return errors.Wrap(err, "unable to push registry template")
	}

	gitUrlPath := fmt.Sprintf("/%s", instance.Name)
	instance.Spec.GitUrlPath = &gitUrlPath

	if err := c.k8sClient.Update(ctx, instance); err != nil {
		return errors.Wrap(err, "unable to update codebase")
	}

	return nil
}

func (c *Controller) pushRegistryTemplate(ctx context.Context, instance *codebaseService.Codebase) error {
	var gerritSecret v1.Secret
	if err := c.k8sClient.Get(ctx, types.NamespacedName{Namespace: instance.Namespace, Name: c.GitUserSecretName},
		&gerritSecret); err != nil {
		return errors.Wrap(err, "unable to get gerrit project creator secret")
	}

	key, ok := gerritSecret.Data["id_rsa"]
	if !ok {
		return errors.New("no data by key id_rsa in gerrit secret")
	}

	tpl, ok := instance.Annotations[registry.TemplateNameAnnotation]
	if !ok {
		return errors.New("template annotation not found")
	}

	reposPath := path.Join(c.TempDir, "repos")
	if err := os.MkdirAll(reposPath, 0777); err != nil {
		return errors.Wrap(err, "unable to create repo folder")
	}

	gitService := git.Make(path.Join(reposPath, instance.Name), c.GitUsername, string(key))
	gerritSSHURL, err := c.gerritSSHURL()
	if err != nil {
		return errors.Wrap(err, "unable to get gerrit ssh url")
	}

	if err := gitService.Clone(fmt.Sprintf("%s/%s", gerritSSHURL, tpl)); err != nil {
		return errors.Wrap(err, "unable to clone repo")
	}

	if err := gitService.AddRemote("registry", fmt.Sprintf("%s/%s", gerritSSHURL, instance.Name)); err != nil {
		return errors.Wrap(err, "unable to add registry remote")
	}

	if err := gitService.Push("registry", "--all"); err != nil {
		return errors.Wrap(err, "unable to push [all] changes to repo registry")
	}

	if err := gitService.Push("registry", "--tags"); err != nil {
		return errors.Wrap(err, "unable to push tags to repo registry")
	}

	return nil
}

func (c *Controller) getGerritProject(ctx context.Context, name string) (*gerritService.GerritProject, error) {
	var projectList gerritService.GerritProjectList
	if err := c.k8sClient.List(ctx, &projectList); err != nil {
		return nil, errors.Wrap(err, "unable to list gerrit projects")
	}

	for _, prj := range projectList.Items {
		if prj.Spec.Name == name {
			return &prj, nil
		}
	}

	return nil, service.ErrNotFound("unable to find gerrit project")
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
