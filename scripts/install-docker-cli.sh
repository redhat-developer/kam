#!/bin/bash

# Platform check
PLATFORM=`uname -s | awk '{print tolower($0)}'`
echo "Your platform is $PLATFORM" 

DOCKER_VERSION=$(curl --silent "https://api.github.com/repos/docker/docker-ce/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/' | cut -c2-9)
DOCKER_LATEST_BINARY="docker-$VERSION.tgz"

if [ $PLATFORM == "darwin" ]
then
    curl -sSL -o bin/docker https://download.docker.com/mac/static/stable/x86_64/$DOCKER_BINARY
elif [ $PLATFORM == "linux" ]
then
    curl -sSL -o bin/docker https://download.docker.com/linux/static/stable/x86_64/$DOCKER_BINARY
else
    echo "Your OS is unsupported for installing required dependency."
fi

# Make the docker CLI executable
chmod +x bin/docker