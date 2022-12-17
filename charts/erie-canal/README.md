# ErieCanal 

![GitHub](https://img.shields.io/github/license/flomesh-io/ErieCanal?style=flat-square)
![GitHub release (latest by date including pre-releases)](https://img.shields.io/github/v/release/flomesh-io/ErieCanal?include_prereleases&style=flat-square)
![GitHub tag (latest SemVer pre-release)](https://img.shields.io/github/v/tag/flomesh-io/ErieCanal?include_prereleases&style=flat-square)
![GitHub (Pre-)Release Date](https://img.shields.io/github/release-date-pre/flomesh-io/ErieCanal?style=flat-square)

[ErieCanal ](https://github.com/flomesh-io/ErieCanal) with Pipy proxy at its core is Kubernetes North-South Traffic Manager and provides Ingress controllers, Gateway API, and cross-cluster service registration and service discovery. Thanks to Pipy's “ARM Ready” capabilities, ErieCanal is well suited for cloud and edge computing.

## Introduction

This chart bootstraps a ErieCanal deployment on a [Kubernetes](http://kubernetes.io) cluster using the [Helm](https://helm.sh) package manager.

## Prerequisites

- Kubernetes 1.19+

## Installing the Chart

To install the chart with the release name `erie-canal` run:

```bash
$ helm repo add erie-canal https://flomesh-io.github.io/ErieCanal
$ helm install erie-canal erie-canal/erie-canal --namespace erie-canal --create-namespace
```

The command deploys ErieCanal on the Kubernetes cluster using the default configuration in namespace `flomesh` and creates the namespace if it doesn't exist. The [configuration](#configuration) section lists the parameters that can be configured during installation.

As soon as all pods are up and running, you can start to evaluate ErieCanal.

## Uninstalling the Chart

To uninstall the `erie-canal` deployment run:

```bash
$ helm uninstall erie-canal --namespace erie-canal
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Configuration

Please see the [values schema reference documentation](https://artifacthub.io/packages/helm/erie-canal/erie-canal?modal=values-schema) for a list of the configurable parameters of the chart and their default values.

Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`. For example,

```bash
$ helm install erie-canal erie-canal/erie-canal --namespace erie-canal --create-namespace \
  --set ec.image.pullPolicy=Always
```

Alternatively, a YAML file that specifies the values for the parameters can be provided while installing the chart. For example,

```bash
$ helm install erie-canal erie-canal/erie-canal --namespace erie-canal --create-namespace -f values-override.yaml
```
