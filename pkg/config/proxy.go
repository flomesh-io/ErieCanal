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

package config

type ProxyInitEnvironmentConfiguration struct {
	MatchedProxyProfile string   `envconfig:"MATCHED_PROXY_PROFILE" required:"true" split_words:"true"`
	ProxyRepoBaseUrl    string   `envconfig:"PROXY_REPO_BASE_URL" required:"true" split_words:"true"`
	ProxyRepoRootUrl    string   `envconfig:"PROXY_REPO_ROOT_URL" required:"true" split_words:"true"`
	ProxyParentPath     string   `envconfig:"PROXY_PARENT_PATH" required:"true" split_words:"true"`
	ProxyPaths          []string `envconfig:"PROXY_PATHS" required:"true" split_words:"true"`
}
