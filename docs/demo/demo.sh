#!/bin/bash

cd $(dirname ${BASH_SOURCE})
. ./common.sh

DEMO_AUTO_RUN=true

NAMESPACE=httpbin
desc "deploy sample app httpbin under the ${NAMESPACE} on clusters cluster-1 and cluster-3"
for CONFIG in kubeconfig_c1 kubeconfig_c3
do
    CLUSTER_NAME=$(if [ "${CONFIG}" == "kubeconfig_c1" ]; then echo "cluster-1"; else echo "cluster-3"; fi)
    desc "installing on cluster ${CLUSTER_NAME}"
    kube="kubectl --kubeconfig ${!CONFIG}"
    run "$kube create ns ${NAMESPACE}"
    run "KUBECONFIG=${!CONFIG} $osm_binary namespace add ${NAMESPACE}"
    run "$kube apply -n ${NAMESPACE} -f - <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: httpbin
  labels:
    app: pipy
spec:
  replicas: 1
  selector:
    matchLabels:
      app: pipy
  template:
    metadata:
      labels:
        app: pipy
    spec:
      containers:
        - name: pipy
          image: flomesh/pipy:latest
          ports:
            - containerPort: 8080
          command:
            - pipy
            - -e
            - |
              pipy()
              .listen(8080)
              .serveHTTP(new Message('Hi, I am from ${CLUSTER_NAME} and controlled by mesh!\n'))
---
apiVersion: v1
kind: Service
metadata:
  name: httpbin
spec:
  ports:
    - port: 8080
      targetPort: 8080
      protocol: TCP
  selector:
    app: pipy
---
apiVersion: v1
kind: Service
metadata:
  name: httpbin-${CLUSTER_NAME}
spec:
  ports:
    - port: 8080
      targetPort: 8080
      protocol: TCP
  selector:
    app: pipy
EOF"
  run "sleep 1"
  run "$kube wait --for=condition=ready pod -n ${NAMESPACE} --all --timeout=60s"
done


NAMESPACE=curl
desc "deploy sample app curl under the ${NAMESPACE} on cluster-2"
kube="kubectl --kubeconfig ${kubeconfig_c2}"
run "$kube create ns ${NAMESPACE}"
run "KUBECONFIG=${kubeconfig_c2} $osm_binary namespace add ${NAMESPACE}"
run "$kube apply -n ${NAMESPACE} -f - <<EOF
apiVersion: v1
kind: ServiceAccount
metadata:
  name: curl
  namespace: curl 
---
apiVersion: v1
kind: Service
metadata:
  name: curl
  labels:
    app: curl
    service: curl
spec:
  ports:
    - name: http
      port: 80
  selector:
    app: curl
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: curl
spec:
  replicas: 1
  selector:
    matchLabels:
      app: curl
  template:
    metadata:
      labels:
        app: curl
    spec:
      serviceAccountName: curl
      containers:
      - image: curlimages/curl
        imagePullPolicy: IfNotPresent
        name: curl
        command: ["sleep", "365d"]
EOF"

run "sleep 1"
run "$kube wait --for=condition=ready pod -n ${NAMESPACE} --all --timeout=60s"

desc "Let's export services in cluster-1 and cluster-3"
NAMESPACE_MESH=httpbin
for CONFIG in kubeconfig_c1 kubeconfig_c3
do
    CLUSTER_NAME=$(if [ "${CONFIG}" == "kubeconfig_c1" ]; then echo "cluster-1"; else echo "cluster-3"; fi)
    desc "exporting service on cluster ${CLUSTER_NAME}"
    kube="kubectl --kubeconfig ${!CONFIG}"
    run "$kube apply -f - <<EOF
apiVersion: flomesh.io/v1alpha1
kind: ServiceExport
metadata:
  namespace: ${NAMESPACE_MESH}
  name: httpbin
spec:
  serviceAccountName: \"*\"
  rules:
    - portNumber: 8080
      path: \"/${CLUSTER_NAME}/httpbin-mesh\"
      pathType: Prefix
---
apiVersion: flomesh.io/v1alpha1
kind: ServiceExport
metadata:
  namespace: ${NAMESPACE_MESH}
  name: httpbin-${CLUSTER_NAME}
spec:
  serviceAccountName: \"*\"
  rules:
    - portNumber: 8080
      path: \"/${CLUSTER_NAME}/httpbin-mesh-${CLUSTER_NAME}\"
      pathType: Prefix
EOF"
run "sleep 1"   
done

desc "After exporting the services, FSM will automatically create ingress rules for them, and with the rules, you can access these services through Ingress"
for CONFIG in kubeconfig_c1 kubeconfig_c3
do
    CLUSTER_NAME=$(if [ "${CONFIG}" == "kubeconfig_c1" ]; then echo "cluster-1"; elif [ "${CONFIG}" == "kubeconfig_c2" ]; then echo "cluster-2"; else echo "cluster-3"; fi)
    ((PORT=80+${CLUSTER_NAME: -1}))
    kube="kubectl --kubeconfig ${!CONFIG}"
    desc "Getting service exported in cluster ${CLUSTER_NAME}"
    run "$kube get serviceexports.flomesh.io -A"
    desc "calling service in cluster ${CLUSTER_NAME}"
    run "curl -s http://${HOST_IP}:${PORT}/${CLUSTER_NAME}/httpbin-mesh"
    echo ""
    run "curl -s http://${HOST_IP}:${PORT}/${CLUSTER_NAME}/httpbin-mesh-${CLUSTER_NAME}"
    echo ""
done

desc "exported services can be imported into other managed clusters."
desc "For example, let's look at cluster-2, and we can see multiple services imported"
run "$k2 get serviceimports -A"

desc "Let's see if we can access these imported services"
curl_client="$($k2 get pod -n curl -l app=curl -o jsonpath='{.items[0].metadata.name}')"
run "$k2 exec "${curl_client}" -n curl -c curl -- curl -s http://httpbin.httpbin:8080/"
desc "by default no other cluster instance will be used to respond to requests. To access cross cluster services"
desc "we need to work with GlobalTrafficPolicy CRD"
desc "Note that all global traffic policies are set on the user’s side, so this demo is about setting global traffic policies on the cluster-2"
desc "For example: if we want to access http://httpbin.httpbin:8080, we need to create GlobalTrafficPolicy resource"
run "$k2 apply -n httpbin -f - <<EOF
apiVersion: flomesh.io/v1alpha1
kind: GlobalTrafficPolicy
metadata:
  name: httpbin
spec:
  lbType: FailOver
  targets:
    - clusterKey: default/default/default/cluster-1
    - clusterKey: default/default/default/cluster-3
EOF"

run "sleep 3"
desc "We have a multi-cluster service!"
desc "See for yourself"
run "$k2 exec "${curl_client}" -n curl -c curl -- curl -s http://httpbin.httpbin:8080/"
run "$k2 exec "${curl_client}" -n curl -c curl -- curl -s http://httpbin.httpbin:8080/"
desc "(Enter to exit)"
read -s
