Argo submit is a workflow server written in go.

The workload server allowes you to POST a config to the server and it will submit a job with the posted configs.


##Example: 
    #Submnit a job from config.json
    curl  https://<argo-subit-hostname>/submit -X POST -d "@config.json"

    #Read and wait for the job with the name test-training-4ps94 to complete
    curl  https://<argo-subit-hostname>/waitforstatus/test-training-4ps94 -X POST -d "@config.json"


## Setup

0. Create a secret to your private registry
    kubectl create secret docker-registry regcred3 --docker-server=https://<registry-url> --docker-username=<username> --docker-password=<password> -n argo-events

1. Import your kube-config into go-client/config
2. Write your settings in a config json. Example:


{
    "workflowImage": "<registry_url>/<repo>:<tag>",
    "inBucketName": "<local name of bucket>",
    "inBucketPath": "<local path in container>",
    "inputEndpoint": "<S3 url>",
    "inputBucketName": "<S3 name of bucket>",
    "inputKey": "<file or directory in bucket to fetch>",
    "inputAccessKey": "<name of accesskey in secret>",
    "inputCredentialSecretName": "<name of secret>",
    "inputSecretKey": "name of secretkey in secret",
    "outBucketName": "<container local name of bucket>",
    "outBucketPath": "<container local path of data to upload>",
    "outputEndpoint": "<endpoint of s3>",
    "outputBucketName": "<S3 bucket name>",
    "outputAccessKey": "name of access key in secret",
    "outputCredentialSecretName": "<name of secret>",
    "outputSecretKey": "name of security key in secret",
    "envs": [
        {
            "name": "test",
            "value": "test_value"
        },
        {
            "name": "test2",
            "value": "test_value2"
        }
    ]
}

3. Set the corrent registry in buildAndPush.sh
4. change values in .charts/templates
5. run buildAndPush.sh
6. run helm install argo-submit-server --name argo-submit


---------

TODO:

1. Store configs for later use
2. Enable/disable input S3 and output S3
3. Upload and store new workflow templates 
4. Stream logging from workflows/pods
5. Improve API to integrate with other services.
6. Implement support for other sources of data