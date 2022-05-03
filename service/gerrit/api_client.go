package gerrit

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/resty.v1"
	coreV1Api "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (s *Service) initRestyClient() error {
	var secret coreV1Api.Secret
	if err := s.k8sClient.Get(context.Background(),
		types.NamespacedName{Namespace: s.Namespace,
			Name: fmt.Sprintf("%s-admin-password", s.RootGerritName)}, &secret); err != nil {
		return errors.Wrap(err, "unable to get admin secret")
	}

	s.apiClient = resty.New().
		SetHostURL(strings.ReplaceAll(s.GerritAPIUrlTemplate, "{HOST}",
			fmt.Sprintf("%s.%s", s.RootGerritName, s.Namespace))).
		SetBasicAuth("admin", string(secret.Data["password"])).
		SetDisableWarn(true)

	return nil
}

func (s *Service) GetFileContents(ctx context.Context, projectName, branch, filePath string) (string, error) {
	rsp, err := s.apiClient.NewRequest().SetContext(ctx).
		Get(fmt.Sprintf("projects/%s/branches/%s/files/%s/content", projectName, branch, filePath))
	if err != nil {
		return "", errors.Wrap(err, "unable to get file content")
	}

	if rsp.StatusCode() >= 300 {
		return "", errors.Wrapf(err, "wrong response code: %d, content: %s", rsp.StatusCode(), rsp.String())
	}

	bts, err := base64.StdEncoding.DecodeString(rsp.String())
	if err != nil {
		return "", errors.Wrap(err, "unable to decode response")
	}

	return string(bts), nil
}
