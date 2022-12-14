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

package certmanager

import (
	"github.com/flomesh-io/ErieCanal/pkg/certificate"
	"github.com/flomesh-io/ErieCanal/pkg/config"
	cmmeta "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	certmgrclient "github.com/jetstack/cert-manager/pkg/client/clientset/versioned"
	certmgrlister "github.com/jetstack/cert-manager/pkg/client/listers/certmanager/v1"
	"k8s.io/client-go/kubernetes"
	"time"
)

const (
	CAIssuerName          = "flomesh.io"
	SelfSignedIssuerName  = "self-signed.flomesh.io"
	IssuerKind            = "Issuer"
	IssuerGroup           = "cert-manager.io"
	CertManagerRootCAName = "flomesh-root-ca"

	DefaultPollInterval = 1 * time.Second
	DefaultPollTimeout  = 60 * time.Second
)

type Client struct {
	ns                       string // namespace
	mc                       *config.MeshConfig
	cmClient                 certmgrclient.Interface
	kubeClient               kubernetes.Interface
	certificateRequestLister certmgrlister.CertificateRequestNamespaceLister
	certificateLister        certmgrlister.CertificateNamespaceLister
}

type CertManager struct {
	ca           *certificate.Certificate
	client       *Client
	issuerRef    cmmeta.ObjectReference // it's the CA Issuer
	certificates map[string]*certificate.Certificate
}
