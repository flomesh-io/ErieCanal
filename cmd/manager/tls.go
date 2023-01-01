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

package main

import (
	"github.com/flomesh-io/ErieCanal/pkg/certificate"
	"github.com/flomesh-io/ErieCanal/pkg/commons"
	"github.com/flomesh-io/ErieCanal/pkg/config"
	"github.com/flomesh-io/ErieCanal/pkg/repo"
	"k8s.io/klog/v2"
	"os"
)

func setupTLS(certMgr certificate.Manager, repoClient *repo.PipyRepoClient, mc *config.MeshConfig) {
	klog.V(5).Infof("mc.Ingress.TLS=%#v", mc.Ingress.TLS)
	if mc.Ingress.TLS.Enabled {
		if mc.Ingress.TLS.SSLPassthrough.Enabled {
			// SSL Passthrough
			if err := config.UpdateSSLPassthrough(
				commons.DefaultIngressBasePath,
				repoClient,
				mc.Ingress.TLS.SSLPassthrough.Enabled,
				mc.Ingress.TLS.SSLPassthrough.UpstreamPort,
			); err != nil {
				os.Exit(1)
			}
		} else {
			// TLS Offload
			if err := config.IssueCertForIngress(commons.DefaultIngressBasePath, repoClient, certMgr, mc); err != nil {
				os.Exit(1)
			}
		}
	}
}
