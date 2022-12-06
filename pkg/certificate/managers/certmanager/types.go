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

package certmanager

import (
	"github.com/flomesh-io/ErieCanal/pkg/certificate"
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
	cmClient                 certmgrclient.Interface
	kubeClient               kubernetes.Interface
	certificateRequestLister certmgrlister.CertificateRequestNamespaceLister
	certificateLister        certmgrlister.CertificateNamespaceLister
}

type CertManager struct {
	ca        *certificate.Certificate
	client    *Client
	issuerRef cmmeta.ObjectReference // it's the CA Issuer
}
