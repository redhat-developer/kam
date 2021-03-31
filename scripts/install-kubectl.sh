#!/bin/bash

# Platform check
PLATFORM=`uname -s | awk '{print tolower($0)}'`
echo "Your platform is $PLATFORM" 

# Download and install kubectl
echo -e "\nGet kubectl binary\n"
curl -sSL -o bin/kubectl https://storage.googleapis.com/kubernetes-release/release/v1.20.5/bin/$PLATFORM/amd64/kubectl

# Make the kubectl executable 
chmod +x bin/kubectl
