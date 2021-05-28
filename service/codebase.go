/*
 * Copyright 2020 EPAM Systems.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package service

import (
	"context"
	"ddm-admin-console/k8s"
	"ddm-admin-console/models"
	"ddm-admin-console/models/command"
	edperror "ddm-admin-console/models/error"
	"ddm-admin-console/models/query"
	"ddm-admin-console/repository"
	"ddm-admin-console/service/logger"
	"ddm-admin-console/util/consts"
	"encoding/base64"
	"fmt"
	"time"

	edpv1alpha1 "github.com/epmd-edp/codebase-operator/v2/pkg/apis/edp/v1alpha1"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

var log = logger.GetLogger()

const (
	gerritCreatorUsername = "user"
	gerritCreatorPassword = "password"
)

type CodebaseService struct {
	Clients                 *k8s.ClientSet
	ICodebaseRepository     repository.ICodebaseRepository
	BranchService           *CodebaseBranchService
	GerritCreatorSecretName string
	Namespace               string
}

func MakeCodebaseService(clients *k8s.ClientSet,
	iCodebaseRepository repository.ICodebaseRepository,
	branchService *CodebaseBranchService,
	gerritCreatorSecretName,
	namespace string) *CodebaseService {
	return &CodebaseService{
		Clients:                 clients,
		ICodebaseRepository:     iCodebaseRepository,
		Namespace:               namespace,
		GerritCreatorSecretName: gerritCreatorSecretName,
		BranchService:           branchService,
	}
}

func (s CodebaseService) CreateCodebase(codebase command.CreateCodebase) (*edpv1alpha1.Codebase, error) {
	log.Info("start creating Codebase resource", zap.String("name", codebase.Name))

	codebaseCr, err := GetCodebaseCR(s.Clients.EDPRestClientV2, codebase.Name, s.Namespace)
	if err != nil && !k8sErrors.IsNotFound(errors.Cause(err)) {
		log.Info("an error has occurred while fetching Codebase CR from cluster",
			zap.String("name", codebase.Name))
		return nil, err
	}
	if codebaseCr != nil {
		log.Info("codebase already exists in cluster", zap.String("name", codebaseCr.Name))
		return nil, edperror.NewCodebaseAlreadyExistsError()
	}

	edpClient := s.Clients.EDPRestClientV2

	annotations := make(map[string]string)
	if codebase.Admins != "" {
		annotations[consts.AdminsAnnotation] = base64.StdEncoding.EncodeToString([]byte(codebase.Admins))
	}

	c := &edpv1alpha1.Codebase{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v2.edp.epam.com/v1alpha1",
			Kind:       consts.CodebaseKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        codebase.Name,
			Namespace:   s.Namespace,
			Finalizers:  []string{"foregroundDeletion"},
			Annotations: annotations,
		},
		Spec: convertData(codebase),
		Status: edpv1alpha1.CodebaseStatus{
			Available:       false,
			LastTimeUpdated: time.Now(),
			Status:          "initialized",
			Username:        codebase.Username,
			Action:          "codebase_registration",
			Result:          "success",
			Value:           "inactive",
		},
	}
	log.Debug("CR was generated. Waiting to save ...", zap.String("name", c.Name))

	if err := s.createTempSecrets(codebase); err != nil {
		return nil, err
	}

	result := &edpv1alpha1.Codebase{}
	err = edpClient.Post().Namespace(s.Namespace).Resource(consts.CodebasePlural).Body(c).Do().Into(result)
	if err != nil {
		log.Error("an error has occurred while creating codebase resource in cluster", zap.Error(err))
		return &edpv1alpha1.Codebase{}, err
	}

	p := setCodebaseBranchCr(codebase.Versioning.Type, codebase.Username, codebase.Versioning.StartFrom, codebase.DefaultBranch)

	if _, err = s.BranchService.CreateCodebaseBranch(p, codebase.Name); err != nil {
		log.Error("an error has been occurred during the master branch creation", zap.Error(err))
		return &edpv1alpha1.Codebase{}, err
	}
	return result, nil
}

const (
	key6SecretKey                        = "Key-6.dat"
	digitalSignatureKeyIssuerSecretKey   = "digital-signature-key-issuer"
	digitalSignatureKeyPasswordSecretKey = "digital-signature-key-password"
	caCertificatesSecretKey              = "CACertificates.p7b"
	CAsJSONSecretKey                     = "CAs.json"
)

func (s *CodebaseService) UpdateKeySecret(key6, caCert, casJSON []byte, signKeyIssuer, signKeyPwd,
	registryName string) error {
	keySecret, err := s.Clients.CoreClient.Secrets(registryName).Get(
		fmt.Sprintf("system-digital-sign-%s-key", registryName), metav1.GetOptions{})
	if err != nil {
		return errors.Wrap(err, "unable to get secret")
	}

	keySecret.Data = map[string][]byte{
		key6SecretKey:                        key6,
		digitalSignatureKeyIssuerSecretKey:   []byte(signKeyIssuer),
		digitalSignatureKeyPasswordSecretKey: []byte(signKeyPwd),
	}

	if _, err := s.Clients.CoreClient.Secrets(registryName).Update(keySecret); err != nil {
		return errors.Wrap(err, "unable to update secret")
	}

	caSecret, err := s.Clients.CoreClient.Secrets(registryName).Get(
		fmt.Sprintf("system-digital-sign-%s-ca", registryName), metav1.GetOptions{})
	if err != nil {
		return errors.Wrap(err, "unable to get secret")
	}

	caSecret.Data = map[string][]byte{
		caCertificatesSecretKey: caCert,
		CAsJSONSecretKey:        casJSON,
	}

	if _, err := s.Clients.CoreClient.Secrets(registryName).Update(caSecret); err != nil {
		return errors.Wrap(err, "unable to update secret")
	}

	return nil
}

func (s *CodebaseService) CreateKeySecret(key6, caCert, casJSON []byte, signKeyIssuer, signKeyPwd,
	registryName string) error {

	keysSecretName := fmt.Sprintf("system-digital-sign-%s-key", registryName)
	if _, err := s.Clients.CoreClient.Secrets(s.Namespace).Get(keysSecretName, metav1.GetOptions{}); err == nil {
		if err := s.Clients.CoreClient.Secrets(s.Namespace).Delete(keysSecretName, &metav1.DeleteOptions{}); err != nil {
			return errors.Wrapf(err, "unable to delete secret: %s", keysSecretName)
		}
	}

	if _, err := s.Clients.CoreClient.Secrets(s.Namespace).Create(&v1.Secret{Data: map[string][]byte{
		key6SecretKey:                        key6,
		digitalSignatureKeyIssuerSecretKey:   []byte(signKeyIssuer),
		digitalSignatureKeyPasswordSecretKey: []byte(signKeyPwd),
	}, ObjectMeta: metav1.ObjectMeta{
		Name:      keysSecretName,
		Namespace: s.Namespace,
	}}); err != nil {
		return errors.Wrap(err, "unable to create registry key secret")
	}

	certsSecretName := fmt.Sprintf("system-digital-sign-%s-ca", registryName)
	if _, err := s.Clients.CoreClient.Secrets(s.Namespace).Get(certsSecretName, metav1.GetOptions{}); err == nil {
		if err := s.Clients.CoreClient.Secrets(s.Namespace).Delete(certsSecretName, &metav1.DeleteOptions{}); err != nil {
			return errors.Wrapf(err, "unable to delete secret: %s", certsSecretName)
		}
	}
	if _, err := s.Clients.CoreClient.Secrets(s.Namespace).Create(&v1.Secret{Data: map[string][]byte{
		caCertificatesSecretKey: caCert,
		CAsJSONSecretKey:        casJSON,
	}, ObjectMeta: metav1.ObjectMeta{
		Name:      certsSecretName,
		Namespace: s.Namespace,
	}}); err != nil {
		return errors.Wrap(err, "unable to create registry key secret")
	}

	return nil
}

func (s *CodebaseService) GetCodebasesByCriteria(criteria query.CodebaseCriteria) ([]*query.Codebase, error) {
	codebases, err := s.ICodebaseRepository.GetCodebasesByCriteria(criteria)
	if err != nil {
		log.Error("an error has occurred while getting codebase objects", zap.Error(err))
		return nil, err
	}
	log.Debug("fetched codebases", zap.Int("count", len(codebases)))

	return codebases, nil
}

func (s *CodebaseService) k8sCodebase2queryCodebase(clientV2 rest.Interface,
	cb *edpv1alpha1.Codebase, loadBranches bool) (*query.Codebase, error) {

	description := ""
	if cb.Spec.Description != nil {
		description = *cb.Spec.Description
	}

	qcb := query.Codebase{
		Name:        cb.Name,
		Description: description,
		CreatedAt:   &cb.ObjectMeta.CreationTimestamp.Time,
		Status:      query.Status(cb.Status.Value),
		Available:   cb.Status.Available,
	}

	if cb.Annotations != nil && cb.Annotations[consts.AdminsAnnotation] != "" {
		admins, err := base64.StdEncoding.DecodeString(cb.Annotations[consts.AdminsAnnotation])
		if err != nil {
			return nil, errors.Wrapf(err, "unable to decode admin from annotation: %s, codebase: %s",
				cb.Annotations[consts.AdminsAnnotation], cb.Name)
		}
		qcb.Admins = string(admins)
	}

	if cb.ObjectMeta.DeletionTimestamp != nil {
		qcb.ForegroundDeletion = true
	}

	if loadBranches {
		var edpCodebaseBranchList edpv1alpha1.CodebaseBranchList
		if err := clientV2.Get().Namespace(s.Namespace).Resource(consts.CodebaseBranchPlural).Do().
			Into(&edpCodebaseBranchList); err != nil {
			return nil, errors.Wrap(err, "unable to get codebase branches from k8s")
		}

		qcb.CodebaseBranch = make([]*query.CodebaseBranch, 0, 2)
		for _, v := range edpCodebaseBranchList.Items {
			if v.Spec.CodebaseName == qcb.Name {
				qcb.CodebaseBranch = append(qcb.CodebaseBranch, &query.CodebaseBranch{
					Name:             v.Spec.BranchName,
					Version:          v.Spec.Version,
					Status:           v.Status.Value,
					Build:            v.Status.Build,
					LastSuccessBuild: v.Status.LastSuccessfulBuild,
				})
			}
		}
	}

	return &qcb, nil
}

func (s *CodebaseService) GetCodebasesByCriteriaK8s(ctx context.Context, criteria query.CodebaseCriteria) ([]*query.Codebase, error) {
	clV2, err := s.Clients.GetEDPRestClientV2(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get edp rest client v2")
	}

	var edpCodebasesList edpv1alpha1.CodebaseList
	if err := clV2.Get().Namespace(s.Namespace).Resource(consts.CodebasePlural).Do().
		Into(&edpCodebasesList); err != nil {
		return nil, errors.Wrap(err, "unable to get codebase list from k8s")
	}

	codebases := make([]*query.Codebase, 0, len(edpCodebasesList.Items))
	for i, v := range edpCodebasesList.Items {
		if v.Spec.Type != string(criteria.Type) {
			continue
		}

		qcb, err := s.k8sCodebase2queryCodebase(clV2, &edpCodebasesList.Items[i], true)
		if err != nil {
			return nil, errors.Wrap(err, "unable to convert k8s codebase to query codebase")
		}
		codebases = append(codebases, qcb)
	}

	return codebases, nil
}

func (s CodebaseService) GetCodebaseByName(name string) (*query.Codebase, error) {
	c, err := s.ICodebaseRepository.GetCodebaseByName(name)
	if err != nil {
		return nil, errors.Wrapf(err, "an error has occurred while getting %v codebase from db", name)
	}
	log.Info("codebase has been fetched from db", zap.String("name", name))
	return c, nil
}

type RegistryNotFound struct {
	cause error
}

func (r RegistryNotFound) Error() string {
	return r.cause.Error()
}

func (s CodebaseService) GetCodebaseByNameK8s(ctx context.Context, name string) (*query.Codebase, error) {
	clientV2, err := s.Clients.GetEDPRestClientV2(ctx)
	if err != nil {
		return nil, err
	}

	var edpCodebase edpv1alpha1.Codebase
	if err := clientV2.Get().Namespace(s.Namespace).Resource(consts.CodebasePlural).Name(name).
		Do().Into(&edpCodebase); err != nil {
		if errStatus, ok := err.(*k8sErrors.StatusError); ok && errStatus.ErrStatus.Code == 404 {
			return nil, RegistryNotFound{cause: errStatus}
		}

		if edpCodebase.ObjectMeta.DeletionTimestamp != nil {
			return nil, RegistryNotFound{}
		}

		return nil, errors.Wrap(err, "unable to get codebase from k8s api")
	}

	qcb, err := s.k8sCodebase2queryCodebase(clientV2, &edpCodebase, true)
	if err != nil {
		return nil, errors.Wrap(err, "unable to convert k8s codebase to query codebase")
	}

	return qcb, nil
}

func (s *CodebaseService) findCodebaseByName(name string) bool {
	exist := s.ICodebaseRepository.FindCodebaseByName(name)
	log.Debug("codebase exists", zap.Bool("exists", exist), zap.String("name", name))
	return exist
}

func (s *CodebaseService) findCodebaseByProjectPath(gitProjectPath *string) bool {
	if gitProjectPath == nil {
		return false
	}
	exist := s.ICodebaseRepository.FindCodebaseByProjectPath(gitProjectPath)
	log.Debug("codebase exists", zap.Bool("exists", exist), zap.String("url", *gitProjectPath))
	return exist
}

func (s CodebaseService) ExistCodebaseAndBranch(cbName, brName string) bool {
	return s.ICodebaseRepository.ExistCodebaseAndBranch(cbName, brName)
}

func createSecret(namespace string, secret *v1.Secret, coreClient k8s.CoreClient) (*v1.Secret, error) { //nolint
	createdSecret, err := coreClient.Secrets(namespace).Create(secret)
	if err != nil {
		log.Error("an error has occurred while saving secret", zap.Error(err))
		return &v1.Secret{}, err
	}
	return createdSecret, nil
}

func (s CodebaseService) createTempSecrets(codebase command.CreateCodebase) error {
	if codebase.Repository != nil && (codebase.Repository.Login == "" && codebase.Repository.Password == "") {
		secret, err := s.Clients.CoreClient.Secrets(s.Namespace).
			Get(s.GerritCreatorSecretName, metav1.GetOptions{})
		if err != nil {
			return errors.Wrap(err, "unable to get gerrit creator secret")
		}

		if username, ok := secret.Data[gerritCreatorUsername]; ok {
			codebase.Repository.Login = string(username)
		}

		if password, ok := secret.Data[gerritCreatorPassword]; ok {
			codebase.Repository.Password = string(password)
		}
	}

	if codebase.Repository != nil && (codebase.Repository.Login != "" && codebase.Repository.Password != "") {
		repoSecretName := fmt.Sprintf("repository-codebase-%s-temp", codebase.Name)
		tempRepoSecret := getSecret(repoSecretName, codebase.Repository.Login, codebase.Repository.Password)

		if _, err := createSecret(s.Namespace, tempRepoSecret, s.Clients.CoreClient); err != nil {
			log.Error("an error has occurred while creating repository secret", zap.Error(err))
			return err
		}
		log.Info("repository secret for codebase was created", zap.String("codebase", codebase.Name))
	}

	if codebase.Vcs != nil {
		vcsSecretName := fmt.Sprintf("vcs-autouser-codebase-%s-temp", codebase.Name)
		tempVcsSecret := getSecret(vcsSecretName, codebase.Vcs.Login, codebase.Vcs.Password)

		if _, err := createSecret(s.Namespace, tempVcsSecret, s.Clients.CoreClient); err != nil {
			log.Error("an error has occurred while creating vcs secret", zap.Error(err))
			return err
		}
		log.Info("VCS secret for codebase was created", zap.String("codebase", codebase.Name))
	}

	return nil
}

func getSecret(name string, username string, password string) *v1.Secret {
	return &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		StringData: map[string]string{
			"username": username,
			"password": password,
		},
	}
}

func convertData(codebase command.CreateCodebase) edpv1alpha1.CodebaseSpec {
	cs := edpv1alpha1.CodebaseSpec{
		Lang:                 codebase.Lang,
		Framework:            codebase.Framework,
		BuildTool:            codebase.BuildTool,
		Strategy:             edpv1alpha1.Strategy(codebase.Strategy),
		Type:                 codebase.Type,
		GitServer:            codebase.GitServer,
		JenkinsSlave:         codebase.JenkinsSlave,
		JobProvisioning:      codebase.JobProvisioning,
		DeploymentScript:     codebase.DeploymentScript,
		JiraServer:           codebase.JiraServer,
		CommitMessagePattern: codebase.CommitMessageRegex,
		TicketNamePattern:    codebase.TicketNameRegex,
		CiTool:               codebase.CiTool,
	}
	if cs.Strategy == "import" {
		cs.GitUrlPath = codebase.GitURLPath
	}
	if codebase.Framework != nil {
		cs.Framework = codebase.Framework
	}
	if codebase.Repository != nil {
		cs.Repository = &edpv1alpha1.Repository{
			Url: codebase.Repository.URL,
		}
	}
	if codebase.Route != nil {
		cs.Route = &edpv1alpha1.Route{
			Site: codebase.Route.Site,
		}
		if len(codebase.Route.Path) > 0 {
			cs.Route.Path = codebase.Route.Path
		}
	}
	if codebase.Database != nil {
		cs.Database = &edpv1alpha1.Database{
			Kind:     codebase.Database.Kind,
			Version:  codebase.Database.Version,
			Capacity: codebase.Database.Capacity,
			Storage:  codebase.Database.Storage,
		}
	}
	if codebase.TestReportFramework != nil {
		cs.TestReportFramework = codebase.TestReportFramework
	}
	if codebase.Description != nil {
		cs.Description = codebase.Description
	}
	cs.Versioning.Type = edpv1alpha1.VersioningType(codebase.Versioning.Type)
	cs.Versioning.StartFrom = codebase.Versioning.StartFrom
	return cs
}

func (s CodebaseService) GetApplicationsToPromote(cdPipelineID int) ([]string, error) {
	appsToPromote, err := s.ICodebaseRepository.SelectApplicationToPromote(cdPipelineID)
	if err != nil {
		return nil, fmt.Errorf("an error has occurred while fetching Ids of applications which shoould be promoted: %v", err)
	}
	return s.selectApplicationNames(appsToPromote)
}

func (s CodebaseService) selectApplicationNames(applicationsToPromote []*query.ApplicationsToPromote) ([]string, error) {
	var result []string
	for _, app := range applicationsToPromote {
		codebase, err := s.ICodebaseRepository.GetCodebaseByID(app.CodebaseID)
		if err != nil {
			return nil, fmt.Errorf("an error has occurred while fetching Codebase by ID %v: %v", app.CodebaseID, err)
		}
		result = append(result, codebase.Name)
	}
	log.Debug("Applications to promote have been fetched", zap.Any("applications", result))
	return result, nil
}

func (s CodebaseService) Delete(name, codebaseType string) error {
	log.Debug("start executing service delete method", zap.String("codebase", name))

	if err := s.deleteCodebase(name); err != nil {
		return errors.Wrap(err, "unable to delete codebase")
	}
	log.Info("end executing service codebase delete method", zap.String("codebase", name))
	return nil
}

func (s CodebaseService) deleteCodebase(name string) error {
	log.Debug("start executing codebase delete request", zap.String("codebase", name))
	r := &edpv1alpha1.Codebase{}
	err := s.Clients.EDPRestClientV2.Delete().
		Namespace(s.Namespace).
		Resource(consts.CodebasePlural).
		Name(name).
		Do().Into(r)
	if err != nil {
		return errors.Wrapf(err, "couldn't delete codebase %v from cluster", name)
	}
	log.Debug("end executing codebase delete request", zap.String("codebase", name))
	return nil
}

func setCodebaseBranchCr(vt string, username string, version *string, defaultBranch string) command.CreateCodebaseBranch {
	if vt == consts.DefaultVersioningType {
		return command.CreateCodebaseBranch{
			Name:     defaultBranch,
			Username: username,
			Build:    &consts.DefaultBuildNumber,
		}
	}

	return command.CreateCodebaseBranch{
		Name:     defaultBranch,
		Username: username,
		Version:  version,
		Build:    &consts.DefaultBuildNumber,
	}
}

func GetCodebaseCR(c rest.Interface, name, namespace string) (*edpv1alpha1.Codebase, error) {
	var cb edpv1alpha1.Codebase
	err := c.Get().Namespace(namespace).Resource(consts.CodebasePlural).Name(name).Do().Into(&cb)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get codebase")
	}

	return &cb, nil
}

func (s *CodebaseService) UpdateDescription(reg *models.Registry) error {
	c, err := GetCodebaseCR(s.Clients.EDPRestClientV2, reg.Name, s.Namespace)
	if err != nil {
		return errors.Wrapf(err, "couldn't get codebase from cluster %v", reg.Name)
	}

	c.Spec.Description = &reg.Description
	if c.Annotations == nil {
		c.Annotations = make(map[string]string)
	}
	c.Annotations[consts.AdminsAnnotation] = base64.StdEncoding.EncodeToString([]byte(reg.Admins))

	if err := s.executeUpdateRequest(c); err != nil {
		return errors.Wrap(err, "unable to update codebase")
	}

	return nil
}

func (s *CodebaseService) Update(command command.UpdateCodebaseCommand) (*edpv1alpha1.Codebase, error) {
	log.Debug("start executing Update method fort codebase", zap.String("name", command.Name))
	c, err := GetCodebaseCR(s.Clients.EDPRestClientV2, command.Name, s.Namespace)
	if err != nil {
		return nil, errors.Wrapf(err, "couldn't get codebase from cluster %v", command.Name)
	}

	c.Spec.CommitMessagePattern = &command.CommitMessageRegex
	c.Spec.TicketNamePattern = &command.TicketNameRegex
	log.Debug("new values",
		zap.String("commitMessagePattern", *c.Spec.CommitMessagePattern),
		zap.String("ticketNamePattern", *c.Spec.TicketNamePattern))

	if err := s.executeUpdateRequest(c); err != nil {
		return nil, err
	}
	log.Info("codebase has been updated", zap.String("name", c.Name))
	return c, nil
}

func (s *CodebaseService) executeUpdateRequest(c *edpv1alpha1.Codebase) error {
	err := s.Clients.EDPRestClientV2.Put().
		Namespace(s.Namespace).
		Resource("codebases").
		Name(c.Name).
		Body(c).
		Do().
		Into(c)
	if err != nil {
		return errors.Wrap(err, "an error has occurred while updating Codebase CR in cluster")
	}
	return nil
}
