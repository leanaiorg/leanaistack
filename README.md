# Lean AI Stack

## Introduction
The Lean AI Stack is a open source project aiming to be a complete solution for working on End to End machine learning. From experiments and exploring datasets to large-scale training to end user serving and monitoring of models and their performance in production. In addition to supporting the basic workflow of finding good models it also support extensive customization and adding value through setting up automated machine learning pipelines. This solution builds on best of breed open source software to provide a complete but totally configurable solution for organizational, institutional or individual needs. We are happy to have your support on this project. See Contributing.

As this project is still in early development expect to see a high paced iteration and added functionality in the coming months. For additional feature requests or bugs please use the issue ticketing system. A rough roadmap is still in development but your input is very welcome. See Contributing.

## Documentation
We aim to provide documentation for the general case setup. If your are a moderate to advanced user and/or have already a K8s cluster running or other prereqs already met you can skip steps below and start on the step corresponding to your current state.

### License
Read more about the license [here](LICENSE)

## How to contribute
Read more on how to contribute [here](CONTRIBUTING.md)

## Built in collaboration with
[![NeiC](https://user-images.githubusercontent.com/2098408/65333320-91eddf00-dbc0-11e9-8bfb-3c9774b62af2.png)](https://github.com/neicnordic)


[![Scaleout](https://user-images.githubusercontent.com/2098408/65333699-42f47980-dbc1-11e9-9db3-f0e5dcdadc8b.png)](www.scaleoutsystems.com)

[![uu_logo](https://user-images.githubusercontent.com/46811/65514759-439d5080-dede-11e9-8389-b22cffffd892.png)](http://www.farmbio.uu.se)

[![pharmbio-logo](https://user-images.githubusercontent.com/46811/65514764-46984100-dede-11e9-9c1d-834d11b82816.png)](https://pharmb.io)



# Setup
## 1. Install cluster
Follow the guide to setup the reference cluster infrastructure.
[Infrastructure Setup Guide](/infrastructure/)

## 2. Setup prerequisites

### 2.1 Generate a wildcard domain certificate.
Follow along in the readme in [Cert](extra/cert)



## 3. Install charts

### 3.1 Prerequisites
1. Ensure you have a cluster ready. From instructions above or other.
2. Ensure you have a loaded `$KUBECONFIG` from env or other place.
1. Ensure you have installed and configured **helm**, check that `helm version` shows also the server version and you are ready to go!
### 3.2 Add helm repository access to published charts
```bash
helm repo add leanai https://leanaiorg.github.io/leanaistack/helm-charts/leanaistack/
helm repo update
```
### 3.3 Refresh dependencies
To refresh dependencies before installing run the following command from this ``"root"`` directory
```bash
helm dep up
```

### 3.4 Copy example values to your local.
```bash
cp values.yaml values-local.yaml
```
Edit as appropriate.

### 3.5 Install charts
from `"root"` directory and override values with your values file.
```bash
helm install leanai/leanaistack -n leanai  --values=values-local.yaml
```

### Upgrade only values that changed.
```bash
helm install --upgrade leanai/leanaistack -n leanai --values=values-local.yaml
```

### Uninstall charts
```bash
helm delete --purge leanai
```
# Stack Components

### Experiments and collaboration
JupyterHub is provided as a hub for your experiments and collaboration.

### Storage
#### S3 Compatible storage
Minio is provided as a S3 compatible storage backend for your datasets and files.

#### Dynamic storage provider
The default cluster sets up a dynamic storage provisioner that can be utilized for your services and workflows to store datasets and files.

#### Docker Registry
Docker registry provides a storage location for your docker image harbouring needs.

### Workflow and pipelines
#### Workflow Engine
The workflow engine powered by Argoproj enables versatile workflow definitions to complete arbitrary tasks. In examples there will be applied usages of workflows for ML/AI.

#### Signals
The signals system powered by Argoproj enables versatile extensions and allow for sensor and triggering events customizable to allow for event-action coupling of workflows. In examples there will be applied usages of eventing and worklows for ML/AI.

### Serving models
For serving models the OpenFaaS project solution is used that can serve models packaged as docker containers and can scale up based on usage or scale to zero on long periods of non use.





### Stack metacomponents
#### Security
The provided examples are meant to be run in already secured environments as this solution is experimental at the moment. There are however configuration options to allow for adding basic-auth protection to services and allow for TLS wrapped communication.
> The user of this open source project is fully aware that this project comes with absolutely no warranty or insurance.

### Ingress
Ingress is provided if required and can be configured. See example `values.yaml`.

# Where to go from here

## Deploy Examples
Several examples are in the making and adopting from real world applications. The examples archive you can find here and will be added to continiously:
On example is to try out the basic workflow engine with `hello-world.yaml`from [Examples](https://github.com/leanaiorg/examples).


### Additional steps
Create docker secret based on your credentials for pulling images from private repos.
```bash
kubectl create secret docker-registry regcred --docker-server=<your-registry-server> --docker-username=<your-name> --docker-password=<your-pword> --docker-email=<your-email>
```
