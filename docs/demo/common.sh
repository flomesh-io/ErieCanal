#!/bin/bash

HOST_IP=$(if [ "$(uname)" == "Darwin" ]; then ipconfig getifaddr en0; else ip -o route get to 8.8.8.8 | sed -n 's/.*src \([0-9.]\+\).*/\1/p'; fi) 
kubeconfig_cp=${KUBECONFIG_CP:-"cp.kubeconfig"}
kubeconfig_c1=${KUBECONFIG_C1:-"c1.kubeconfig"}
kubeconfig_c2=${KUBECONFIG_C2:-"c2.kubeconfig"}
kubeconfig_c3=${KUBECONFIG_C3:-"c3.kubeconfig"}

system=$(uname -s | tr [:upper:] [:lower:])
arch=$(if [ "$(uname)" == "Darwin" ]; then uname -m; else dpkg --print-architecture; fi)
osm_binary="$(pwd)/${system}-${arch}/osm"

k0="kubectl --kubeconfig ${kubeconfig_cp}"
k1="kubectl --kubeconfig ${kubeconfig_c1}"
k2="kubectl --kubeconfig ${kubeconfig_c2}"
k3="kubectl --kubeconfig ${kubeconfig_c3}"

readonly  reset=$(tput sgr0)
readonly  green=$(tput bold; tput setaf 2)
readonly yellow=$(tput bold; tput setaf 3)
readonly   blue=$(tput bold; tput setaf 6)
readonly timeout=$(if [ "$(uname)" == "Darwin" ]; then echo "1"; else echo "0.1"; fi) 

function desc() {
    maybe_first_prompt
    echo "$blue# $@$reset"
    prompt
}

function prompt() {
    echo -n "$yellow\$ $reset"
}

started=""
function maybe_first_prompt() {
    if [ -z "$started" ]; then
        prompt
        started=true
    fi
}

# After a `run` this variable will hold the stdout of the command that was run.
# If the command was interactive, this will likely be garbage.
DEMO_RUN_STDOUT=""

function run() {
    maybe_first_prompt
    rate=250
    if [ -n "$DEMO_RUN_FAST" ]; then
      rate=1000
    fi
    echo "$green$1$reset" | pv -qL $rate
    if [ -n "$DEMO_RUN_FAST" ]; then
      sleep 0.5
    fi
    OFILE="$(mktemp -t $(basename $0).XXXXXX)"
    script -eq -c "$1" -f "$OFILE"
    r=$?
    read -d '' -t "${timeout}" -n 10000 # clear stdin
    prompt
    if [ -z "$DEMO_AUTO_RUN" ]; then
      read -s
    fi
    DEMO_RUN_STDOUT="$(tail -n +2 $OFILE | sed 's/\r//g')"
    return $r
}

function relative() {
    for arg; do
        echo "$(realpath $(dirname $(which $0)))/$arg" | sed "s|$(realpath $(pwd))|.|"
    done
}

#SSH_NODE=$(kubectl get nodes | tail -1 | cut -f1 -d' ')

trap "echo" EXIT
