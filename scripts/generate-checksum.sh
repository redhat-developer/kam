#!/bin/sh

# fail if some commands fails
set -e

# show commands
set -x

checksum_file="dist/kam_checksums.txt"
if test -f $checksum_file ; then
    echo -n "" > $checksum_file
fi
for file in $(ls dist/*) ; do
    if [[ "${file: -3}" != "txt" ]] ; then
        sha256sum "${file}" | sed 's/dist\///' >> $checksum_file ;
    fi
done