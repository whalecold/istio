#!/bin/bash

#
# Copyright MSE Authors. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
################################################################################
#
# Script to configure and start the Istio sidecar.

mkdir -p /var/lib/istio/envoy
mkdir -p /etc/istio/proxy
mkdir -p /etc/istio/config
mkdir -p /etc/certs
mkdir -p /var/run/secrets/tokens
mkdir -p /var/run/secrets/workload-spiffe-uds
mkdir -p /var/log/istio

cat /tmp/istio/mesh > /etc/istio/config/mesh
cat /tmp/istio/cluster.env > /var/lib/istio/envoy/cluster.env
cat /tmp/istio/root-cert.pem > /etc/certs/root-cert.pem
cat /tmp/istio/istio-token > /var/run/secrets/tokens/istio-token
cat /opt/istio-proxy/envoy_bootstrap_tmpl.json > /var/lib/istio/envoy/envoy_bootstrap_tmpl.json

#chown -R 1337  /usr/local/bin/envoy
chown -R 1337:1337 /var/lib/istio /etc/istio /etc/certs /var/run/secrets/tokens /var/run/secrets/workload-spiffe-uds /var/log/istio
