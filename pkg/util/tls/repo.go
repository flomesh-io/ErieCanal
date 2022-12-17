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

package tls

import (
	"fmt"
	"github.com/flomesh-io/ErieCanal/pkg/certificate"
	"github.com/flomesh-io/ErieCanal/pkg/commons"
	"github.com/flomesh-io/ErieCanal/pkg/config"
	"github.com/flomesh-io/ErieCanal/pkg/repo"
	"github.com/tidwall/sjson"
	"k8s.io/klog/v2"
)

func IssueCertForIngress(basepath string, repoClient *repo.PipyRepoClient, certMgr certificate.Manager, mc *config.MeshConfig) error {
	// 1. issue cert
	cert, err := certMgr.IssueCertificate("ingress-pipy", commons.DefaultCAValidityPeriod, []string{})
	if err != nil {
		klog.Errorf("Issue certificate for ingress-pipy error: %s", err)
		return err
	}

	// 2. get main.json
	path := fmt.Sprintf("%s/config/main.json", basepath)
	json, err := repoClient.GetFile(path)
	if err != nil {
		klog.Errorf("Get %q from pipy repo error: %s", path, err)
		return err
	}

	// 3. update CertificateChain
	newJson, err := sjson.Set(json, "certificates.cert", string(cert.CrtPEM))
	if err != nil {
		klog.Errorf("Failed to update certificates.cert: %s", err)
		return err
	}
	// 4. update Private Key
	newJson, err = sjson.Set(newJson, "certificates.key", string(cert.KeyPEM))
	if err != nil {
		klog.Errorf("Failed to update certificates.key: %s", err)
		return err
	}

	// 5. update CA
	//newJson, err = sjson.Set(newJson, "certificates.ca", string(cert.CA))
	//if err != nil {
	//	klog.Errorf("Failed to update certificates.key: %s", err)
	//	return err
	//}

	// 6. update main.json
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
		klog.Errorf("Failed to update %q: %s", path, err)
		return err
	}

	return nil
}

func UpdateSSLPassthrough(basepath string, repoClient *repo.PipyRepoClient, enabled bool, upstreamPort int32) error {
	klog.V(5).Infof("SSL passthrough is enabled, updating repo config ...")
	// 1. get main.json
	path := fmt.Sprintf("%s/config/main.json", basepath)
	json, err := repoClient.GetFile(path)
	if err != nil {
		klog.Errorf("Get %q from pipy repo error: %s", path, err)
		return err
	}

	// 2. update ssl passthrough config
	klog.V(5).Infof("SSLPassthrough enabled=%t", enabled)
	klog.V(5).Infof("SSLPassthrough upstreamPort=%d", upstreamPort)
	newJson, err := sjson.Set(json, "sslPassthrough", map[string]interface{}{
		"enabled":      enabled,
		"upstreamPort": upstreamPort,
	})
	if err != nil {
		klog.Errorf("Failed to update sslPassthrough: %s", err)
		return err
	}

	// 3. update main.json
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
		klog.Errorf("Failed to update %q: %s", path, err)
		return err
	}

	return nil
}
