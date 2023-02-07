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
	"encoding/json"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type mutatingHandler struct {
	defaulter Defaulter
	decoder   *admission.Decoder
}

var _ admission.DecoderInjector = &mutatingHandler{}

// InjectDecoder injects the decoder into a mutatingHandler.
func (h *mutatingHandler) InjectDecoder(d *admission.Decoder) error {
	h.decoder = d
	return nil
}

// Handle handles admission requests.
func (h *mutatingHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	if h.defaulter == nil {
		panic("defaulter should never be nil")
	}

	obj := h.defaulter.RuntimeObject()
	if obj == nil {
		return admission.Allowed("Not supported Kind")
	}

	if err := h.decoder.Decode(req, obj); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	// Default the object
	h.defaulter.SetDefaults(obj)
	marshalled, err := json.Marshal(obj)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	// Create the patch
	return admission.PatchResponseFromRaw(req.Object.Raw, marshalled)
}

func DefaultingWebhookFor(defaulter Defaulter) *admission.Webhook {
	return &admission.Webhook{
		Handler: &mutatingHandler{defaulter: defaulter},
	}
}
