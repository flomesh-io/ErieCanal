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

func NewConfig(k8sApi *kube.K8sAPI, mc *config.MeshConfig) *Config {
	return &Config{
		k8sApi: k8sApi,
		mc:     mc,
		//managerType: managerType,
	}
}

func (c *Config) GetCertificateManager() (certificate.Manager, error) {
	switch certificate.CertificateManagerType(c.mc.Certificate.Manager) {
	case certificate.Manual:
		return c.getManualCertificateManager()
	case certificate.Archon:
		return c.getArchonCertificateManager()
	case certificate.CertManager:
		return c.getCertManagerCertificateManager()
	default:
		return nil, fmt.Errorf("%q is not a valid certificate manager", c.mc.Certificate.Manager)
	}
}

func (c *Config) getArchonCertificateManager() (certificate.Manager, error) {
	rootCert, err := archon.NewRootCA(
		commons.DefaultCACommonName, commons.DefaultCAValidityPeriod,
		commons.DefaultCACountry, commons.DefaultCALocality, commons.DefaultCAOrganization,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate root CA, manager type = %q, %s", c.mc.Certificate.Manager, err.Error())
	}

	rootCert, err = c.getOrSaveCertificate(c.mc.GetCaBundleNamespace(), c.mc.GetCaBundleName(), rootCert)
	if err != nil {
		return nil, fmt.Errorf("failed to get/create CA, %s", err.Error())
	}

	return archon.NewManager(rootCert)
}

func (c *Config) getCertManagerCertificateManager() (certificate.Manager, error) {
	client := certmanager.NewClient(c.k8sApi, c.mc)

	rootCert, err := certmanager.NewRootCA(
		client,
		commons.DefaultCACommonName, commons.DefaultCAValidityPeriod,
		commons.DefaultCACountry, commons.DefaultCALocality, commons.DefaultCAOrganization,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create/get root CA, manager type = %q, %s", c.mc.Certificate.Manager, err.Error())
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
