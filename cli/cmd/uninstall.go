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
