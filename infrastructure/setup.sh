#!/bin/bash

echo "Building deploy tool"
#docker build -t leanai-juju:latest .
#docker pull leanai-juju:latest
echo "\n\n\n"
echo "Running deploy tool"
docker run -v $(pwd):/home/ubuntu -it leanaiorg/deployer:latest ./in-docker-setup.sh "$@"
