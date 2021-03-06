package gerrit

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	coreV1Api "k8s.io/api/core/v1"

	"gopkg.in/resty.v1"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	pkgScheme "sigs.k8s.io/controller-runtime/pkg/scheme"

	"ddm-admin-console/service"
)

const (
	//ViewTimeFormat = "02.01.2006 15:04"
	ViewTimeFormat = "2006-01-02 15:04:05"
)

type Service struct {
	Config
	service.UserConfig
	k8sClient  client.Client
	scheme     *runtime.Scheme
	restConfig *rest.Config
	apiClient  *resty.Client
}

type MergeRequest struct {
	Name          string
	ProjectName   string
	TargetBranch  string
	SourceBranch  string
	CommitMessage string
	AuthorName    string
	AuthorEmail   string
	Labels        map[string]string
	Annotations   map[string]string
}

type MRConfigMapFile struct {
	Path     string `json:"path"`
	Contents string `json:"contents"`
}

type Config struct {
	Namespace            string
	RootGerritName       string
	GerritAPIUrlTemplate string
}

func Make(s *runtime.Scheme, k8sConfig *rest.Config, config Config) (*Service, error) {
	builder := pkgScheme.Builder{GroupVersion: schema.GroupVersion{Group: "v2.edp.epam.com", Version: "v1alpha1"}}
	builder.Register(&Gerrit{}, &GerritList{}, &GerritProject{}, &GerritProjectList{}, &GerritMergeRequest{},
		&GerritMergeRequestList{})

	if err := builder.AddToScheme(s); err != nil {
		return nil, errors.Wrap(err, "error during builder add to scheme")
	}

	cl, err := client.New(k8sConfig, client.Options{
		Scheme: s,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to init k8s jenkins client")
	}

	svc := Service{
		k8sClient: cl,
		scheme:    s,
		UserConfig: service.UserConfig{
			RestConfig: k8sConfig,
		},
		restConfig: k8sConfig,
		Config:     config,
	}

	if err := svc.initRestyClient(); err != nil {
		return nil, errors.Wrap(err, "unable to init resty client")
	}

	return &svc, nil
}

func (s *Service) GetProjects(ctx context.Context) ([]GerritProject, error) {
	var projectList GerritProjectList
	if err := s.k8sClient.List(ctx, &projectList); err != nil {
		return nil, errors.Wrap(err, "unable to list gerrit projects")
	}

	return projectList.Items, nil
}

func (s *Service) GetProject(ctx context.Context, name string) (*GerritProject, error) {
	prjs, err := s.GetProjects(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get projects")
	}

	for _, prj := range prjs {
		if prj.Spec.Name == name {
			return &prj, nil
		}
	}

	return nil, service.ErrNotFound("unable to find gerrit project")
}

func (s *Service) CreateProject(ctx context.Context, name string) error {
	if err := s.k8sClient.Create(ctx, &GerritProject{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: s.Namespace,
			Name:      fmt.Sprintf("gerrit-%s", name),
		},
		Spec: GerritProjectSpec{
			Name:              name,
			CreateEmptyCommit: false,
			Parent:            "All-Projects",
		},
	}); err != nil {
		return errors.Wrap(err, "unable to create gerrit project")
	}

	return nil
}

func (s *Service) GetMergeRequest(ctx context.Context, name string) (*GerritMergeRequest, error) {
	var mr GerritMergeRequest
	if err := s.k8sClient.Get(ctx, types.NamespacedName{Name: name, Namespace: s.Namespace}, &mr); err != nil {
		return nil, errors.Wrap(err, "unable to get gerrit merge request")
	}

	return &mr, nil
}

func (s *Service) GetMergeRequestByProject(ctx context.Context, projectName string) ([]GerritMergeRequest, error) {
	var mrs GerritMergeRequestList
	if err := s.k8sClient.List(ctx, &mrs); err != nil {
		return nil, errors.Wrap(err, "unable to list gerrit merge requests")
	}

	var result []GerritMergeRequest
	for _, mr := range mrs.Items {
		if mr.Spec.ProjectName == projectName {
			result = append(result, mr)
		}
	}

	return result, nil
}

func (s *Service) CreateMergeRequest(ctx context.Context, mr *MergeRequest) error {
	if err := s.k8sClient.Create(ctx, &GerritMergeRequest{
		ObjectMeta: metav1.ObjectMeta{Namespace: s.Namespace, Name: mr.Name, Labels: mr.Labels,
			Annotations: mr.Annotations},
		Spec: GerritMergeRequestSpec{
			OwnerName:     s.RootGerritName,
			ProjectName:   mr.ProjectName,
			TargetBranch:  mr.TargetBranch,
			SourceBranch:  mr.SourceBranch,
			CommitMessage: mr.CommitMessage,
			AuthorEmail:   mr.AuthorEmail,
			AuthorName:    mr.AuthorName,
		},
	}); err != nil {
		return errors.Wrap(err, "unable to create merge request")
	}

	return nil
}

func (s *Service) CreateMergeRequestWithContents(ctx context.Context, mr *MergeRequest, contents map[string]string) error {
	changesCMName := fmt.Sprintf("mr-%s-values-%d", mr.ProjectName, time.Now().Unix())

	cmData := make(map[string]string)
	for filePath, content := range contents {
		bts, err := json.Marshal(MRConfigMapFile{Path: filePath, Contents: content})
		if err != nil {
			return errors.Wrap(err, "unable to encode file")
		}

		cmData[filepath.Base(filePath)] = string(bts)
	}

	mergeRequestConfigMap := coreV1Api.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: changesCMName, Namespace: s.Namespace},
		Data:       cmData,
	}

	if err := s.k8sClient.Create(ctx, &mergeRequestConfigMap); err != nil {
		return errors.Wrap(err, "unable to create changes config map")
	}

	if err := s.k8sClient.Create(ctx, &GerritMergeRequest{
		ObjectMeta: metav1.ObjectMeta{Namespace: s.Namespace, Name: mr.Name, Labels: mr.Labels,
			Annotations: mr.Annotations},
		Spec: GerritMergeRequestSpec{
			OwnerName:        s.RootGerritName,
			ProjectName:      mr.ProjectName,
			TargetBranch:     mr.TargetBranch,
			CommitMessage:    mr.CommitMessage,
			AuthorEmail:      mr.AuthorEmail,
			AuthorName:       mr.AuthorName,
			ChangesConfigMap: changesCMName,
			SourceBranch:     "",
		},
	}); err != nil {
		return errors.Wrap(err, "unable to create merge request")
	}

	return nil
}
