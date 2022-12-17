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
	"bufio"
	"fmt"
	"github.com/flomesh-io/ErieCanal/pkg/kube"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"helm.sh/helm/v3/pkg/action"
	helm "helm.sh/helm/v3/pkg/action"
	"io"
	"strings"
	"time"
)

type uninstallMeshCmd struct {
	out       io.Writer
	in        io.Reader
	meshName  string
	namespace string
	force     bool
	k8sApi    *kube.K8sAPI
}

func newCmdUninstall(config *action.Configuration, in io.Reader, out io.Writer) *cobra.Command {
	uninstall := &uninstallMeshCmd{
		out: out,
		in:  in,
	}

	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "uninstall ErieCanal control plane instance",
		Long:  "uninstall ErieCanal control plane instance",
		Args:  cobra.ExactArgs(0),
		RunE: func(_ *cobra.Command, args []string) error {
			api, err := kube.NewAPI(30 * time.Second)
			if err != nil {
				return errors.Errorf("Error creating K8sAPI Client: %s", err)
			}
			uninstall.k8sApi = api

			return uninstall.run(config)
		},
	}

	f := cmd.Flags()
	f.StringVar(&uninstall.meshName, "mesh-name", "erie-canal", "Name of the service mesh")
	f.StringVar(&uninstall.namespace, "namespace", "erie-canal", "Namespace of the service mesh")
	f.BoolVarP(&uninstall.force, "force", "f", false, "Attempt to uninstall the ErieCanal control plane instance without prompting for confirmation.")

	return cmd
}

func (u *uninstallMeshCmd) run(config *helm.Configuration) error {
	uninstallClient := action.NewUninstall(config)

	confirm, err := confirm(u.in, u.out, fmt.Sprintf("\nUninstall ErieCanal [mesh name: %s] in namespace [%s] and/or ErieCanal resources ?", u.meshName, u.namespace), 3)
	if !confirm || err != nil {
		return err
	}
	_, err = uninstallClient.Run(u.meshName)
	if err != nil {
		return err
	}
	if err == nil {
		fmt.Fprintf(u.out, "ErieCanal [mesh name: %s] in namespace [%s] uninstalled\n", u.meshName, u.namespace)
	}

	return nil
}

func confirm(stdin io.Reader, stdout io.Writer, s string, tries int) (bool, error) {
	r := bufio.NewReader(stdin)

	for ; tries > 0; tries-- {
		fmt.Fprintf(stdout, "%s [y/n]: ", s)

		res, err := r.ReadString('\n')
		if err != nil {
			return false, err
		}

		// Empty input (i.e. "\n")
		if len(res) < 2 {
			continue
		}

		switch strings.ToLower(strings.TrimSpace(res)) {
		case "y":
			return true, nil
		case "n":
			return false, nil
		default:
			fmt.Fprintf(stdout, "Invalid input.\n")
			continue
		}
	}

	return false, nil
}
