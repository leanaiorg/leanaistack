package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"

	//Workflow variables
	wfv1 "github.com/argoproj/argo/pkg/apis/workflow/v1alpha1"
	wfclientset "github.com/argoproj/argo/pkg/client/clientset/versioned"
	"github.com/argoproj/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"main/AutoScaler"
)

type EnvPair struct {
	Name  					string 				`json:"name"`
	Value 					string 				`json:"value"`
}

type ResourcesRequest struct {
	Request 				bool				`json:"request"`
	Cpu     				int			    	`json:"cpu"`
	Memory  				string 				`json:"memory"`
	InstanceType			string				`json:"instanceType"`
}

type WorkflowConfig struct {
	WorkflowSteps			[]WorkflowStep		`json:"steps"`
	ResourcesRequest        ResourcesRequest	`json:"resources"`
}

type WorkflowStep struct {
	Step					int					`json:"step"`
	WorkflowImage           string           	`json:"workflowImage"`
	InputBucket				Bucket				`json:"inputBucket"`
	OutputBucket			Bucket				`json:"outputBucket"`
	Environments            []EnvPair        	`json:"envs"`
	WorkQueueInfo           WorkQueueInfo    	`json:"workQueue"`
	NumOfWorkers            int              	`json:"numOfWorkers"`
}

type WorkQueueInfo struct {
	Enabled   				bool   				`json:"enabled"`
	S3Uri     				string 				`json:"s3Uri"`
	S3Bucket  				string 				`json:"s3bucket"`
	Prefix    				string 				`json:"prefix"`
	QueueName 				string 				`json:"queueName"`
}

type Bucket struct {
	Enabled					bool			 `json:"enabled"`
	BucketPath				string           `json:"bucketPath"`
	Endpoint				string           `json:"endpoint"`
	BucketName				string           `json:"bucketName"`
	Key						string           `json:"key"`
	AccessKey				string           `json:"accessKey"`
	CredentialSecretName	string           `json:"credentialSecretName"`
	SecretKey				string           `json:"secretKey"`
}

//Used to generate a queuename
const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var GlobalConfig *rest.Config
var Namespace string = "default"//os.Getenv("WORKFLOW_NAMESPACE")

var rabbitUri = os.Getenv("RABBIT_URI")
var rabbitPwd = os.Getenv("WORKQUEUE_PASSWORD")
var rabbitUsr = os.Getenv("WORKQUEUE_USER")
var rabbitConnection = fmt.Sprintf("amqp://%s:%s@%s", rabbitUsr, rabbitPwd, rabbitUri)

var accessKey string = os.Getenv("MINIO_USER")
var secretKey string = os.Getenv("MINIO_PASSWORD")

// Fetched the kube config as a global var
func init() {
	usr, err := user.Current()
	checkErr(err)
	kubeWatcher := flag.String("kubeconf", filepath.Join(usr.HomeDir, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	flag.Parse()
	config, err := clientcmd.BuildConfigFromFlags("", *kubeWatcher)
	checkErr(err)

	GlobalConfig = config
}

func main() {
	var router = mux.NewRouter()
	router.HandleFunc("/healthcheck", healthCheck).Methods("GET")
	router.HandleFunc("/submit", subitWorkflow).Methods("POST")
	router.HandleFunc("/waitforstatus/{podname}", watchWorkflow).Methods("POST")

	fmt.Println("Starting autoscaler")
	go AutoScaler.Autoscaler(GlobalConfig, Namespace)
	fmt.Println("Running Server")
	log.Fatal(http.ListenAndServe(":3000", router))

}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode("Still alive!")
}

func subitWorkflow(w http.ResponseWriter, r *http.Request) {
	byteValue, _ := ioutil.ReadAll(r.Body)
	var workflowConfig WorkflowConfig
	json.Unmarshal(byteValue, &workflowConfig)
	ret := runArgoWorkflow(workflowConfig)
	json.NewEncoder(w).Encode(ret)
}

func watchWorkflow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	byteValue, _ := ioutil.ReadAll(r.Body)
	var workflowConfig WorkflowConfig
	json.Unmarshal(byteValue, &workflowConfig)
	podName := vars["podname"]

	ret := watchJob(podName, workflowConfig)
	json.NewEncoder(w).Encode(ret)
}

func loadWorkflowTemplate() wfv1.Workflow {
	file, _ := ioutil.ReadFile("workflowTemplate.json")
	trainingWorkflow := wfv1.Workflow{}
	_ = json.Unmarshal([]byte(file), &trainingWorkflow)

	return trainingWorkflow
}

func runArgoWorkflow(workflowConfig WorkflowConfig) string {
	//Read template
	trainingWorkflow := loadWorkflowTemplate()

	//Submit workqueue:
	var createdWFResponse = defineWorkQueue(workflowConfig)
	createdWFResponse = createdWFResponse + "Created the workflows: "

	trainingWorkflow = processWorkflow(trainingWorkflow, workflowConfig)

	// create the workflow client
	wfClient := wfclientset.NewForConfigOrDie(GlobalConfig).ArgoprojV1alpha1().Workflows(Namespace)
	createdWf, err := wfClient.Create(&trainingWorkflow)
	checkErr(err)

	createdWFResponse = fmt.Sprintf(createdWFResponse+"%s, ", createdWf.Name)

	time.Sleep(2 * time.Second)
	return createdWFResponse
}

func watchJob(name string, workflowConfig WorkflowConfig) string {
	wfClient := wfclientset.NewForConfigOrDie(GlobalConfig).ArgoprojV1alpha1().Workflows(Namespace)

	getIf, err := wfClient.Get(name, metav1.GetOptions{})
	errors.CheckError(err)
	fmt.Sprintf("Workflow Name %s, %s\n", getIf.Name, getIf.Status)
	return fmt.Sprintf("Status: %s \n", getIf.Status)
}

func checkErr(err error) {
	if err != nil {
		panic(err.Error())
	}
}
