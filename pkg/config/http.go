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
	"github.com/flomesh-io/ErieCanal/pkg/repo"
	"github.com/tidwall/sjson"
	"k8s.io/klog/v2"
)

func UpdateIngressHTTPConfig(basepath string, repoClient *repo.PipyRepoClient, mc *MeshConfig) error {
	json, err := getMainJson(basepath, repoClient)
	if err != nil {
		return err
	}

	newJson, err := sjson.Set(json, "http", map[string]interface{}{
		"enabled": mc.Ingress.HTTP.Enabled,
		"listen":  mc.Ingress.HTTP.Listen,
	})
	if err != nil {
		klog.Errorf("Failed to update HTTP config: %s", err)
		return err
	}

	return updateMainJson(basepath, repoClient, newJson)
}
