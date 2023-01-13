#!/bin/bash

cd $(dirname ${BASH_SOURCE})
. ./common.sh

set -e

echo "cleaning up"
for cluster in control-plane cluster-1 cluster-2 cluster-3
do
    echo "deleting cluster ${cluster}"
    k3d cluster delete ${cluster}
done

for config in ${!kubeconfig*}
do
  rm -f ${!config}
done

rm -rf "./${system}-${arch}"