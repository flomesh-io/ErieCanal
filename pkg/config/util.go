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
	"fmt"
	"github.com/flomesh-io/ErieCanal/pkg/repo"
	"k8s.io/klog/v2"
)

func getMainJson(basepath string, repoClient *repo.PipyRepoClient) (string, error) {
	path := getPathOfMainJson(basepath)

	json, err := repoClient.GetFile(path)
	if err != nil {
		klog.Errorf("Get %q from pipy repo error: %s", path, err)
		return "", err
	}

	return json, nil
}

func updateMainJson(basepath string, repoClient *repo.PipyRepoClient, newJson string) error {
	batch := repo.Batch{
		Basepath: basepath,
		Items: []repo.BatchItem{
			{
				Path:     "/config",
				Filename: "main.json",
				Content:  newJson,
			},
		},
	}

	if err := repoClient.Batch([]repo.Batch{batch}); err != nil {
		klog.Errorf("Failed to update %q: %s", getPathOfMainJson(basepath), err)
		return err
	}

	return nil
}

func getPathOfMainJson(basepath string) string {
	return fmt.Sprintf("%s/config/main.json", basepath)
}
