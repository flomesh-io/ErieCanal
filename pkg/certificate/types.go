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

package certificate

import (
	"math/big"
	"time"
)

var (
	SerialNumberLimit = new(big.Int).Lsh(big.NewInt(1), 128)
)

type Certificate struct {
	CommonName   string
	SerialNumber string
	CA           []byte
	CrtPEM       []byte
	KeyPEM       []byte
	Expiration   time.Time
}

// Manager is for Certificate management.
type Manager interface {
	// IssueCertificate issues a new certificate.
	IssueCertificate(cn string, validityPeriod time.Duration, dnsNames []string) (*Certificate, error)

	// GetCertificate returns a certificate given its Common Name (CN)
	GetCertificate(cn string) (*Certificate, error)

	// GetRootCertificate returns the root certificate in PEM format and its expiration.
	GetRootCertificate() (*Certificate, error)
}

type CertificateManagerType string

const (
	Manual      CertificateManagerType = "manual"
	Archon      CertificateManagerType = "archon"
	CertManager CertificateManagerType = "cert-manager"
)
