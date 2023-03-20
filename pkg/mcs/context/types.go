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

package context

import (
	"context"
	"github.com/flomesh-io/ErieCanal/pkg/mcs/config"
	"k8s.io/client-go/rest"
	"time"
)

type ConnectorContext struct {
	context.Context
	ClusterKey      string
	SpecHash        string
	KubeConfig      *rest.Config
	ConnectorConfig *config.ConnectorConfig
	Cancel          func()
	StopCh          chan struct{}
}

// ConnectorCtxKey the pointer is the key that a ConnectorContext returns itself for.
var ConnectorCtxKey int

// Deadline returns the time when work done on behalf of this context
// should be canceled. Deadline returns ok==false when no deadline is
// set. Successive calls to Deadline return the same results.
func (c *ConnectorContext) Deadline() (deadline time.Time, ok bool) {
	return
}

// Done returns a channel that's closed when work done on behalf of this
// context should be canceled. Done may return nil if this context can
// never be canceled. Successive calls to Done return the same value.
// The close of the Done channel may happen asynchronously,
// after the cancel function returns.
func (c *ConnectorContext) Done() <-chan struct{} {
	return nil
}

// Err returns nil, if Done is not yet closed,
// If Done is closed, Err returns a non-nil error explaining why:
// Canceled if the context was canceled
// or DeadlineExceeded if the context's deadline passed.
// After Err returns a non-nil error, successive calls to Err return the same error.
func (c *ConnectorContext) Err() error {
	return nil
}

// Value returns the value associated with this context for key, or nil
// if no value is associated with key. Successive calls to Value with
// the same key returns the same result.
func (c *ConnectorContext) Value(key interface{}) interface{} {
	if key == &ConnectorCtxKey {
		return c
	}
	return nil
}
