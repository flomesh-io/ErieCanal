#!/bin/bash

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

if [ -n "$DEBUG" ]; then
	set -x
fi

set -o errexit
set -o nounset
set -o pipefail

K8S_DEFAULT_VERSION=1.19

DIR=$(cd $(dirname "${BASH_SOURCE}")/.. && pwd -P)
echo "Current DIR is ${DIR}"

# clean
rm -rf ${DIR}/manifests/*

TEMPLATE_DIR="${DIR}/hack/manifests"
echo "TEMPLATE_DIR is ${TEMPLATE_DIR}"

TARGETS=$(dirname $(cd $DIR/hack/manifests/ && find . -type f -name "values.yaml" ) | cut -d'/' -f2-)
K8S_VERSION=${K8S_DEFAULT_VERSION}

for TARGET in ${TARGETS}
do
  echo "TARGET is ${TARGET}"
  TARGET_DIR="${TEMPLATE_DIR}/${TARGET}"
  echo "TARGET_DIR is ${TARGET_DIR}"
  MANIFEST="${TEMPLATE_DIR}/manifest.yaml" # intermediate manifest
  OUTPUT_DIR="${DIR}/deploy/${TARGET}"
  echo "OUTPUT_DIR is ${OUTPUT_DIR}"

  mkdir -p ${OUTPUT_DIR}
  cd ${TARGET_DIR}
  helm template erie-canal ${DIR}/charts/erie-canal \
    --values values.yaml \
    --namespace erie-canal \
    --no-hooks \
    --kube-version ${K8S_VERSION} \
    --set ec.version=${ERIE_CANAL_IMAGE_TAG:-latest} \
    --set ec.logLevel=${ERIE_CANAL_LOG_LEVEL:-2} \
    --set ec.image.pullPolicy=${ERIE_CANAL_IMAGE_PULL_POLICY:-IfNotPresent} \
    > $MANIFEST
  kustomize --load-restrictor=LoadRestrictionsNone build . > ${OUTPUT_DIR}/${ERIE_CANAL_DEPLOY_YAML}
  rm $MANIFEST
  cd ~-
done

