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
