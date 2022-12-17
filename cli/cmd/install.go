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
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"github.com/flomesh-io/ErieCanal/pkg/kube"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	helm "helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/strvals"
	"io"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"time"
)

//go:embed chart.tgz
var chartSource []byte

type installCmd struct {
	out            io.Writer
	meshName       string
	namespace      string
	version        string
	timeout        time.Duration
	chartRequested *chart.Chart
	setOptions     []string
	atomic         bool
	k8sApi         *kube.K8sAPI
}

func newCmdInstall(config *helm.Configuration, out io.Writer) *cobra.Command {
	inst := &installCmd{
		out: out,
	}

	cmd := &cobra.Command{
		Use:   "install [flags]",
		Args:  cobra.NoArgs,
		Short: "Install ErieCanal components and its dependencies.",
		Long:  "Install ErieCanal components and its dependencies.",

		RunE: func(cmd *cobra.Command, args []string) error {
			api, err := kube.NewAPI(30 * time.Second)
			if err != nil {
				return errors.Errorf("Error creating K8sAPI Client: %s", err)
			}
			inst.k8sApi = api

			return inst.run(config)
		},
	}

	f := cmd.Flags()
	f.StringVar(&inst.meshName, "mesh-name", "erie-canal", "name for the new control plane instance")
	f.StringVar(&inst.namespace, "namespace", "erie-canal", "namespace for the new control plane instance")
	//f.StringVar(&inst.version, "version", "", "version of ErieCanal")
	f.DurationVar(&inst.timeout, "timeout", 10*time.Minute, "Time to wait for installation and resources in a ready state, zero means no timeout")
	f.StringArrayVar(&inst.setOptions, "set", nil, "Set arbitrary chart values (can specify multiple or separate values with commas: key1=val1,key2=val2)")

	_ = config.Init(&genericclioptions.ConfigFlags{
		Namespace: &inst.namespace,
	}, inst.namespace, "secret", func(format string, v ...interface{}) {})

	return cmd
}

func (i *installCmd) run(config *helm.Configuration) error {
	if err := i.loadChart(); err != nil {
		return err
	}

	values, err := i.resolveValues()
	if err != nil {
		return err
	}
	installClient := helm.NewInstall(config)
	installClient.ReleaseName = i.meshName
	installClient.Namespace = i.namespace
	installClient.CreateNamespace = true
	installClient.Wait = true
	installClient.Atomic = i.atomic
	installClient.Timeout = i.timeout
	//installClient.Version = i.version

	if _, err = installClient.Run(i.chartRequested, values); err != nil {

		pods, _ := i.k8sApi.Client.CoreV1().Pods(i.namespace).List(context.TODO(), metav1.ListOptions{})

		for _, pod := range pods.Items {
			fmt.Fprintf(i.out, "Status for pod %s in namespace %s:\n %v\n\n", pod.Name, pod.Namespace, pod.Status)
		}
		return err
	}

	fmt.Fprintf(i.out, "ErieCanal installed successfully in namespace [%s] with mesh name [%s]\n", i.namespace, i.meshName)

	return nil
}

func (i *installCmd) loadChart() error {
	chartRequested, err := loader.LoadArchive(bytes.NewReader(chartSource))
	if err != nil {
		return errors.Errorf("error loading chart for installation: %s", err)
	}

	i.chartRequested = chartRequested
	return nil
}

func (i *installCmd) resolveValues() (map[string]interface{}, error) {
	finalValues := map[string]interface{}{}

	if err := parseVal(i.setOptions, finalValues); err != nil {
		return nil, errors.Wrap(err, "invalid format for --set")
	}

	return finalValues, nil
}

func parseVal(vals []string, parsedVals map[string]interface{}) error {
	for _, v := range vals {
		if err := strvals.ParseInto(v, parsedVals); err != nil {
			return err
		}
	}
	return nil
}
