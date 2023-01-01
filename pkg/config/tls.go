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
	"github.com/flomesh-io/ErieCanal/pkg/certificate"
	"github.com/flomesh-io/ErieCanal/pkg/commons"
	"github.com/flomesh-io/ErieCanal/pkg/repo"
	"github.com/tidwall/sjson"
	"k8s.io/klog/v2"
)

func UpdateIngressTLSConfig(basepath string, repoClient *repo.PipyRepoClient, mc *MeshConfig) error {
	json, err := getMainJson(basepath, repoClient)
	if err != nil {
		return err
	}

	newJson, err := sjson.Set(json, "tls.enabled", mc.Ingress.TLS.Enabled)
	if err != nil {
		klog.Errorf("Failed to update tls.enabled: %s", err)
		return err
	}
	newJson, err = sjson.Set(newJson, "tls.listen", mc.Ingress.TLS.Listen)
	if err != nil {
		klog.Errorf("Failed to update tls.listen: %s", err)
		return err
	}
	newJson, err = sjson.Set(newJson, "tls.mTLS", mc.Ingress.TLS.MTLS)
	if err != nil {
		klog.Errorf("Failed to update tls.mTLS: %s", err)
		return err
	}

	return updateMainJson(basepath, repoClient, newJson)
}

func IssueCertForIngress(basepath string, repoClient *repo.PipyRepoClient, certMgr certificate.Manager, mc *MeshConfig) error {
	// 1. issue cert
	cert, err := certMgr.IssueCertificate("ingress-pipy", commons.DefaultCAValidityPeriod, []string{})
	if err != nil {
		klog.Errorf("Issue certificate for ingress-pipy error: %s", err)
		return err
	}

	// 2. get main.json
	json, err := getMainJson(basepath, repoClient)
	if err != nil {
		return err
	}

	newJson, err := sjson.Set(json, "tls", map[string]interface{}{
		"enabled": mc.Ingress.TLS.Enabled,
		"listen":  mc.Ingress.TLS.Listen,
		"mTLS":    mc.Ingress.TLS.MTLS,
		"certificate": map[string]interface{}{
			"cert": string(cert.CrtPEM),
			"key":  string(cert.KeyPEM),
			"ca":   string(cert.CA),
		},
	})
	if err != nil {
		klog.Errorf("Failed to update TLS config: %s", err)
		return err
	}

	// 6. update main.json
	return updateMainJson(basepath, repoClient, newJson)
}

func UpdateSSLPassthrough(basepath string, repoClient *repo.PipyRepoClient, enabled bool, upstreamPort int32) error {
	klog.V(5).Infof("SSL passthrough is enabled, updating repo config ...")
	// 1. get main.json
	json, err := getMainJson(basepath, repoClient)
	if err != nil {
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
	return updateMainJson(basepath, repoClient, newJson)
}
