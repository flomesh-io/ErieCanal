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

package archon

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"github.com/flomesh-io/ErieCanal/pkg/certificate"
	"github.com/flomesh-io/ErieCanal/pkg/certificate/utils"
	"github.com/flomesh-io/ErieCanal/pkg/commons"
	"time"
)

func NewRootCA(cn string, validityPeriod time.Duration,
	country, locality, organization string) (*certificate.Certificate, error) {
	serialNumber, err := rand.Int(rand.Reader, certificate.SerialNumberLimit)
	if err != nil {
		return nil, fmt.Errorf("generate serial number: %s", err.Error())
	}

	now := time.Now()
	ca := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   cn,
			Country:      []string{country},
			Locality:     []string{locality},
			Organization: []string{organization},
		},
		NotBefore:             now,
		NotAfter:              now.Add(validityPeriod),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
	}

	// CA private key
	caPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	// Self-signed CA certificate
	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &caPrivateKey.PublicKey, caPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("create cert: %s", err.Error())
	}

	// PEM encode CA cert
	pemCACrt, err := utils.CertToPEM(caBytes)
	if err != nil {
		return nil, err
	}

	// PEM encode CA private key
	pemCAKey, err := utils.RSAKeyToPEM(caPrivateKey)
	if err != nil {
		return nil, err
	}

	return &certificate.Certificate{
		CommonName:   cn,
		SerialNumber: serialNumber.String(),
		CA:           pemCACrt,
		CrtPEM:       pemCACrt,
		KeyPEM:       pemCAKey,
		Expiration:   ca.NotAfter,
	}, nil
}

func NewManager(ca *certificate.Certificate) (*ArchonManager, error) {
	return &ArchonManager{
		ca: ca,
	}, nil
}

func (m *ArchonManager) IssueCertificate(cn string, validityPeriod time.Duration, dnsNames []string) (*certificate.Certificate, error) {
	serialNumber, err := rand.Int(rand.Reader, certificate.SerialNumberLimit)
	if err != nil {
		return nil, fmt.Errorf("generate serial number: %s", err.Error())
	}

	now := time.Now()
	cert := x509.Certificate{
		SerialNumber: serialNumber,
		DNSNames:     dnsNames,
		Subject: pkix.Name{
			CommonName:   cn,
			Organization: []string{commons.DefaultCAOrganization},
		},
		NotBefore:             now,
		NotAfter:              now.Add(validityPeriod),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// ROOT CA
	rootCert, err := utils.ConvertPEMCertToX509(m.ca.CrtPEM)
	if err != nil {
		return nil, err
	}

	// ROOT private key
	rootKey, err := utils.ConvertPEMPrivateKeyToX509(m.ca.KeyPEM)
	if err != nil {
		return nil, err
	}

	// TLS private key
	tlsKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("generate private key: %s", err.Error())
	}

	// sign the certificate
	tlsBytes, err := x509.CreateCertificate(rand.Reader, &cert, rootCert, &tlsKey.PublicKey, rootKey)
	if err != nil {
		return nil, fmt.Errorf("create cert: %s", err.Error())
	}

	// PEM encode cert
	pemTlsCert, err := utils.CertToPEM(tlsBytes)
	if err != nil {
		return nil, err
	}

	// PEM encode key
	pemTlsKey, err := utils.RSAKeyToPEM(tlsKey)
	if err != nil {
		return nil, err
	}

	return &certificate.Certificate{
		CommonName:   cn,
		SerialNumber: serialNumber.String(),
		CA:           m.ca.CrtPEM,
		CrtPEM:       pemTlsCert,
		KeyPEM:       pemTlsKey,
		Expiration:   cert.NotAfter,
	}, nil
}

func (m *ArchonManager) GetCertificate(cn string) (*certificate.Certificate, error) {
	panic("Not implemented yet")
}

func (m *ArchonManager) GetRootCertificate() (*certificate.Certificate, error) {
	return m.ca, nil
}
