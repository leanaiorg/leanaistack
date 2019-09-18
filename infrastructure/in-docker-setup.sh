#!/bin/bash

#NAME="safespring_qa"
#METADATA="$(pwd)/streams"
#NETWORK_ID="d6ccc707-cacb-42a5-b547-f1f151869185"
#EXTERNAL_NETWORK_ID="71b10496-2617-47ae-abbc-36239f0863bb"
IMAGE_ID="9ddfcfd5-78bf-41c3-acb3-7f87216cb311"
#METADATA=$(pwd)/${NEW_UUID}
ENDPOINT="https://hpc2n.cloud.snic.se:5000/v3"
REGION="HPC2N"

for ARGUMENT in "$@"
do

    KEY=$(echo $ARGUMENT | cut -f1 -d=)
    VALUE=$(echo $ARGUMENT | cut -f2 -d=)

    case "$KEY" in
            --name)              NAME=${VALUE} ;;
            --metadata_dir)          METADATA=${VALUE} ;;
            --network_id)          NETWORK_ID=${VALUE} ;;
            --external_network_id)          EXTERNAL_NETWORK_ID=${VALUE} ;;
            --image_id)          IMAGE_ID=${VALUE} ;;
            --endpoint)          ENDPOINT=${VALUE};;
            --region)            REGION=${VALUE};;
            *)
    esac


done

juju add-cloud $NAME -f providers/$NAME/cloud.yaml
juju add-credential $NAME -f providers/$NAME/credentials.yaml

mkdir $METADATA
juju metadata generate-image -d $METADATA -i $IMAGE_ID -s bionic -r $REGION -u $ENDPOINT
juju bootstrap $NAME $NAME-controller --config network=$NETWORK_ID --config external-network=$EXTERNAL_NETWORK_ID --config use-floating-ip=true --metadata-source ${METADATA} --bootstrap-constraints "mem=4G cores=2"
juju model-config enable-os-refresh-update=false
juju model-config enable-os-upgrade=false
juju set-default-credential $NAME $NAME-cred
juju deploy juju-implementation/cluster-defs/bundle_docker_ceph_flannel.yaml
juju add-storage ceph-osd/0 osd-devices=cinder,100G,1
juju add-storage ceph-osd/1 osd-devices=cinder,100G,1
juju add-storage ceph-osd/2 osd-devices=cinder,100G,1
juju config kubernetes-worker ingress=true
juju expose kubernetes-worker
juju config kubernetes-master allow-privileged=true
# TODO - try loading in a loop and after success return
#

echo "SUCCESS! WATCHING PROGRESS"
echo "WHEN READY RUN:"
echo "juju scp kubernetes-master/0:config juju_config"
echo "watch -c juju status --color"
juju gui > juju-admin-credentials.txt
/bin/sh
