#!/bin/bash

cd $(dirname ${BASH_SOURCE})
. ./common.sh

${k0} delete ns --ignore-not-found=true httpbin curl
${k1} delete ns --ignore-not-found=true httpbin curl
${k2} delete ns --ignore-not-found=true httpbin curl
${k3} delete ns --ignore-not-found=true httpbin curl
