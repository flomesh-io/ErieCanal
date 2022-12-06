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
