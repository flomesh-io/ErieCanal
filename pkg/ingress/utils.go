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
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/klog/v2"
)

func IsValidPipyIngress(ing *networkingv1.Ingress) bool {
	// 1. with annotation or IngressClass
	ingressClass, ok := ing.GetAnnotations()[IngressAnnotationKey]
	if !ok && ing.Spec.IngressClassName != nil {
		ingressClass = *ing.Spec.IngressClassName
	}

	defaultClass := DefaultIngressClass
	klog.V(3).Infof("IngressClassName/IngressAnnotation = %s", ingressClass)
	klog.V(3).Infof("DefaultIngressClass = %s, and IngressPipyClass = %s", defaultClass, IngressPipyClass)

	// 2. empty IngressClass, and pipy is the default IngressClass or no default at all
	if len(ingressClass) == 0 && (defaultClass == IngressPipyClass || len(defaultClass) == 0) {
		return true
	}

	// 3. with IngressClass
	return ingressClass == IngressPipyClass
}
