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
	"context"

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

const (
	UserTokenKey = "access-token"
)

type ClientSet struct {
	CoreClient           CoreClient
	EDPRestClientV2      rest.Interface // v2 version
	EDPRestClientV1      rest.Interface // v1 version
	restConfig           *rest.Config
	schemeGroupVersionV1 *schema.GroupVersion
	schemeGroupVersionV2 *schema.GroupVersion
}

func (cs *ClientSet) GetConfig() *rest.Config {
	return cs.restConfig
}

func MakeK8SClients() (*ClientSet, error) {
	k8sConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)

	restConfig, err := k8sConfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	cs := ClientSet{restConfig: restConfig,
		schemeGroupVersionV1: &schema.GroupVersion{Group: "v1.edp.epam.com", Version: "v1alpha1"},
		schemeGroupVersionV2: &schema.GroupVersion{Group: "v2.edp.epam.com", Version: "v1alpha1"},
	}

	cs.CoreClient, err = coreV1Client.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	cs.EDPRestClientV1, err = createCrdClient(cs.restConfig, cs.schemeGroupVersionV1, cs.knownTypesV1)
	if err != nil {
		return nil, err
	}

	cs.EDPRestClientV2, err = createCrdClient(cs.restConfig, cs.schemeGroupVersionV2, cs.knownTypesV2)
	if err != nil {
		return nil, err
	}

	return &cs, nil
}

func createCrdClient(conf *rest.Config, groupVersion *schema.GroupVersion,
	knownTypes func(*runtime.Scheme) error) (*rest.RESTClient, error) {

	scheme := runtime.NewScheme()
	SchemeBuilder := runtime.NewSchemeBuilder(knownTypes)
	if err := SchemeBuilder.AddToScheme(scheme); err != nil {
		return nil, err
	}

	config := *conf
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

func (cs *ClientSet) knownTypesV1(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(*cs.schemeGroupVersionV1,
		&v1alpha1.EDPComponent{},
		&v1alpha1.EDPComponentList{},
	)

	metav1.AddToGroupVersion(scheme, *cs.schemeGroupVersionV1)
	return nil
}

func (cs *ClientSet) knownTypesV2(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(*cs.schemeGroupVersionV2,
		&edpv1alpha1.Codebase{},
		&edpv1alpha1.CodebaseList{},
		&edpv1alpha1.CodebaseBranch{},
		&edpv1alpha1.CodebaseBranchList{},
		&edppipelinesv1alpha1.CDPipeline{},
		&edppipelinesv1alpha1.CDPipelineList{},
		&edppipelinesv1alpha1.Stage{},
		&edppipelinesv1alpha1.StageList{},
	)

	metav1.AddToGroupVersion(scheme, *cs.schemeGroupVersionV2)
	return nil
}

func (cs *ClientSet) GetCoreClient(ctx context.Context) (CoreClient, error) {
	userConfig, changed := cs.userConfig(ctx)
	if !changed {
		return cs.CoreClient, nil
	}

	userCoreClient, err := coreV1Client.NewForConfig(userConfig)
	if err != nil {
		return nil, err
	}

	return userCoreClient, nil
}

func (cs *ClientSet) userConfig(ctx context.Context) (config *rest.Config, changed bool) {
	tok := ctx.Value(UserTokenKey)
	if tok == nil {
		return cs.restConfig, false
	}

	tokString, ok := tok.(string)
	if !ok {
		return cs.restConfig, false
	}

	userConfig := rest.AnonymousClientConfig(cs.restConfig)
	userConfig.BearerToken = tokString

	return userConfig, true
}

func (cs *ClientSet) GetEDPRestClientV1(ctx context.Context) (rest.Interface, error) {
	userConfig, changed := cs.userConfig(ctx)
	if !changed {
		return cs.EDPRestClientV1, nil
	}

	edpRestClientV1, err := createCrdClient(userConfig, cs.schemeGroupVersionV1, cs.knownTypesV1)
	if err != nil {
		return nil, err
	}

	return edpRestClientV1, nil
}

func (cs *ClientSet) GetEDPRestClientV2(ctx context.Context) (rest.Interface, error) {
	userConfig, changed := cs.userConfig(ctx)
	if !changed {
		return cs.EDPRestClientV2, nil
	}

	edpRestClientV2, err := createCrdClient(userConfig, cs.schemeGroupVersionV2, cs.knownTypesV2)
	if err != nil {
		return nil, err
	}

	return edpRestClientV2, nil
}
