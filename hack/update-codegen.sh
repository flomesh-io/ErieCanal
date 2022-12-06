#!/usr/bin/env bash

#
# MIT License
#
# Copyright (c) since 2021,  flomesh.io Authors.
#
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included in all
# copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
# SOFTWARE.
#

set -o errexit
set -o nounset
set -o pipefail

PROJECT_PKG="github.com/flomesh-io/ErieCanal"
ROOT_DIR="$(git rev-parse --show-toplevel)"

CODEGEN_VERSION="v0.24.3"
go get k8s.io/code-generator@${CODEGEN_VERSION}
CODEGEN_PKG="$(echo `go env GOPATH`/pkg/mod/k8s.io/code-generator@${CODEGEN_VERSION})"

echo ">>> using codegen: ${CODEGEN_PKG}"
chmod +x ${CODEGEN_PKG}/generate-groups.sh

TEMP_DIR=$(mktemp -d)

"${CODEGEN_PKG}"/generate-groups.sh all \
  "${PROJECT_PKG}/pkg/generated" \
  "${PROJECT_PKG}/apis" \
  "cluster:v1alpha1 namespacedingress:v1alpha1 serviceexport:v1alpha1 serviceimport:v1alpha1 multiclusterendpoint:v1alpha1 globaltrafficpolicy:v1alpha1" \
  --go-header-file "${ROOT_DIR}"/hack/boilerplate.go.txt \
  --output-base "${TEMP_DIR}"

cp -rfv "${TEMP_DIR}/${PROJECT_PKG}/." "${ROOT_DIR}/"
rm -rf ${TEMP_DIR}