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

package ingress

import (
	"github.com/blang/semver"
	"github.com/flomesh-io/ErieCanal/pkg/commons"
)

var (
	MinK8sVersionForIngressV1           = semver.Version{Major: 1, Minor: 19, Patch: 0}
	MinK8sVersionForIngressV1beta1      = semver.Version{Major: 1, Minor: 16, Patch: 0}
	MinK8sVersionForIngressClassV1beta1 = semver.Version{Major: 1, Minor: 18, Patch: 0}
)

const (
	IngressPipyController     = commons.AnnotationPrefix + "/ingress-pipy"
	IngressPipyClass          = "pipy"
	NoDefaultIngressClass     = ""
	IngressAnnotationKey      = "kubernetes.io/ingress.class"
	IngressClassAnnotationKey = "ingressclass.kubernetes.io/is-default-class"

	PipyIngressAnnotationPrefix             = "pipy.ingress.kubernetes.io"
	PipyIngressAnnotationRewriteFrom        = PipyIngressAnnotationPrefix + "/rewrite-target-from"
	PipyIngressAnnotationRewriteTo          = PipyIngressAnnotationPrefix + "/rewrite-target-to"
	PipyIngressAnnotationSessionSticky      = PipyIngressAnnotationPrefix + "/session-sticky"
	PipyIngressAnnotationLoadBalancer       = PipyIngressAnnotationPrefix + "/lb-type"
	PipyIngressAnnotationUpstreamSSLName    = PipyIngressAnnotationPrefix + "/upstream-ssl-name"
	PipyIngressAnnotationUpstreamSSLSecret  = PipyIngressAnnotationPrefix + "/upstream-ssl-secret"
	PipyIngressAnnotationUpstreamSSLVerify  = PipyIngressAnnotationPrefix + "/upstream-ssl-verify"
	PipyIngressAnnotationTLSVerifyClient    = PipyIngressAnnotationPrefix + "/tls-verify-client"
	PipyIngressAnnotationTLSVerifyDepth     = PipyIngressAnnotationPrefix + "/tls-verify-depth"
	PipyIngressAnnotationTLSTrustedCASecret = PipyIngressAnnotationPrefix + "/tls-trusted-ca-secret"
)
