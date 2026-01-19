#!/bin/bash
set -euo pipefail

echo "Generating self-signed certificates..."

# Generate private key
openssl genrsa -out server.key 2048

# Generate certificate signing request (CSR)
# Note: Common Name (CN) is set to the IP address for basic validity
openssl req -new -key server.key -out server.csr -subj "/C=US/ST=State/L=City/O=ThousandWorlds/CN=10.0.0.17"

# Generate self-signed certificate (valid for 365 days)
# Using x509 v3 extension to include SAN (Subject Alternative Name) for the IP
echo "subjectAltName=IP:10.0.0.17" > extfile.cnf
openssl x509 -req -days 365 -in server.csr -signkey server.key -out server.crt -extfile extfile.cnf

echo "Creating Kubernetes Secret..."
# Create the mud-world namespace if it doesn't exist (to prevent errors)
kubectl create namespace mud-world --dry-run=client -o yaml | kubectl apply -f -

# delete existing secret if any
kubectl -n mud-world delete secret nginx-certs --ignore-not-found

# Create tls secret
kubectl -n mud-world create secret tls nginx-certs --key server.key --cert server.crt

echo "Cleaning up local files..."
rm server.key server.csr server.crt extfile.cnf

echo "Done! Secret 'nginx-certs' created in 'mud-world' namespace."
