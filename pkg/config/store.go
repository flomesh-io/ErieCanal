/*
 * MIT License
 *
 * Copyright (c) since 2021,  flomesh.io Authors.
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
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
