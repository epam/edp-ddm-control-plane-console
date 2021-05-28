package service

import (
	"ddm-admin-console/k8s"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Jenkins struct {
	clients   *k8s.ClientSet
	namespace string
}

func MakeJenkins(clients *k8s.ClientSet, namespace string) *Jenkins {
	return &Jenkins{
		clients:   clients,
		namespace: namespace,
	}
}

func (j *Jenkins) CreateJobBuildRun(name, jobPath string, jobParams map[string]string) error {
	jbr := k8s.JenkinsJobBuildRun{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: j.namespace,
			Name:      name,
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v2.edp.epam.com/v1alpha1",
			Kind:       "JenkinsJobBuildRun",
		},
		Spec: k8s.JenkinsJobBuildRunSpec{
			Params:  jobParams,
			JobPath: jobPath,
			Retry:   5,
		},
	}

	if err := j.clients.EDPRestClientV2.Post().Namespace(j.namespace).Resource("jenkinsjobbuildruns").
		Body(&jbr).Do().Error(); err != nil {
		return errors.Wrap(err, "unable to create JenkinsJobBuildRun")
	}

	return nil
}
