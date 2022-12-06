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
