#!/bin/bash

# Find the latest ArgoCD version
VERSION=$(curl --silent "https://api.github.com/repos/argoproj/argo-cd/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
echo $VERSION

# Platform check
PLATFORM=`uname -s | awk '{print tolower($0)}'`
echo $PLATFORM 

# Fetch the appropriate OS binary for ArgoCD
curl -sSL -o bin/argocd https://github.com/argoproj/argo-cd/releases/download/$VERSION/argocd-$PLATFORM-amd64

# Make the argocd CLI executable
chmod +x argocd