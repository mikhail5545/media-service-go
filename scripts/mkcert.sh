#!/usr/bin/env bash

#
# Copyright (c) 2026. Mikhail Kulik
#
# This program is free software: you can redistribute it and/or modify
#  it under the terms of the GNU Affero General Public License as published
#  by the Free Software Foundation, either version 3 of the License, or
#  (at your option) any later version.
#
# This program is distributed in the hope that it will be useful,
#  but WITHOUT ANY WARRANTY; without even the implied warranty of
#  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
#  GNU Affero General Public License for more details.
#
# You should have received a copy of the GNU Affero General Public License
#  along with this program.  If not, see <https://www.gnu.org/licenses/>.
#

set -e

if ! [ -d "./certs" ]; then
  mkdir ./certs
fi

cd ./certs

# CA
openssl genrsa -out ca.key 4096
openssl req -x509 -new -nodes -key ca.key -sha256 -days 3650 -subj "/CN=Local Dev CA" -out ca.pem

# server ext file
if ! [ -f "server_ext.cnf" ]; then
  cat > server_ext.cnf << EOF
req_extensions = v3_req
distinguished_name = req_distinguished_name
[req_distinguished_name]
[ v3_req ]
subjectAltName = @alt_names
[alt_names]
DNS.1 = localhost
DNS.2 = host.docker.internal
DNS.3 = product-server
IP.1  = 127.0.0.1
EOF
fi

# Server cert
openssl genrsa -out server.key 2048
openssl req -new -key server.key -out server.csr -subj "/CN=localhost" -config server_ext.cnf
openssl x509 -req -in server.csr -CA ca.pem -CAkey ca.key -CAcreateserial -out server.crt -days 365 -sha256 -extfile server_ext.cnf -extensions v3_req

openssl genrsa -out client.key 2048
openssl req -new -key client.key -out client.csr -subj "/CN=client1"
openssl x509 -req -in client.csr -CA ca.pem -CAkey ca.key -CAcreateserial -out client.crt -days 365 -sha256

echo "Created certs in $(pwd). Mount them into containers with -v $(pwd):/certs:ro"