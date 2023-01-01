#!/bin/bash

#
# MIT License
#
# Copyright (c) since 2021,  flomesh.io Authors.
#
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included in all
# copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
# SOFTWARE.
#

openssl genrsa -out ca.key 4096

openssl req -new -x509 -nodes -days 365000 \
   -key ca.key \
   -out ca.crt \
   -subj '/CN=flomesh.io'

openssl genrsa -out ingress-pipy.key 4096
openssl req -new -key ingress-pipy.key -out ingress-pipy.csr -subj '/CN=fsm-ingress-pipy-controller.flomesh'
openssl x509 -req -in ingress-pipy.csr -CA ca.crt -CAkey ca.key -extfile extfile.cnf -CAcreateserial -out ingress-pipy.crt -days 365000

openssl genrsa -out client.key 4096
openssl req -new -key client.key -out client.csr -subj '/CN=client.flomesh'
openssl x509 -req -in client.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out client.crt -days 365000