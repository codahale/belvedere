#!/bin/bash
set -euo pipefail
IFS=$'\n\t'

# Generate a self-signed TLS certificate.
mkdir -p /etc/tls/
self-signed-tls -c=US -s=XX -l=XX -o=XX -u=XX -n=tmp -e=robots@example.com -p=/etc/tls/ -d=1000

# Run Nginx.
exec nginx -c helloworld.nginx
