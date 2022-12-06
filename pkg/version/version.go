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
