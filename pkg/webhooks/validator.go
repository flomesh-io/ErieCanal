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

package webhooks

import (
	"context"
	goerrors "errors"
	admissionv1 "k8s.io/api/admission/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type validatingHandler struct {
	validator Validator
	decoder   *admission.Decoder
}

var _ admission.DecoderInjector = &validatingHandler{}

// InjectDecoder injects the decoder into a validatingHandler.
func (h *validatingHandler) InjectDecoder(d *admission.Decoder) error {
	h.decoder = d
	return nil
}

// Handle handles admission requests.
func (h *validatingHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	if h.validator == nil {
		panic("validator should never be nil")
	}

	// Get the object in the request
	obj := h.validator.RuntimeObject()
	if obj == nil {
		return admission.Allowed("Not supported Kind")
	}

	if err := h.decoder.Decode(req, obj); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	switch req.Operation {
	case admissionv1.Create:
		err := h.decoder.Decode(req, obj)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}

		err = h.validator.ValidateCreate(obj)
		if err != nil {
			var apiStatus apierrors.APIStatus
			if goerrors.As(err, &apiStatus) {
				return validationResponseFromStatus(false, apiStatus.Status())
			}
			return admission.Denied(err.Error())
		}
	case admissionv1.Update:
		oldObj := obj.DeepCopyObject()

		err := h.decoder.DecodeRaw(req.Object, obj)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		err = h.decoder.DecodeRaw(req.OldObject, oldObj)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}

		err = h.validator.ValidateUpdate(oldObj, obj)
		if err != nil {
			var apiStatus apierrors.APIStatus
			if goerrors.As(err, &apiStatus) {
				return validationResponseFromStatus(false, apiStatus.Status())
			}
			return admission.Denied(err.Error())
		}
	case admissionv1.Delete:
		err := h.decoder.DecodeRaw(req.OldObject, obj)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}

		err = h.validator.ValidateDelete(obj)
		if err != nil {
			var apiStatus apierrors.APIStatus
			if goerrors.As(err, &apiStatus) {
				return validationResponseFromStatus(false, apiStatus.Status())
			}
			return admission.Denied(err.Error())
		}
	}

	return admission.Allowed("")
}

func validationResponseFromStatus(allowed bool, status metav1.Status) admission.Response {
	resp := admission.Response{
		AdmissionResponse: admissionv1.AdmissionResponse{
			Allowed: allowed,
			Result:  &status,
		},
	}
	return resp
}

func ValidatingWebhookFor(validator Validator) *admission.Webhook {
	return &admission.Webhook{
		Handler: &validatingHandler{validator: validator},
	}
}
