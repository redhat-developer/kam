#!/bin/bash

# Platform check
PLATFORM=`uname`_`uname -m`

PWD=`pwd`

echo $PWD

# Fetch the appropriate OS binary for glab
wget -O bin/glab_1.22.0_$PLATFORM.tar.gz  https://github.com/profclems/glab/releases/download/v1.22.0/glab_1.22.0_$PLATFORM.tar.gz

tar -xvzf bin/glab_1.22.0_$PLATFORM.tar.gz

# Make the glab CLI executable
chmod +x bin/glab

# Remove the tar file
rm bin/glab_1.22.0_$PLATFORM.tar.gz