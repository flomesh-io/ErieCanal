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
	"context"
	"fmt"
	"github.com/flomesh-io/ErieCanal/pkg/certificate"
	"github.com/flomesh-io/ErieCanal/pkg/certificate/managers/archon"
	"github.com/flomesh-io/ErieCanal/pkg/certificate/managers/certmanager"
	"github.com/flomesh-io/ErieCanal/pkg/certificate/utils"
	"github.com/flomesh-io/ErieCanal/pkg/commons"
	"github.com/flomesh-io/ErieCanal/pkg/config"
	"github.com/flomesh-io/ErieCanal/pkg/kube"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

func NewConfig(k8sApi *kube.K8sAPI, managerType certificate.CertificateManagerType) *Config {
	return &Config{
		k8sApi:      k8sApi,
		managerType: managerType,
	}
}

func (c *Config) GetCertificateManager() (certificate.Manager, error) {
	switch c.managerType {
	case certificate.Manual:
		return c.getManualCertificateManager()
	case certificate.Archon:
		return c.getArchonCertificateManager()
	case certificate.CertManager:
		return c.getCertManagerCertificateManager()
	default:
		return nil, fmt.Errorf("%q is not a valid certificate manager", c.managerType)
	}
}

func (c *Config) getArchonCertificateManager() (certificate.Manager, error) {
	rootCert, err := archon.NewRootCA(
		commons.DefaultCACommonName, commons.DefaultCAValidityPeriod,
		commons.DefaultCACountry, commons.DefaultCALocality, commons.DefaultCAOrganization,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate root CA, manager type = %q, %s", c.managerType, err.Error())
	}

	rootCert, err = c.getOrSaveCertificate(config.GetErieCanalNamespace(), commons.DefaultCABundleName, rootCert)
	if err != nil {
		return nil, fmt.Errorf("failed to get/create CA, %s", err.Error())
	}

	return archon.NewManager(rootCert)
}

func (c *Config) getCertManagerCertificateManager() (certificate.Manager, error) {
	client := certmanager.NewClient(c.k8sApi)

	rootCert, err := certmanager.NewRootCA(
		client,
		commons.DefaultCACommonName, commons.DefaultCAValidityPeriod,
		commons.DefaultCACountry, commons.DefaultCALocality, commons.DefaultCAOrganization,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create/get root CA, manager type = %q, %s", c.managerType, err.Error())
	}

	return certmanager.NewManager(rootCert, client)
}

func (c *Config) getManualCertificateManager() (certificate.Manager, error) {
	panic("Not implemented yet.")
}

func (c *Config) getOrSaveCertificate(ns string, secretName string, cert *certificate.Certificate) (*certificate.Certificate, error) {
	secret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{Kind: "Secret", APIVersion: "v1"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: ns,
		},
		Data: map[string][]byte{
			commons.RootCACertName:       cert.CrtPEM,
			commons.RootCAPrivateKeyName: cert.KeyPEM,
		},
	}

	if _, err := c.k8sApi.Client.CoreV1().Secrets(ns).Create(context.TODO(), secret, metav1.CreateOptions{}); err == nil {
		klog.V(2).Infof("Secret %s/%s created successfully", ns, secretName)
	} else if apierrors.IsAlreadyExists(err) {
		// it's normal in case of race condition
		klog.V(2).Infof("Secret %s/%s already exists.", ns, secretName)
	} else {
		klog.Errorf("Error creating certificate secret %s/%s", ns, secretName)
		return nil, err
	}

	// get it from kubernetes
	secret, err := c.k8sApi.Client.CoreV1().Secrets(ns).Get(context.Background(), secretName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("CA bundle not found: %s", err.Error())
	}

	pemCACrt, ok := secret.Data[commons.RootCACertName]
	if !ok {
		klog.Errorf("Secret %s/%s doesn't have required %q data", ns, secretName, commons.RootCACertName)
		return nil, fmt.Errorf("invalid secret data for cert")
	}

	pemCAKey, ok := secret.Data[commons.RootCAPrivateKeyName]
	if !ok {
		klog.Errorf("Secret %s/%s doesn't have required %q data", ns, secretName, commons.RootCAPrivateKeyName)
		return nil, fmt.Errorf("invalid secret data for cert")
	}

	x509Cert, err := utils.ConvertPEMCertToX509(pemCACrt)
	if err != nil {
		return nil, err
	}

	return &certificate.Certificate{
		CommonName:   x509Cert.Subject.CommonName,
		SerialNumber: x509Cert.SerialNumber.String(),
		CA:           pemCACrt,
		CrtPEM:       pemCACrt,
		KeyPEM:       pemCAKey,
		Expiration:   x509Cert.NotAfter,
	}, nil
}
