PASSWD=`echo -n <rabbit-mq-password>|base64`
USERNAME=`echo -n <rabbit-mq-username>|base64`

echo "apiVersion: v1
kind: Secret
metadata:
  name: workqueue-secret
type: Opaque
data:
  <rabbit-mq-username>: ${USERNAME}
  <rabbit-mq-password>: ${PASSWD}"
