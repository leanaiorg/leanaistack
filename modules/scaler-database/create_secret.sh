#!/bin/bash
PASSWD=`echo -n <root-mysql-password>|base64`
USERNAME=`echo -n root|base64`

echo "apiVersion: v1
kind: Secret
metadata:
  name: scaler-db-secret
type: Opaque
data:
  mysqlUser: ${USERNAME}
  mysqlPassword: ${PASSWD}"
