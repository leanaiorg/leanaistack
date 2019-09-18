#!/bin/bash
JUJU_CONTROLLER_ENDPOINT=`echo -n yourhost|base64`
JUJU_USERNAME=`echo -n youruser|base64`
JUJU_PASSWORD=`echo -n yourpassword|base64`
JUJU_CACERT=`cat yourcert.ca | base64`

echo "apiVersion: v1
kind: Secret
metadata:
  name: python-scaler-secret
type: Opaque
data:
  controller_endpoint: ${JUJU_CONTROLLER_ENDPOINT}
  username: ${JUJU_USERNAME}
  password: ${JUJU_PASSWORD}
  cacert: ${JUJU_CACERT}"
