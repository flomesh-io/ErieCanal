# Setup local environment for testing Multi-Cluster

## Prerequisites
* Golang 1.19
* kind
* Helm
* kubecm(recommended)

## Checkout code
```shell
git clone -b feature/service-export-n-import git@github.com:flomesh-io/ErieCanalgit
```

Then, you need to make the project, generate Helm charts and copy them to desired folders:
```shell
make dev
```

## Create Kubernetes Clusters and Install ErieCanal

### Control Plane
#### Create Control Plane Cluster
```shell
kind create cluster --config=samples/setup/kind/control-plane.yaml
```

You'll see the output like below:
```shell
❯ kind create cluster --config=samples/setup/kind/control-plane.yaml
Creating cluster "control-plane" ...
 ✓ Ensuring node image (kindest/node:v1.21.12) 🖼
 ✓ Preparing nodes 📦
 ✓ Writing configuration 📜
 ✓ Starting control-plane 🕹️
 ✓ Installing CNI 🔌
 ✓ Installing StorageClass 💾
Set kubectl context to "kind-control-plane"
You can now use your cluster with:

kubectl cluster-info --context kind-control-plane

Not sure what to do next? 😅  Check out https://kind.sigs.k8s.io/docs/user/quick-start/
```

#### Install ErieCanal to Control Plane
```shell
helm install --namespace flomesh --create-namespace --set ErieCanal.version=0.2.0-alpha.1-dev --set ErieCanal.logLevel=5 --set ErieCanal.serviceLB.enabled=true erie-canal charts/erie-canal/
```

### Cluster 1
#### Create Cluster1
```shell
kind create cluster --config=samples/setup/kind/cluster1.yaml
```

You'll see the output like below:
```shell
❯ kind create cluster --config=samples/setup/kind/cluster1.yaml
Creating cluster "cluster1" ...
 ✓ Ensuring node image (kindest/node:v1.21.12) 🖼
 ✓ Preparing nodes 📦
 ✓ Writing configuration 📜
 ✓ Starting control-plane 🕹️
 ✓ Installing CNI 🔌
 ✓ Installing StorageClass 💾
Set kubectl context to "kind-cluster1"
You can now use your cluster with:

kubectl cluster-info --context kind-cluster1

Not sure what to do next? 😅  Check out https://kind.sigs.k8s.io/docs/user/quick-start/
```

#### Install ErieCanal to Cluster1
```shell
helm install --namespace flomesh --create-namespace --set ErieCanal.version=0.2.0-alpha.1-dev --set ErieCanal.logLevel=5 --set ErieCanal.serviceLB.enabled=true erie-canal charts/erie-canal/
```

### Cluster 2
#### Create Cluster1
```shell
kind create cluster --config=samples/setup/kind/cluster2.yaml
```

You'll see the output like below:
```shell
❯ kind create cluster --config=samples/setup/kind/cluster2.yaml
Creating cluster "cluster2" ...
 ✓ Ensuring node image (kindest/node:v1.21.12) 🖼
 ✓ Preparing nodes 📦
 ✓ Writing configuration 📜
 ✓ Starting control-plane 🕹️
 ✓ Installing CNI 🔌
 ✓ Installing StorageClass 💾
Set kubectl context to "kind-cluster2"
You can now use your cluster with:

kubectl cluster-info --context kind-cluster2

Not sure what to do next? 😅  Check out https://kind.sigs.k8s.io/docs/user/quick-start/
```

#### Install ErieCanal to Cluster2
```shell
helm install --namespace flomesh --create-namespace --set ErieCanal.version=0.2.0-alpha.1-dev --set ErieCanal.logLevel=5 --set ErieCanal.serviceLB.enabled=true erie-canal charts/erie-canal/
```

## Create/Update Cluster CRD yamls

### Get kubeconfig of cluster1
```shell
kind get kubeconfig -n cluster1
```

Copy the output and replace the kubeconfig section of `samples/cluster/cluster1.yaml`:
```yaml
apiVersion: flomesh.io/v1alpha1
kind: Cluster
metadata:
  name: cluster1
spec:
  gateway: demo.flomesh.internal:8091
  kubeconfig: |+
    apiVersion: v1
    clusters:
    - cluster:
        certificate-authority-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUJlRENDQVIyZ0F3SUJBZ0lCQURBS0JnZ3Foa2pPUFFRREFqQWpNU0V3SHdZRFZRUUREQmhyTTNNdGMyVnkKZG1WeUxXTmhRREUyTmpZMk5qQTJOVGd3SGhjTk1qSXhNREkxTURFeE56TTRXaGNOTXpJeE1ESXlNREV4TnpNNApXakFqTVNFd0h3WURWUVFEREJock0zTXRjMlZ5ZG1WeUxXTmhRREUyTmpZMk5qQTJOVGd3V1RBVEJnY3Foa2pPClBRSUJCZ2dxaGtqT1BRTUJCd05DQUFUUUxrbEs4NExOVjZnV1FJbmUwTk9McFBDUmloK0gvNXhQb0Jod1lLdUYKaUhwNHc2UEs3VE94Y2JON3FlWEtDNnFhNG4xWjRLc2lhaVBYQjJZY1l5djJvMEl3UURBT0JnTlZIUThCQWY4RQpCQU1DQXFRd0R3WURWUjBUQVFIL0JBVXdBd0VCL3pBZEJnTlZIUTRFRmdRVXhXVFJjVVZTY3UySUdsV0FEa0xxClJ4bDUvOXN3Q2dZSUtvWkl6ajBFQXdJRFNRQXdSZ0loQVAwakRZZHFUdlErN1FMRDRhMDdZRENNSjUvK2k3M0kKdGpVbzF6OWZzT3RRQWlFQXhCN2ZzVXEwS3RwaFhYNUQ3OXEzWEgrRUdoKzllY0FlN3Z0ZVRFYlQxS0U9Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K
        server: https://demo.flomesh.internal:6446
      name: k3d-cluster1
    contexts:
    - context:
        cluster: k3d-cluster1
        user: admin@k3d-cluster1
      name: k3d-cluster1
    current-context: k3d-cluster1
    kind: Config
    preferences: {}
    users:
    - name: admin@k3d-cluster1
      user:
        client-certificate-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUJrVENDQVRlZ0F3SUJBZ0lJZlJCRitvdG5WU2t3Q2dZSUtvWkl6ajBFQXdJd0l6RWhNQjhHQTFVRUF3d1kKYXpOekxXTnNhV1Z1ZEMxallVQXhOalkyTmpZd05qVTRNQjRYRFRJeU1UQXlOVEF4TVRjek9Gb1hEVEl6TVRBeQpOVEF4TVRjek9Gb3dNREVYTUJVR0ExVUVDaE1PYzNsemRHVnRPbTFoYzNSbGNuTXhGVEFUQmdOVkJBTVRESE41CmMzUmxiVHBoWkcxcGJqQlpNQk1HQnlxR1NNNDlBZ0VHQ0NxR1NNNDlBd0VIQTBJQUJEWVQ3eWFBajV1ZnM5bVkKVjN1blBqMTR2N1dSN05CM2Y0Q3pNWGhaUDdZaVFHQ1orRVhYdEMybFBPc2NiRFdlR21pVGJqT0s4WkE4bG1iRAo3L0lVMFNDalNEQkdNQTRHQTFVZER3RUIvd1FFQXdJRm9EQVRCZ05WSFNVRUREQUtCZ2dyQmdFRkJRY0RBakFmCkJnTlZIU01FR0RBV2dCU2NtRU9aTU8zTUNtTVdOdWZ6UFBubWxyT1k1VEFLQmdncWhrak9QUVFEQWdOSUFEQkYKQWlBZm1TMGRQWVhaemJjVmdwb2kvZXYzL3FqeGFNUTR6WHd6RmJEYUtPUGpJQUloQU9yYlZ4WlV1NlBIK25aRQpyZDJyK1FnKzgxTVpYc2k0eUIyNmVCLzBPNUNFCi0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0KLS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUJkekNDQVIyZ0F3SUJBZ0lCQURBS0JnZ3Foa2pPUFFRREFqQWpNU0V3SHdZRFZRUUREQmhyTTNNdFkyeHAKWlc1MExXTmhRREUyTmpZMk5qQTJOVGd3SGhjTk1qSXhNREkxTURFeE56TTRXaGNOTXpJeE1ESXlNREV4TnpNNApXakFqTVNFd0h3WURWUVFEREJock0zTXRZMnhwWlc1MExXTmhRREUyTmpZMk5qQTJOVGd3V1RBVEJnY3Foa2pPClBRSUJCZ2dxaGtqT1BRTUJCd05DQUFUb1VOcU9jK3FiVjE3MUxSUEE0dFRLdXAzUFoyd1hhaGprZzc5TVZmd3QKblNSWW45RGd6SHUwS2gvbmY0aUxraW4rK010UGdtYUlpS3dMK2NSYWU1Qm1vMEl3UURBT0JnTlZIUThCQWY4RQpCQU1DQXFRd0R3WURWUjBUQVFIL0JBVXdBd0VCL3pBZEJnTlZIUTRFRmdRVW5KaERtVER0ekFwakZqYm44eno1CjVwYXptT1V3Q2dZSUtvWkl6ajBFQXdJRFNBQXdSUUlnWEhyaUV2Qmt5cUFpWFkvemlXeXBhai85MmVWYzBZeXAKK3lYWnQ3OWFocUlDSVFEckFBd0tzWnpKNjAweU5WYjBoT28yZFd4UnhhTHU0TnZ4VE9TOGtwZXpwUT09Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K
        client-key-data: LS0tLS1CRUdJTiBFQyBQUklWQVRFIEtFWS0tLS0tCk1IY0NBUUVFSUFiQjhJd2E2WTIwclJFQmx0aTVzQ1JyM0luVTZ5SjBDN3VYMlVtaURvYStvQW9HQ0NxR1NNNDkKQXdFSG9VUURRZ0FFTmhQdkpvQ1BtNSt6MlpoWGU2YytQWGkvdFpIczBIZC9nTE14ZUZrL3RpSkFZSm40UmRlMApMYVU4Nnh4c05aNGFhSk51TTRyeGtEeVdac1B2OGhUUklBPT0KLS0tLS1FTkQgRUMgUFJJVkFURSBLRVktLS0tLQo=
```

Just repeat the steps for cluster2 and remember to replace `cluster1` with `cluster2`.

## Add clusters to be managed by Control Plane
### Switch the context to Control Plane
```yaml
kubecm switch
```

Then select `kind-control-plane`:
```shell
❯ kc switch
Use the arrow keys to navigate: ↓ ↑ → ←  and / toggles search
Select Kube Context
    kind-cluster1(*)
    kind-cluster2
  😼 kind-control-plane
↓   lima-k3s

--------- Info ----------
Name:           kind-control-plane
Cluster:        kind-control-plane
User:           kind-control-plane
```

You'll see:
```shell
❯ kc switch
😸 Select:kind-control-plane
「/Users/linyang/.kube/config」 write successful!
+------------+-----------------------+-----------------------+-----------------------+-------------------------------+--------------+
|   CURRENT  |          NAME         |        CLUSTER        |          USER         |             SERVER            |   Namespace  |
+============+=======================+=======================+=======================+===============================+==============+
|            |     kind-cluster1     |     kind-cluster1     |     kind-cluster1     |      https://0.0.0.0:6446     |    default   |
+------------+-----------------------+-----------------------+-----------------------+-------------------------------+--------------+
|            |     kind-cluster2     |     kind-cluster2     |     kind-cluster2     |      https://0.0.0.0:6447     |    default   |
+------------+-----------------------+-----------------------+-----------------------+-------------------------------+--------------+
|      *     |   kind-control-plane  |   kind-control-plane  |   kind-control-plane  |      https://0.0.0.0:6445     |    default   |
+------------+-----------------------+-----------------------+-----------------------+-------------------------------+--------------+


Switched to context 「kind-control-plane」
```

### Add clusters
```shell
kubectl apply -f samples/cluster/cluster1.yaml
kubectl apply -f samples/cluster/cluster2.yaml
```

## Deploy test service 

### Deploy service to cluster1
Switch context to cluster1
```shell
kubecm switch
```

Select `kind-cluster1`
```shell
❯ kc switch
Use the arrow keys to navigate: ↓ ↑ → ←  and / toggles search
Select Kube Context
    kind-control-plane(*)
    k3s-72
  😼 kind-cluster1
↓   kind-cluster2

--------- Info ----------
Name:           kind-cluster1
Cluster:        kind-cluster1
User:           kind-cluster1
```

Then deploy the service:
```shell
kubectl apply -f samples/multi-cluster/001-pipy-deployment.yaml
```

### Deploy service to cluster2
Just repeat the steps for cluster2 and remember to replace `cluster1` with `cluster2`.


## Export service

Remember to check the current context before run the command:
```shell
kubectl apply -f samples/multi-cluster/100-service-export.yaml
```