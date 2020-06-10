#!/bin/bash

# This script deploys the gogsfilter web server to the dev project.

set -e

cd $(dirname $0)
BASE=$(pwd)
cd - >> /dev/null

oc new-project dev || oc project dev

oc new-app \
  --name=gogsfilter \
  --binary \
  -n dev \
  --build-env=IMPORT_URL=. \
  --build-env=INSTALL_URL=github.com/kwkoo/gogsfilter/cmd/gogsfilter \
  --docker-image=docker.io/centos/go-toolset-7-centos7:latest \
  --env=RULESJSON='[{"ref":"refs/heads/master","target":"http://el-go-pipeline.dev.svc.cluster.local:8080"}]'

echo -n "Waiting for imagestreamtag to appear..."

set +e

while true; do
  oc get -n dev istag/go-toolset-7-centos7:latest > /dev/null 2>&1
  if [ $? -eq 0 ]; then
    echo "done"
    break
  fi
  echo -n "."
  sleep 1
done

set -e

oc start-build -n dev gogsfilter \
  --follow \
  --from-dir=${BASE}/src

oc expose -n dev dc/gogsfilter --port 8080

echo "gogsfilter is now available at http://gogsfilter.dev.svc.cluster.local:8080"