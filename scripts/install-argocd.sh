#!/bin/bash

# Platform check
PLATFORM=`uname -s | awk '{print tolower($0)}'`
echo $PLATFORM 

# Fetch the appropriate OS binary for ArgoCD
curl -sSL -o bin/argocd https://github.com/argoproj/argo-cd/releases/download/v1.8.7/argocd-$PLATFORM-amd64

# Make the argocd CLI executable
chmod +x bin/argocd