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

package config

type ProxyInitEnvironmentConfiguration struct {
	MatchedProxyProfile string   `envconfig:"MATCHED_PROXY_PROFILE" required:"true" split_words:"true"`
	ProxyRepoBaseUrl    string   `envconfig:"PROXY_REPO_BASE_URL" required:"true" split_words:"true"`
	ProxyRepoRootUrl    string   `envconfig:"PROXY_REPO_ROOT_URL" required:"true" split_words:"true"`
	ProxyParentPath     string   `envconfig:"PROXY_PARENT_PATH" required:"true" split_words:"true"`
	ProxyPaths          []string `envconfig:"PROXY_PATHS" required:"true" split_words:"true"`
}
