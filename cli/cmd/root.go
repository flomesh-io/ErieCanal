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
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"helm.sh/helm/v3/pkg/action"
	"os"
)

var (
	stdout = color.Output
	stderr = color.Error
)

var RootCmd = &cobra.Command{
	Use:   "erie-canal",
	Short: "erie-canal manages the ErieCanal",
	Long:  "erie-canal manages the ErieCanal",

	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	actionConfig := new(action.Configuration)
	RootCmd.AddCommand(newCmdInstall(actionConfig, stdout))
	RootCmd.AddCommand(newCmdUninstall(actionConfig, os.Stdin, stdout))
	RootCmd.AddCommand(newCmdVersion(stdout))

	// run when each command's execute method is called
	//cobra.OnInitialize(func() {
	//	if err := actionConfig.Init(settings.RESTClientGetter(), settings.Namespace(), "secret", debug); err != nil {
	//		os.Exit(1)
	//	}
	//})
}
