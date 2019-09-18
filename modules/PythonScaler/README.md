1. Get your configs into the docker container
 cp ~/.local/share/juju/* configs/.
2. Build and deploy
    ./buildAndDeploy
3. Run the helm chart
    helm install .charts --name python-scaler -f .charts/values.yaml



########################################################
For now this is a simple webbased API for scaling up/down resoruces. Improvents should be made further on in regards of stability and connections.

The API will add resources to our cluster in terms of a "unit". A unit is a package of a virtual machine, etcd connections, flannel connections and containerd connections. We can (for now) pass constraints to our scaler, where we say the minimum CPU and Memory resources we need. 

Further on we should connect this to our (rancher) prometheus endpoint so that we can scale according to HTTP-requests/memory utilization/CPU, etc. Will perhaps come in a later stage
########################################################

Scaling up:
    curl https://<url>/scaleup/<cpu[cores]>/<memory[GB]>

    Example: curl https://scaler.coolcloud.fake.se/scaleup/2/2

    When we scale up, a name to the unit will be returned, in the format of: kubernetes-worker/<num>. That num, can be used later on to scale down the worker. 

Scaledown: 
    curl https://<url>/scaledown/<num>
    
    Example: curl https://scaler.coolcloud.fake.se/scaledown/4

    When we scale down, we pass the unit number. The scaledown will remove the kubernetes worker, etcd connection, flannel connection and the containerd integration. 

    Once the application connections/integrations are removed the virtual machine will be removed. 

    IMPORTANT TODO: Pause the machine before removing to migrate the current workload.

Status: 
    curl https://<urk>/status 

    Example: curl https://scaler.coolcloud.fake.se/status

    Returns information of the units running in the cluster and on what machines the units are deployed on.