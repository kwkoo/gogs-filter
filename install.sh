#!/bin/bash

# This script deploys the gogsfilter web server to the dev project.
PROJ=dev

set -e

cd $(dirname $0)
BASE=$(pwd)
cd - >> /dev/null

oc new-project $PROJ || oc project $PROJ

oc new-app \
  --name=gogsfilter \
  --binary \
  -n $PROJ \
  --docker-image=ghcr.io/kwkoo/go-toolset-7-centos7:1.15.2 \
  --env=RULESJSON='[{"ref":"refs/heads/master","target":"http://el-go-pipeline.dev.svc.cluster.local:8080"}]'

echo -n "Waiting for imagestreamtag to appear..."

set +e

while true; do
  oc get -n $PROJ istag/go-toolset-7-centos7:1.15.2 > /dev/null 2>&1
  if [ $? -eq 0 ]; then
    echo "done"
    break
  fi
  echo -n "."
  sleep 1
done

set -e

oc start-build gogsfilter \
  -n $PROJ \
  --follow \
  --from-dir=${BASE}/src

oc expose -n $PROJ deploy/gogsfilter --port 8080

echo "gogsfilter is now available at http://gogsfilter.${PROJ}.svc.cluster.local:8080"