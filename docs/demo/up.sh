#!/bin/bash

cd $(dirname ${BASH_SOURCE})
. ./common.sh

set -e

DEMO_AUTO_RUN=true

function check_command() {
    local installer="$2"
    if ! command -v $1 &> /dev/null
    then
        echo "$1 could not be found"
        if [[ -v $installer ]]; then
        exit 1
        fi
        echo "Installing $1"
        eval $installer
    else
        echo "command $1 - exists"
    fi
}

function create_clusters() {
    API_PORT=6444
    PORT=80
    for CLUSTER_NAME in control-plane cluster-1 cluster-2 cluster-3
    do
    desc "creating cluster ${CLUSTER_NAME}"
    k3d cluster create ${CLUSTER_NAME} \
        --image docker.io/rancher/k3s:v1.23.8-k3s2 \
        --api-port "${HOST_IP}:${API_PORT}" \
        --port "${PORT}:80@server:0" \
        --servers-memory 2g \
        --k3s-arg "--disable=traefik@server:0" \
        --network multi-clusters \
        --timeout 120s \
        --wait
        ((API_PORT=API_PORT+1))
        ((PORT=PORT+1))
    done    
}

function install_eriecanal() {
    desc "Adding ErieCanal helm repo"
    helm repo add ec https://ec.flomesh.io --force-update
    helm repo update

    EC_NAMESPACE=erie-canal
    EC_VERSION=0.1.0-beta.2

    for CLUSTER in ${!kubeconfig*}
    do
       CLUSTER_NAME=$(if [ "${CLUSTER}" == "kubeconfig_c1" ]; then echo "cluster-1"; elif [ "${CLUSTER}" == "kubeconfig_c2" ]; then echo "cluster-2"; \
        elif [ "${CLUSTER}" == "kubeconfig_c3" ]; then echo "cluster-3";else echo "control-plane"; fi) 
       desc "installing ErieCanal on cluster ${CLUSTER_NAME}"
       helm upgrade -i --kubeconfig ${!CLUSTER} --namespace ${EC_NAMESPACE} --create-namespace --version=${EC_VERSION} --set ec.logLevel=5 ec ec/erie-canal
       sleep 1
       kubectl --kubeconfig ${!CLUSTER} wait --for=condition=ready pod --all -n $EC_NAMESPACE --timeout=120s
    done
}

function join_clusters() {
    PORT=81
    for CLUSTER_NAME in cluster-1 cluster-2 cluster-3
    do
        desc "Joining ${CLUSTER_NAME}"
        kubectl --kubeconfig ${kubeconfig_cp} apply -f - <<EOF
apiVersion: flomesh.io/v1alpha1
kind: Cluster
metadata:
  name: ${CLUSTER_NAME}
spec:
  gatewayHost: ${HOST_IP}
  gatewayPort: ${PORT}
  kubeconfig: |+
`k3d kubeconfig get ${CLUSTER_NAME} | sed 's|^|    |g' | sed "s|0.0.0.0|$HOST_IP|g"`
EOF
    ((PORT=PORT+1))        
    done
}

function install_osm_edge_binary() {
    release=v1.3.0-beta.4
    desc "downloading osm-edge cli release - ${release}"
    curl -L https://github.com/flomesh-io/osm-edge/releases/download/${release}/osm-edge-${release}-${system}-$arch.tar.gz | tar -vxzf -
    osm_binary="$(pwd)/${system}-${arch}/osm"
    $osm_binary version
}

function install_edge() {
    OSM_NAMESPACE=osm-system
    OSM_MESH_NAME=osm
    for CONFIG in kubeconfig_c1 kubeconfig_c2 kubeconfig_c3
    do
      DNS_SVC_IP="$(kubectl --kubeconfig ${!CONFIG} get svc -n kube-system -l k8s-app=kube-dns -o jsonpath='{.items[0].spec.clusterIP}')"
      CLUSTER_NAME=$(if [ "${CONFIG}" == "kubeconfig_c1" ]; then echo "cluster-1"; elif [ "${CONFIG}" == "kubeconfig_c2" ]; then echo "cluster-2"; else echo "cluster-3"; fi)
      desc "Installing osm-edge service mesh in cluster ${CLUSTER_NAME}"
      KUBECONFIG=${!CONFIG} $osm_binary install \
        --mesh-name "$OSM_MESH_NAME" \
        --osm-namespace "$OSM_NAMESPACE" \
        --set=osm.certificateProvider.kind=tresor \
        --set=osm.image.pullPolicy=Always \
        --set=osm.sidecarLogLevel=error \
        --set=osm.controllerLogLevel=warn \
        --timeout=900s \
        --set=osm.localDNSProxy.enable=true \
        --set=osm.localDNSProxy.primaryUpstreamDNSServerIPAddr="${DNS_SVC_IP}"
    
      kubectl --kubeconfig ${!CONFIG} wait --for=condition=ready pod --all -n $OSM_NAMESPACE --timeout=120s
    done
}

echo "Checking for pre-requiste commands"
# check for docker
check_command "docker"

# check for kubectl
check_command "kubectl"

# check for k3d
check_command "k3d" "curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash"

# check for helm
check_command "helm" "curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash"

# check for pv
check_command "pv" "sudo apt-get install pv -y"

# check for jq
check_command "jq" "sudo apt-get install jq -y"

echo "creating k3d clusters"
create_clusters

k3d kubeconfig get control-plane > "${kubeconfig_cp}"
k3d kubeconfig get cluster-1 > "${kubeconfig_c1}"
k3d kubeconfig get cluster-2 > "${kubeconfig_c2}"
k3d kubeconfig get cluster-3 > "${kubeconfig_c3}"

desc "installing ErieCanal on clusters"
install_eriecanal

desc "Joining clusters into a ClusterSet"
join_clusters

desc "downloading osm-edge cli"
install_osm_edge_binary

desc "installing osm_edge on clusters"
install_edge

echo "Clusters are ready. Proceed with running demo"
