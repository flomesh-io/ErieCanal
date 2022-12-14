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

package repo

type Codebase struct {
	Version     int64    `json:"version,string,omitempty"`
	Path        string   `json:"path,omitempty"`
	Main        string   `json:"main,omitempty"`
	Base        string   `json:"base,omitempty"`
	Files       []string `json:"files,omitempty"`
	EditFiles   []string `json:"editFiles,omitempty"`
	ErasedFiles []string `json:"erasedFiles,omitempty"`
	Derived     []string `json:"derived,omitempty"`
	// Instances []interface, this field is not used so far by operator, just ignore it
}

type Router struct {
	Routes RouterEntry `json:"routes"`
}

type RouterEntry map[string]ServiceInfo

type ServiceInfo struct {
	Service string   `json:"service,omitempty"`
	Rewrite []string `json:"rewrite,omitempty"`
}

type Batch struct {
	Basepath string
	Items    []BatchItem
}

type BatchItem struct {
	Path     string
	Filename string
	Content  interface{}
}

type ServiceRegistry struct {
	Services ServiceRegistryEntry `json:"services"`
}

type ServiceRegistryEntry map[string][]string
