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

package version

import (
	"github.com/blang/semver"
	"github.com/flomesh-io/ErieCanal/pkg/kube"
	"k8s.io/klog/v2"
	"os"
	"strings"
)

// var needs to be used instead of const for ldflags
var (
	Version           = "unknown"
	GitVersion        = "unknown"
	GitCommit         = "unknown"
	KubernetesVersion = "unknown"
	ImageVersion      = "unknown"
	BuildDate         = "unknown"

	ServerVersion = semver.Version{Major: 0, Minor: 0, Patch: 0}
	//PIPY Operator requires k8s 1.19+
	MinK8sVersion           = semver.Version{Major: 1, Minor: 19, Patch: 0}
	MinEndpointSliceVersion = semver.Version{Major: 1, Minor: 21, Patch: 0}
)

func getServerVersion(k8sApi *kube.K8sAPI) (semver.Version, error) {
	serverVersion, err := k8sApi.DiscoveryClient.ServerVersion()
	if err != nil {
		klog.Error(err, "unable to get Server Version")
		return semver.Version{Major: 0, Minor: 0, Patch: 0}, err
	}

	gitVersion := serverVersion.GitVersion
	if len(gitVersion) > 1 && strings.HasPrefix(gitVersion, "v") {
		gitVersion = gitVersion[1:]
	}

	return semver.MustParse(gitVersion), nil
}

func detectServerVersion(api *kube.K8sAPI) {
	if ServerVersion.EQ(semver.Version{Major: 0, Minor: 0, Patch: 0}) {
		ver, err := getServerVersion(api)
		if err != nil {
			klog.Error(err, "unable to get server version")
			os.Exit(1)
		}

		ServerVersion = ver
	}
}

func IsSupportedK8sVersion(api *kube.K8sAPI) bool {
	detectServerVersion(api)
	return ServerVersion.GTE(MinK8sVersion)
}

func IsEndpointSliceEnabled(api *kube.K8sAPI) bool {
	detectServerVersion(api)
	return ServerVersion.GTE(MinEndpointSliceVersion)
}
