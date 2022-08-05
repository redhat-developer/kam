#!/bin/bash

GLAB_VERSION="1.22.0"
PLATFORM=`uname`

# Platform check
if [ $PLATFORM == "Darwin" ]; then
    PLATFORM="macOS"_`uname -m`
else
    PLATFORM=`uname`_`uname -m`
fi

echo $PLATFORM

# Fetch the appropriate OS binary for glab
wget -O bin/glab_${GLAB_VERSION}_$PLATFORM.tar.gz  https://github.com/profclems/glab/releases/download/v${GLAB_VERSION}/glab_${GLAB_VERSION}_$PLATFORM.tar.gz

tar -xvzf bin/glab_${GLAB_VERSION}_$PLATFORM.tar.gz

# Make the glab CLI executable
chmod +x bin/glab

# Remove the tar file
rm bin/glab_${GLAB_VERSION}_$PLATFORM.tar.gz