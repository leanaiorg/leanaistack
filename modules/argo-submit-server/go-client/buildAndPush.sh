#! /bin/bash

docker build . -t <private-registry-hostname>/argo-go-cli:latest
docker push <private-registry-hostname>/argo-go-cli:latest