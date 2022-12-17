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

package util

import (
	"context"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func CreateOrUpdate(ctx context.Context, c client.Client, obj client.Object) (controllerutil.OperationResult, error) {
	// a copy of new object
	modifiedObj := obj.DeepCopyObject().(client.Object)
	klog.V(5).Infof("Modified: %#v", modifiedObj)

	key := client.ObjectKeyFromObject(obj)
	if err := c.Get(ctx, key, obj); err != nil {
		if !apierrors.IsNotFound(err) {
			klog.Errorf("Get Object %v, %s err: %s", obj.GetObjectKind().GroupVersionKind(), key, err)
			return controllerutil.OperationResultNone, err
		}
		klog.V(5).Infof("Creating Object %v, %s ...", obj.GetObjectKind().GroupVersionKind(), key)
		if err := c.Create(ctx, obj); err != nil {
			klog.Errorf("Create Object %s err: %s", key, err)
			return controllerutil.OperationResultNone, err
		}

		klog.V(5).Infof("Object %v, %s is created successfully.", obj.GetObjectKind().GroupVersionKind(), key)
		return controllerutil.OperationResultCreated, nil
	}
	klog.V(5).Infof("Found Object %v, %s: %#v", obj.GetObjectKind().GroupVersionKind(), key, obj)

	result := controllerutil.OperationResultNone
	if !reflect.DeepEqual(obj, modifiedObj) {
		klog.V(5).Infof("Patching Object %v, %s ...", obj.GetObjectKind().GroupVersionKind(), key)
		patchData, err := client.Merge.Data(modifiedObj)
		if err != nil {
			klog.Errorf("Create ApplyPatch err: %s", err)
			return controllerutil.OperationResultNone, err
		}
		klog.V(5).Infof("Patch data = \n\n%s\n\n", string(patchData))

		// Only issue a Patch if the before and after resources differ
		if err := c.Patch(
			ctx,
			obj,
			client.RawPatch(types.MergePatchType, patchData),
			&client.PatchOptions{FieldManager: "ErieCanal"},
		); err != nil {
			klog.Errorf("Patch Object %v, %s err: %s", obj.GetObjectKind().GroupVersionKind(), key, err)
			return result, err
		}
		result = controllerutil.OperationResultUpdated
	}

	klog.V(5).Infof("Object %v, %s is %s successfully.", obj.GetObjectKind().GroupVersionKind(), key, result)
	return result, nil
}
