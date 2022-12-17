/*
 * Copyright 2022 The flomesh.io Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package config

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/flomesh-io/ErieCanal/pkg/commons"
	"github.com/flomesh-io/ErieCanal/pkg/kube"
	"github.com/go-playground/validator/v10"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	v1 "k8s.io/client-go/listers/core/v1"
	k8scache "k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"time"
)

var (
	validate = validator.New()
)

type MeshConfig struct {
	IsManaged   bool        `json:"isManaged"`
	Repo        Repo        `json:"repo"`
	Images      Images      `json:"images"`
	Webhook     Webhook     `json:"webhook"`
	Ingress     Ingress     `json:"ingress"`
	GatewayApi  GatewayApi  `json:"gatewayApi"`
	Certificate Certificate `json:"certificate"`
	Cluster     Cluster     `json:"cluster"`
	ServiceLB   ServiceLB   `json:"serviceLB"`
}

type Repo struct {
	RootURL string `json:"rootURL" validate:"required,url"`
}

type Images struct {
	Repository     string `json:"repository" validate:"required"`
	PipyImage      string `json:"pipyImage" validate:"required"`
	ProxyInitImage string `json:"proxyInitImage" validate:"required"`
	KlipperLbImage string `json:"klipperLbImage" validate:"required"`
}

type Webhook struct {
	ServiceName string `json:"serviceName" validate:"required,hostname"`
}

type Ingress struct {
	Enabled    bool `json:"enabled"`
	Namespaced bool `json:"namespaced"`
	TLS        TLS  `json:"tls,omitempty"`
}

type TLS struct {
	Enabled        bool           `json:"enabled"`
	SSLPassthrough SSLPassthrough `json:"sslPassthrough,omitempty"`
}

type SSLPassthrough struct {
	Enabled      bool  `json:"enabled"`
	UpstreamPort int32 `json:"upstreamPort" validate:"gte=1,lte=65535"`
}

type GatewayApi struct {
	Enabled bool `json:"enabled"`
}

type Cluster struct {
	UID             string `json:"uid"`
	Region          string `json:"region"`
	Zone            string `json:"zone"`
	Group           string `json:"group"`
	Name            string `json:"name" validate:"required"`
	ControlPlaneUID string `json:"controlPlaneUID"`
}

type ServiceLB struct {
	Enabled bool `json:"enabled"`
}

type Certificate struct {
	Manager string `json:"manager,omitempty"`
}

type MeshConfigClient struct {
	k8sApi   *kube.K8sAPI
	cmLister v1.ConfigMapNamespaceLister
}

func NewMeshConfigClient(k8sApi *kube.K8sAPI) *MeshConfigClient {
	informerFactory := informers.NewSharedInformerFactoryWithOptions(k8sApi.Client, 60*time.Second, informers.WithNamespace(GetErieCanalNamespace()))
	configmapLister := informerFactory.Core().V1().ConfigMaps().Lister().ConfigMaps(GetErieCanalNamespace())
	configmapInformer := informerFactory.Core().V1().ConfigMaps().Informer()
	go configmapInformer.Run(wait.NeverStop)

	if !k8scache.WaitForCacheSync(wait.NeverStop, configmapInformer.HasSynced) {
		runtime.HandleError(fmt.Errorf("timed out waiting for configmap to sync"))
	}

	return &MeshConfigClient{
		k8sApi:   k8sApi,
		cmLister: configmapLister,
	}
}

func (o *MeshConfig) IsControlPlane() bool {
	return o.Cluster.ControlPlaneUID == "" ||
		o.Cluster.UID == o.Cluster.ControlPlaneUID
}

func (o *MeshConfig) PipyImage() string {
	return fmt.Sprintf("%s/%s", o.Images.Repository, o.Images.PipyImage)
}

func (o *MeshConfig) ProxyInitImage() string {
	return fmt.Sprintf("%s/%s", o.Images.Repository, o.Images.ProxyInitImage)
}

func (o *MeshConfig) ServiceLbImage() string {
	return fmt.Sprintf("%s/%s", o.Images.Repository, o.Images.KlipperLbImage)
}

func (o *MeshConfig) RepoRootURL() string {
	return o.Repo.RootURL
}

func (o *MeshConfig) RepoBaseURL() string {
	return fmt.Sprintf("%s%s", o.Repo.RootURL, commons.DefaultPipyRepoPath)
}

func (o *MeshConfig) IngressCodebasePath() string {
	// Format:
	//  /{{ .Region }}/{{ .Zone }}/{{ .Group }}/{{ .Cluster }}/ingress

	return o.GetDefaultIngressPath()
}

func (o *MeshConfig) NamespacedIngressCodebasePath(namespace string) string {
	// Format:
	//  /{{ .Region }}/{{ .Zone }}/{{ .Group }}/{{ .Cluster }}/nsig/{{ .Namespace }}

	//return util.EvaluateTemplate(commons.NamespacedIngressPathTemplate, struct {
	//	Region    string
	//	Zone      string
	//	Group     string
	//	Cluster   string
	//	Namespace string
	//}{
	//	Region:    o.Cluster.Region,
	//	Zone:      o.Cluster.Zone,
	//	Group:     o.Cluster.Group,
	//	Cluster:   o.Cluster.Name,
	//	Namespace: namespace,
	//})

	return fmt.Sprintf("/local/nsig/%s", namespace)
}

func (o *MeshConfig) GetDefaultServicesPath() string {
	// Format:
	//  /{{ .Region }}/{{ .Zone }}/{{ .Group }}/{{ .Cluster }}/services

	//return util.EvaluateTemplate(commons.ServicePathTemplate, struct {
	//	Region  string
	//	Zone    string
	//	Group   string
	//	Cluster string
	//}{
	//	Region:  o.Cluster.Region,
	//	Zone:    o.Cluster.Zone,
	//	Group:   o.Cluster.Group,
	//	Cluster: o.Cluster.Name,
	//})

	return "/local/services"
}

func (o *MeshConfig) GetDefaultIngressPath() string {
	// Format:
	//  /{{ .Region }}/{{ .Zone }}/{{ .Group }}/{{ .Cluster }}/ingress

	//return util.EvaluateTemplate(commons.IngressPathTemplate, struct {
	//	Region  string
	//	Zone    string
	//	Group   string
	//	Cluster string
	//}{
	//	Region:  o.Cluster.Region,
	//	Zone:    o.Cluster.Zone,
	//	Group:   o.Cluster.Group,
	//	Cluster: o.Cluster.Name,
	//})

	return "/local/ingress"
}

func (o *MeshConfig) ToJson() string {
	cfgBytes, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		klog.Errorf("Not able to marshal MeshConfig %#v to json, %s", o, err.Error())
		return ""
	}

	return string(cfgBytes)
}

func (c *MeshConfigClient) GetConfig() *MeshConfig {
	cm := c.getConfigMap()

	if cm != nil {
		cfg, err := ParseMeshConfig(cm)
		if err != nil {
			panic(err)
		}

		return cfg
	}

	//return nil
	panic("MeshConfig is not found or has invalid value")
}

func (c *MeshConfigClient) UpdateConfig(config *MeshConfig) (*MeshConfig, error) {
	if config == nil {
		klog.Errorf("config is nil")
		return nil, fmt.Errorf("config is nil")
	}

	err := validate.Struct(config)
	if err != nil {
		klog.Errorf("Validation error: %#v, rejecting the new config...", err)
		return nil, err
	}

	cm := c.getConfigMap()
	if cm == nil {
		return nil, fmt.Errorf("config map '%s/erie-canal-mesh-config' is not found", GetErieCanalNamespace())
	}
	cm.Data[commons.MeshConfigJsonName] = config.ToJson()

	cm, err = c.k8sApi.Client.CoreV1().
		ConfigMaps(GetErieCanalNamespace()).
		Update(context.TODO(), cm, metav1.UpdateOptions{})

	if err != nil {
		msg := fmt.Sprintf("Update ConfigMap %s/erie-canal-mesh-config error, %s", GetErieCanalNamespace(), err)
		klog.Errorf(msg)
		return nil, fmt.Errorf(msg)
	}

	klog.V(5).Infof("After updating, ConfigMap %s/erie-canal-mesh-config = %#v", GetErieCanalNamespace(), cm)

	return ParseMeshConfig(cm)
}

func (c *MeshConfigClient) getConfigMap() *corev1.ConfigMap {
	cm, err := c.cmLister.Get(commons.MeshConfigName)

	if err != nil {
		// it takes time to sync, perhaps still not in the local store yet
		if apierrors.IsNotFound(err) {
			cm, err = c.k8sApi.Client.CoreV1().
				ConfigMaps(GetErieCanalNamespace()).
				Get(context.TODO(), commons.MeshConfigName, metav1.GetOptions{})

			if err != nil {
				klog.Errorf("Get ConfigMap %s/erie-canal-mesh-config from API server error, %s", GetErieCanalNamespace(), err.Error())
				return nil
			}
		} else {
			klog.Errorf("Get ConfigMap %s/erie-canal-mesh-config error, %s", GetErieCanalNamespace(), err.Error())
			return nil
		}
	}

	return cm
}

func ParseMeshConfig(cm *corev1.ConfigMap) (*MeshConfig, error) {
	cfgJson, ok := cm.Data[commons.MeshConfigJsonName]
	if !ok {
		msg := fmt.Sprintf("Config file mesh_config.json not found, please check ConfigMap %s/erie-canal-mesh-config.", GetErieCanalNamespace())
		klog.Errorf(msg)
		return nil, fmt.Errorf(msg)
	}
	klog.V(5).Infof("Found mesh_config.json, content: %s", cfgJson)

	cfg := MeshConfig{}
	err := json.Unmarshal([]byte(cfgJson), &cfg)
	if err != nil {
		msg := fmt.Sprintf("Unable to unmarshal mesh_config.json to config.MeshConfig, %s", err)
		klog.Errorf(msg)
		return nil, fmt.Errorf(msg)
	}

	err = validate.Struct(cfg)
	if err != nil {
		klog.Errorf("Validation error: %#v", err)
		// in case of validation error, the app doesn't run properly with wrong config, should panic
		//panic(err)
		return nil, err
	}

	return &cfg, nil
}
