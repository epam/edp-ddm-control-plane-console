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

package k8s

import (
	edppipelinesv1alpha1 "github.com/epmd-edp/cd-pipeline-operator/v2/pkg/apis/edp/v1alpha1"
	edpv1alpha1 "github.com/epmd-edp/codebase-operator/v2/pkg/apis/edp/v1alpha1"
	"github.com/epmd-edp/edp-component-operator/pkg/apis/v1/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	coreV1Client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	k8sConfig            clientcmd.ClientConfig
	SchemeGroupVersionV2 = schema.GroupVersion{Group: "v2.edp.epam.com", Version: "v1alpha1"}
	SchemeGroupVersionV1 = schema.GroupVersion{Group: "v1.edp.epam.com", Version: "v1alpha1"}
)

type ClientSet struct {
	CoreClient      CoreClient
	EDPRestClientV2 rest.Interface // v2 version
	EDPRestClientV1 rest.Interface // v1 version
}

func init() {
	k8sConfig = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)
}

func CreateOpenShiftClients() ClientSet {
	coreClient, err := getCoreClient()
	if err != nil {
		panic(err)
	}

	crClientV1, crClientV2, err := getApplicationClient()
	if err != nil {
		panic(err)
	}

	return ClientSet{
		CoreClient:      coreClient,
		EDPRestClientV2: crClientV2,
		EDPRestClientV1: crClientV1,
	}
}

func getCoreClient() (*coreV1Client.CoreV1Client, error) {
	restConfig, err := k8sConfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	coreClient, err := coreV1Client.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	return coreClient, nil
}

func getApplicationClient() (v1Client *rest.RESTClient, v2Client *rest.RESTClient, err error) {
	var config *rest.Config
	config, err = k8sConfig.ClientConfig()

	if err != nil {
		return
	}

	clientV1, tErr := createCrdClient(config, &SchemeGroupVersionV1, addKnownTypesV1)
	if tErr != nil {
		return nil, nil, tErr
	}

	clientV2, tErr := createCrdClient(config, &SchemeGroupVersionV2, addKnownTypesV2)
	if tErr != nil {
		return nil, nil, tErr
	}

	return clientV1, clientV2, nil
}

func createCrdClient(cfg *rest.Config, groupVersion *schema.GroupVersion,
	knownTypes func(*runtime.Scheme) error) (*rest.RESTClient, error) {
	scheme := runtime.NewScheme()
	SchemeBuilder := runtime.NewSchemeBuilder(knownTypes)
	if err := SchemeBuilder.AddToScheme(scheme); err != nil {
		return nil, err
	}
	config := *cfg
	config.GroupVersion = groupVersion
	config.APIPath = "/apis"
	config.ContentType = runtime.ContentTypeJSON
	config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: serializer.NewCodecFactory(scheme)}
	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func addKnownTypesV1(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersionV1,
		&v1alpha1.EDPComponent{},
		&v1alpha1.EDPComponentList{},
	)

	metav1.AddToGroupVersion(scheme, SchemeGroupVersionV1)
	return nil
}

func addKnownTypesV2(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersionV2,
		&edpv1alpha1.Codebase{},
		&edpv1alpha1.CodebaseList{},
		&edpv1alpha1.CodebaseBranch{},
		&edpv1alpha1.CodebaseBranchList{},
		&edppipelinesv1alpha1.CDPipeline{},
		&edppipelinesv1alpha1.CDPipelineList{},
		&edppipelinesv1alpha1.Stage{},
		&edppipelinesv1alpha1.StageList{},
	)

	metav1.AddToGroupVersion(scheme, SchemeGroupVersionV2)
	return nil
}
