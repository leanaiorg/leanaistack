# Lean AI Stack Infrastructure setup

## Prerequisites
1. Installed docker
2. Clone of this repository

### Prepare a provider directory
1. Copy the examples folder from providers/
```bash
cp -r providers/examples providers/yourcloud
```
2. Edit the `providers/yourcloud/clouds.yaml`
> Make settings as applicable to your cloud and cloud provider.

3. Edit the `providers/yourcloud/credentials.yaml`
> Make settings as applicable to your cloud and cloud provider.

### Find parameters for appropriate settings.
In example below taken from OpenStack.
Find your parameters with for example the python openstack client.

## Run the setup
### 1. Start cloud creation using juju
```
./setup.sh --name=safespring_qa --metadata_dir=metadata --network_id=9a2cb619-ed9c-407e-9b94-7d4d8129d80b --external_network_id=71b10496-2617-47ae-abbc-36239f0863bb --image_id=9ddfcfd5-78bf-41c3-acb3-7f87216cb311
```
> In case you need to build the image required then you run `docker build -t leanai-juju:latest .` prior to invoking setup.
### 2. Wait for the setup to show all resources created.

### 3. Copy the config
 Copy and source the config file to local directory and your kubernetes-cluster is ready to connect to!


## Deploy prerequired services

### Create certificate
1. Clone the repo
```bash
git clone https://github.com/certbot/certbot.git
```
2.
```bash
./certbot-auto certonly --manual --preferred-challenges=dns --email YOUR@EMAIL.HERE -d *.YOURDOMAIN.NAME
```
3. Copy the fullchain and a preferred location as tls.crt and tls.key

### Create secret
```bash
kubectl -n cattle-system create secret tls cattle-keys-ingress --cert=tls.crt --key=tls.key
```
### Deploy Rancher
1. Setup your preferred address for rancher dashboard based on your selected wildcard ceriticate.
```bash
kubectl apply -f infrastructure/juju-implementation/cluster-defs/deployment.yaml
```



### Deploy Helm

```bash
kubectl -n kube-system create serviceaccount tiller

kubectl create clusterrolebinding tiller \
  --clusterrole=cluster-admin \
  --serviceaccount=kube-system:tiller

helm init --service-account tiller
```
