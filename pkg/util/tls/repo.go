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
