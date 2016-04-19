#!/usr/bin/env bash
cd `dirname $0`/.. || exit 1;
projectpath=`pwd`
export GOPATH="${projectpath}:$GOPATH"
export GOBIN="${projectpath}/bin"
export PATH=$PATH:$projectpath

if [ -x swallow ]; then
    echo "No swallow found."
    exit 2
fi

gobin=`which go`
exec ${gobin} install  -ldflags "-s -w" ./...
#exec ${gobin} install -v -gcflags "-N -l"  ./...
# ps -ef|grep "League"
# kill -9
echo "go install project done..." $(date +"%Y-%m-%d %H:%M:%S")