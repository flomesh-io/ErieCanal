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
	"os"
	"os/signal"
	"syscall"
)

func RegisterExitHandlers(shutdownFuncs ...func()) (stop chan struct{}) {
	stop = make(chan struct{})

	go func() {
		// Block until any signal is received.
		<-stop

		// execute our shutdown functions
		for _, f := range shutdownFuncs {
			f()
		}
	}()

	return stop
}

func RegisterOSExitHandlers(shutdownFuncs ...func()) (stop chan struct{}) {
	var exitSignals = []os.Signal{syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL}

	stop = make(chan struct{})
	s := make(chan os.Signal, len(exitSignals))
	signal.Notify(s, exitSignals...)

	go func() {
		// Wait for a signal from the OS before dispatching
		// a stop signal to all other goroutines observing this channel.
		<-s
		close(stop)

		// execute our shutdown functions
		for _, f := range shutdownFuncs {
			f()
		}
	}()

	return stop
}
