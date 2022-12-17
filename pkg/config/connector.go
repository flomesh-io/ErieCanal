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

import (
	"fmt"
	"github.com/flomesh-io/ErieCanal/pkg/commons"
	"github.com/flomesh-io/ErieCanal/pkg/util"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/klog/v2"
	"net"
)

type ConnectorConfig struct {
	name            string
	region          string
	zone            string
	group           string
	inCluster       bool
	key             string
	gatewayHost     string
	gatewayIP       net.IP
	gatewayPort     int32
	controlPlaneUID string
}

func NewConnectorConfig(
	region, zone, group, name, gatewayHost string,
	gatewayPort int32,
	inCluster bool,
	controlPlaneUID string,
) (*ConnectorConfig, error) {
	clusterKey := util.EvaluateTemplate(commons.ClusterIDTemplate, struct {
		Region  string
		Zone    string
		Group   string
		Cluster string
	}{
		Region:  region,
		Zone:    zone,
		Group:   group,
		Cluster: name,
	})

	c := &ConnectorConfig{
		region:          region,
		zone:            zone,
		group:           group,
		name:            name,
		inCluster:       inCluster,
		key:             clusterKey,
		controlPlaneUID: controlPlaneUID,
	}

	if !inCluster {
		isDNSName := false
		if ipErrs := validation.IsValidIPv4Address(field.NewPath(""), gatewayHost); len(ipErrs) > 0 {
			// Not IPv4 address
			klog.Warningf("%q is NOT a valid IPv4 address: %v", gatewayHost, ipErrs)
			if dnsErrs := validation.IsDNS1123Subdomain(gatewayHost); len(dnsErrs) > 0 {
				// Not valid DNS domain name
				return nil, fmt.Errorf("invalid DNS name or IP %q: %v", gatewayHost, dnsErrs)
			} else {
				// is DNS name
				isDNSName = true
			}
		}

		var gwIPv4 net.IP
		if isDNSName {
			ipAddr, err := net.ResolveIPAddr("ip4", gatewayHost)
			if err != nil {
				return nil, fmt.Errorf("%q cannot be resolved to IP, %s", gatewayHost, err)
			}
			klog.Infof("%q is resolved to IP: %s", gatewayHost, ipAddr.IP)
			gwIPv4 = ipAddr.IP.To4()
		} else {
			gwIPv4 = net.ParseIP(gatewayHost).To4()
		}

		if gwIPv4 == nil {
			return nil, fmt.Errorf("%q cannot be resolved to a IPv4 address", gatewayHost)
		}

		if gwIPv4 != nil && (gwIPv4.IsLoopback() || gwIPv4.IsUnspecified()) {
			return nil, fmt.Errorf("gateway Host %s is resolved to Loopback IP or Unspecified", gatewayHost)
		}

		c.gatewayHost = gatewayHost
		c.gatewayPort = gatewayPort
		c.gatewayIP = gwIPv4
	}

	return c, nil
}

func (c *ConnectorConfig) Name() string {
	return c.name
}

func (c *ConnectorConfig) Region() string {
	return c.region
}

func (c *ConnectorConfig) Zone() string {
	return c.zone
}

func (c *ConnectorConfig) Group() string {
	return c.group
}

func (c *ConnectorConfig) IsInCluster() bool {
	return c.inCluster
}

func (c *ConnectorConfig) Key() string {
	return c.key
}

func (c *ConnectorConfig) GatewayHost() string {
	if c.inCluster {
		return ""
	}
	return c.gatewayHost
}

func (c *ConnectorConfig) GatewayIP() net.IP {
	if c.inCluster {
		return net.IPv4zero
	}
	return c.gatewayIP
}

func (c *ConnectorConfig) GatewayPort() int32 {
	if c.inCluster {
		return 0
	}
	return c.gatewayPort
}

func (c *ConnectorConfig) ControlPlaneUID() string {
	return c.controlPlaneUID
}
