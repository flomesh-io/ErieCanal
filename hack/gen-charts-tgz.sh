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

DIR=$(cd $(dirname "${BASH_SOURCE}")/.. && pwd -P)
echo "Current DIR is ${DIR}"

CLI_PATH=cli/cmd
ERIE_CANAL_CHART_PATH=charts/erie-canal
NAMESPACED_INGRESS_CHART_PATH=charts/namespaced-ingress
NAMESPACED_INGRESS_CONTROLLER_PATH=controllers/namespacedingress/v1alpha1

########################################################
# package ErieCanal chart
########################################################
${HELM_BIN} dependency update ${ERIE_CANAL_CHART_PATH}/
${HELM_BIN} lint ${ERIE_CANAL_CHART_PATH}/
${HELM_BIN} package ${ERIE_CANAL_CHART_PATH}/ -d ${CLI_PATH}/ --app-version="${PACKAGED_APP_VERSION}" --version=${HELM_CHART_VERSION}
mv ${CLI_PATH}/erie-canal-${HELM_CHART_VERSION}.tgz ${CLI_PATH}/chart.tgz

########################################################
# package namespaced-ingress chart
########################################################
${HELM_BIN} dependency update ${NAMESPACED_INGRESS_CHART_PATH}/
#${HELM_BIN} lint ${NAMESPACED_INGRESS_CHART_PATH}/
${HELM_BIN} package ${NAMESPACED_INGRESS_CHART_PATH}/ -d ${NAMESPACED_INGRESS_CONTROLLER_PATH}/ --app-version="${PACKAGED_APP_VERSION}" --version=${HELM_CHART_VERSION}
mv ${NAMESPACED_INGRESS_CONTROLLER_PATH}/namespaced-ingress-${HELM_CHART_VERSION}.tgz ${NAMESPACED_INGRESS_CONTROLLER_PATH}/chart.tgz