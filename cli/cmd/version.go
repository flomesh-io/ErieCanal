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

package cmd

import (
	"fmt"
	"github.com/flomesh-io/ErieCanal/pkg/version"
	"github.com/spf13/cobra"
	"io"
)

const versionTpl = `

Version: %s
ImageVersion: %s
GitVersion: %s
GitCommit: %s
BuildDate: %s

`

func newCmdVersion(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the client and server version information",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(out, versionTpl, version.Version, version.ImageVersion, version.GitVersion, version.GitCommit, version.BuildDate)
			return nil
		},
	}

	return cmd
}
