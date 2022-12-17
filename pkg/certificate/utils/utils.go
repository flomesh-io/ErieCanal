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

package utils

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"k8s.io/klog/v2"
)

func ConvertPEMCertToX509(pemCrt []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(pemCrt)
	if block == nil {
		klog.Error("No valid certificate in PEM")
		return nil, fmt.Errorf("no valid certificate in PEM")
	}

	x509Cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		klog.Errorf("failed to convert PEM certificate to x509, %s ", err.Error())
		return nil, err
	}
	return x509Cert, nil
}

func ConvertPEMPrivateKeyToX509(pemKey []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(pemKey)
	if block == nil {
		klog.Error("No valid private key in PEM")
		return nil, fmt.Errorf("no valid private key in PEM")
	}

	x509Key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		klog.Errorf("failed to convert PEM private key to x509, %s ", err.Error())
		return nil, err
	}

	return x509Key.(*rsa.PrivateKey), nil
}

func CertToPEM(caBytes []byte) ([]byte, error) {
	caPEM := new(bytes.Buffer)
	if err := pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	}); err != nil {
		return nil, fmt.Errorf("encode cert: %s", err.Error())
	}

	return caPEM.Bytes(), nil
}

func RSAKeyToPEM(privateKey *rsa.PrivateKey) ([]byte, error) {
	privateBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("marshal private key: %s", err.Error())
	}

	keyPEM := new(bytes.Buffer)
	if err := pem.Encode(keyPEM, &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateBytes,
	}); err != nil {
		return nil, fmt.Errorf("encode key: %s", err.Error())
	}

	return keyPEM.Bytes(), nil
}

func CsrToPEM(csrBytes []byte) ([]byte, error) {
	csrPEM := new(bytes.Buffer)
	if err := pem.Encode(csrPEM, &(pem.Block{
		Type:  "CERTIFICATE REQUEST",
		Bytes: csrBytes,
	})); err != nil {
		return nil, fmt.Errorf("encode CSR: %s", err.Error())
	}

	return csrPEM.Bytes(), nil
}
