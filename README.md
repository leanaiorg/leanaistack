# Lean AI Stack

## Introduction

## Documentation
[Docs](/docs/) folder

## 1. Install cluster


[Infrastructure Setup Guide](/infrastructure/)
## 2. Setup prerequisites

### 2.1 Generate a wildcard domain certificate.
Follow along in the readme in [Cert](extra/cert)

## 3. Install charts

### 3.1 Prerequisites
1. Ensure you have a cluster ready. From instructions above or other.
2. Ensure you have a loaded `$KUBECONFIG` from env or other place.
1. Ensure you have installed and configured **helm**, check that `helm version` shows also the server version and you are ready to go!

### 3.2 Refresh dependencies
To refresh dependencies before installing run the following command from this ``"root"`` directory
```bash
helm dep up
```

### 3.3 Copy example values to your local.
```bash
cp values.yaml values-local.yaml
```
Edit as appropriate.

### 3.4 Install charts
from `"root"` directory and override values with your values file.
```bash
helm install -n leanai . --values=values-local.yaml
```

### Upgrade only values that changed.
```bash
helm install --upgrade . -n leanai --values=values-local.yaml
```

### Uninstall charts
```bash
helm delete --purge leanai
```
## Deploy Examples
try out the workflow engine with `hello-world.yaml`from [Examples](https://github.com/leanaiorg/examples).

## Where to go from here

Create docker secret based on your credentials for pulling images from private repos.
```bash
kubectl create secret docker-registry regcred --docker-server=<your-registry-server> --docker-username=<your-name> --docker-password=<your-pword> --docker-email=<your-email>
```
