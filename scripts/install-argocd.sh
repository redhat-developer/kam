#!/bin/bash
set -e

if [[ "$ARGOCD_SERVER" == "" ]]; then
 echo "ArgoCD server address unspecified!"
 exit 1
fi

echo "Removing previous argocd installation from path (if any)..."
rm /usr/local/bin/argocd

echo "Starting argocd installation..."

# Not considering the Windows binary for now
OS_NAME=$(uname -s)

# Find the latest ArgoCD version
VERSION=$(curl --silent "https://api.github.com/repos/argoproj/argo-cd/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')

# Fetch the appropriate OS binary for ArgoCD

curl -sSL -o /usr/local/bin/argocd https://github.com/argoproj/argo-cd/releases/download/$VERSION/argocd-$OS_NAME-amd64

# Make the argocd CLI executable

chmod +x /usr/local/bin/argocd

# Simple check to see if Argocd binary works
argocd version