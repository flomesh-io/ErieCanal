#!/usr/bin/env bash

#
# Copyright 2022 The flomesh.io Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

set -o errexit
set -o nounset
set -o pipefail

PROJECT_PKG="github.com/flomesh-io/ErieCanal"
ROOT_DIR="$(git rev-parse --show-toplevel)"

CODEGEN_VERSION="v0.25.5"
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