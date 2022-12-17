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
	"github.com/flomesh-io/ErieCanal/pkg/commons"
	"github.com/flomesh-io/ErieCanal/pkg/kube"
	"github.com/kelseyhightower/envconfig"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"
)

var (
	meshMetadata             ErieCanalMetadata
	DefaultWatchedConfigMaps = sets.String{}
)

func init() {
	DefaultWatchedConfigMaps.Insert(commons.MeshConfigName)
	meshMetadata = getErieCanalMetadata()
}

type Store struct {
	MeshConfig *MeshConfigClient
}

type ErieCanalMetadata struct {
	PodName            string `envconfig:"POD_NAME" required:"true" split_words:"true"`
	PodNamespace       string `envconfig:"POD_NAMESPACE" required:"true" split_words:"true"`
	ErieCanalNamespace string `envconfig:"NAMESPACE" required:"true" split_words:"true"`
}

func NewStore(k8sApi *kube.K8sAPI) *Store {
	return &Store{
		// create and set default values
		MeshConfig: NewMeshConfigClient(k8sApi),
	}
}

func getErieCanalMetadata() ErieCanalMetadata {
	var metadata ErieCanalMetadata

	err := envconfig.Process("ErieCanal", &metadata)
	if err != nil {
		klog.Error(err, "unable to load ErieCanal metadata from environment")
		panic(err)
	}

	return metadata
}

func GetErieCanalPodName() string {
	return meshMetadata.PodName
}

func GetErieCanalPodNamespace() string {
	return meshMetadata.PodNamespace
}

func GetErieCanalNamespace() string {
	return meshMetadata.ErieCanalNamespace
}
