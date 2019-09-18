----- SETUP Scaler Database -----
1. Navigate to modules/scaler-database
2. Create namespace for submitserver/scaler
    kubectl create namespace submit-scaler
    kubectl create -f cert.yaml -n submit-scaler (FROM WHERE THE CERT IS STORED CREATED WHEN CREATING INFRA)
    kubectl create secret docker-registry regcred3 --docker-server=https://<registry-url> --docker-username=<username> --docker-password=<password> -n submit-scaler
3. Edit values.yaml
4. helm install --name scaler-db stable/mysql --namespace=submit-scaler -f values.yaml
5. Seed database:
    kubectl exec -it -n submit-scaler scaler-db-mysql-85996f7dbf-v8mjl /bin/bash

     mysql -u root -p

     And copy-paste the contents of DB_Def.sql

     Last: Create a protected machine by:
     insert into ClusterInfo (machineId, unitId, nodeName, removable, active) VALUES  ("17","kubernetes-worker/6", "	juju-37883b-default-17", False, True)

6. Enter username/password from values.yaml in the create_secret.sh and run:
    ./create_secret.sh | kubectl create -f- -n submit-scaler

----- SETUP Python Scaler -----
1. Navigate to modules/PythonScaler
2. Navigate to modules/PythonScaler/configs and run:
    cp ~/.local/share/juju/* .
3. Edit and create secrets by:
    3.1 Edit create_secret.sh with
        JUJU_CONTROLLER_ENDPOINT -- From juju gui
        JUJU_USERNAME -- From juju gui
        JUJU_PASSWORD -- From juju gui
    3.2 juju show-controller
        Extract the certificate data and put it in 'tls.ca' (modules/PythonScaler/tls.ca)
    3.3 Then create the secret by:
        ./create_secret.sh | kubectl create -f- -n submit-scaler
4. Run: buildAndPush.sh
5. Edit values.yaml under .charts
6. Deploy the scaler-db
    helm install .charts --name python-scaler -f .charts/values.yaml --namespace=submit-scaler

----- SETUP WORK QUEUE -----
1. Navigate to modules/argo-submit-server/workqueue
2. Edit values.yaml
3. helm install --name rabbit-mq stable/rabbitmq -f values.yaml --namespace=submit-scaler
4. Enter username/password from values.yaml in the create_secret.sh and run:
    ./create_secret.sh | kubectl create -f- -n submit-scaler

----- SETUP ARGO SUBMIT SERVER -----
1. Navigate to: modules/argo-submit-server/go-client
2. Edit argo-cli.go constant
3. Copy your kube config to argo
     cp ~/.kube/config config
3. run ./buildAndPush.sh
4. navigate to modules/argo-submit-server/.charts
    edit values.yaml for correct values
5. helm install .charts --name argo-submit -f .charts/values.yaml --namespace=submit-scaler
6. kubectl create sa argo-events-sa -n submit-scaler
7. ./create_secret.sh | kubectl create -f- -n submit-scaler (from minio)






#####################################################################################
-------------------------------------------------------------------------------------

        NOTES

-------------------------------------------------------------------------------------
#####################################################################################

### TO DUPLICATE THE SECRETS
kubectl get secret minio-credentials --namespace=default --export -o yaml | kubectl apply --namespace=submit-scaler -f -


### CREATE BACKUPS:
juju create-backup
downloading to juju-backup-20170204-105651.tar.gz
juju list-backups
20170204-105651.7f2bca6b-2505-4cf5-886c-6320ca67dc76
juju download-backup 20170204-105651.7f2bca6b-2505-4cf5-886c-6320ca67dc76

## RESTORE FROM BACKUP:
juju restore-backup  -b --file=backup.tar.gz
juju enable-ha

###BUGS AND FIXES:

1. If you have issues with: "Stopped services: kube-controller-manager"

try:
    juju run --application kubernetes-master 'service snap.kube-apiserver.daemon restart'
If that doesnt work try:
    juju run --application etcd 'service snap.etcd.etcd restart'
    juju run --application kubernetes-master 'service snap.kube-apiserver.daemon restart'
    juju run --application kubernetes-master 'service snap.kube-controller-manager.daemon restart'
    juju run --application kubernetes-master 'service snap.kube-scheduler.daemon restart'
    juju run --application kubernetes-worker 'service snap.kubelet.daemon restart'
    juju run --application kubernetes-worker 'service snap.kube-proxy.daemon restart'

2. Integrator causes master/worker to fail:
    Edit: /var/snap/kubelet/current/args
    and change to this:  --cloud-provider="external"
    cat /var/snap/kubelet/current/args

    --cloud-config="/var/snap/kubelet/common/cloud-config.conf"
    --cloud-provider="external"
    --config="/root/cdk/kubelet/config.yaml"
    --container-runtime="docker"
    --dynamic-config-dir="/root/cdk/kubelet/dynamic-config"
    --config="/root/cdk/kubelet/config.yaml"
    --kubeconfig="/root/cdk/kubeconfig"
    --logtostderr
    --network-plugin="cni"
    --node-ip="192.168.100.6"
    --pod-infra-container-image="image-registry.canonical.com:5000/cdk/pause-amd64:3.1"
    --v="0"


    DO THE SAME ON THE MASTER:
    cat /var/snap/kube-controller-manager/current/args

    --cloud-config="/var/snap/kube-controller-manager/common/cloud-config.conf"
    --cloud-provider="external"
    --min-resync-period="3m"
    --tls-cert-file="/root/cdk/server.crt"
    --tls-private-key-file="/root/cdk/server.key"
    --root-ca-file="/root/cdk/ca.crt"
    --service-account-private-key-file="/root/cdk/serviceaccount.key"
    --master="http://127.0.0.1:8080"
    --logtostderr
    --v="2"


3. ArgoSubmitServer crashes due to:
EASON:   Unschedulable
STATUS:   False
MESSAGE:   pod has unbound immediate PersistentVolumeClaims
0
panic: runtime error: index out of range

goroutine 54 [running]:
main.printUnscheduledPods(0x0, 0x0, 0x0, 0x0, 0xc00033a124, 0x7, 0x0, 0x0, 0xc00033a139, 0x7, ...)
	/go/ScalingOperations.go:143 +0x9bc
main.checkForUnscheduledPods()
	/go/ScalingOperations.go:125 +0x1db
created by main.autoscaler
	/go/ScalingOperations.go:34 +0x1fc
exit status 2



Fix asap

4. If pvc doesnt work with openstack-integrator
/var/snap/kube-apiserver/current/args  and set:
--cloud-provider="external"


##### NOTES #####
Take a look at: juju deploy cs:glance-simplestreams-sync-23 --trust
