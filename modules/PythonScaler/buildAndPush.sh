#! /bin/bash

docker build . -t <private-registry-hostname>/resource-scaler:latest
docker push <private-registry-hostname>/resource-scaler:latest