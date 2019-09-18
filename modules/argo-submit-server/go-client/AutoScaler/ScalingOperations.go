package AutoScaler

import (
	"fmt"
	"k8s.io/client-go/rest"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	v12 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type ScaleUpResponse struct {
	UnitId    string `json:"unitId"`
	MachineId string `json:"machineId"`
}

var scalerUrl = os.Getenv("SCALER_URL")

var dbRootUser = os.Getenv("SCALER_MYSQL_USER")
var dbRootPwd = os.Getenv("SCALER_MYSQL_PASSWORD")
var dbHostName = os.Getenv("DB_HOST_NAME")
var database = os.Getenv("SCALER_DATABASE")

var GlobalConfig *rest.Config
var Namespace string
var NodesInRemoval = make(map[string]bool)

func Autoscaler(config *rest.Config, namespace string) {

	Namespace = namespace
	GlobalConfig = config

	for {
		fmt.Println("___________________REFRESHING MACHINE METADATA___________________")
		updateMachineData()
		fmt.Println("__________________________________________________________________")

		fmt.Println("_____________________CHECKING FOR UNUSED NODES_____________________")
		go RemoveUnutilizedNodes()

		fmt.Println("Checking for unscheduled work at: ", time.Now().Add(1*time.Minute))
		time.Sleep(1 * time.Minute)

		fmt.Println("___________________REFRESHING MACHINE METADATA___________________")
		updateMachineData()
		fmt.Println("__________________________________________________________________")

		fmt.Println("_____________________CHECKING FOR UNSCHEDULED PODS_____________________")
		go checkForUnscheduledPods()

		fmt.Println("Checking for unutilized nodes at: ", time.Now().Add(1*time.Minute))
		time.Sleep(1 * time.Minute)
	}
}

func handleError(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func lockNode(name string) {
	NodesInRemoval[name] = true
}

func unlockNode(name string) {
	NodesInRemoval[name] = false
}

//Should check for errors if it doesnt exist
func nodeUnderRemoval(name string) bool {
	return NodesInRemoval[name]
}

func clearNodesNotInRemoval() {
	for nodeName, inRemoval := range NodesInRemoval {
		if !inRemoval {
			delete(NodesInRemoval, nodeName)
			fmt.Printf("Node %s is not under remove-check")
		}
	}
}

func RemoveUnutilizedNodes() {
	clientset, err := kubernetes.NewForConfig(GlobalConfig)
	handleError(err)

	fmt.Println("Known nodes and their states", NodesInRemoval)
	clearNodesNotInRemoval()
	fmt.Println("Known nodes under removal process", NodesInRemoval)

	nodes, err := clientset.CoreV1().Nodes().List(v1.ListOptions{})

	for _, node := range nodes.Items {
		//Add new nodes to the map
		if _, ok := NodesInRemoval[node.Name]; !ok {
			NodesInRemoval[node.Name] = false
		}

		podList, err := getNodePods(clientset, node)
		nodeinfo, err := getNodeAllocatedResources(node, podList)
		handleError(err)

		printNodeInfo(nodeinfo, node.Name)

		go removeUnderUtilizedNode(nodes, nodeinfo, node)
	}
}

//Create a timer here:
func removeUnderUtilizedNode(nodes *v12.NodeList, nodeinfo NodeAllocatedResources, node v12.Node) bool {
	//Make this part async
	if(nodeUnderRemoval(node.Name)) {
		return false
	}

	lockNode(node.Name)
	defer unlockNode(node.Name)

	providerId := node.Spec.ProviderID
	internalIp := extractNodeInternalIp(node)
	machineId, found := findMachineId(providerId, internalIp)

	if !found /*|| providerId=="")*/  || !isRemovable(providerId, internalIp) {
		printRemoveNodeSkipMessage(node, providerId, internalIp, found)
		return false
	}

	//Check if the node is underutilized (for ten minutes in a row)
	for i := 1; i <= 10; i++ {
		if !isNodeUnderutilized(nodes, nodeinfo) {
			return false
		}

		time.Sleep(1*time.Minute)
		fmt.Printf("Check %d of 10 completed to remove node %s \n", i, node.Name)
	}

	fmt.Println("Removing unit with Machine ID: ", machineId)
	callRemoveMachine(machineId)

	return true
}


func isNodeUnderutilized(nodes *v12.NodeList, nodeinfo NodeAllocatedResources) bool {
	return len(nodes.Items) > 1 && nodeinfo.CPURequestsFraction < 15.0 && nodeinfo.MemoryRequestsFraction < 10.0
}

func extractNodeInternalIp(node v12.Node) string {
	for _, address := range node.Status.Addresses {
		if address.Type == "InternalIP" {
			return address.Address
		}
	}
	return ""
}

func checkForUnscheduledPods() {
	clientset, err := kubernetes.NewForConfig(GlobalConfig)
	handleError(err)

	podClient := clientset.CoreV1().Pods(Namespace)
	podList, err := podClient.List(v1.ListOptions{})

	for _, currentPod := range podList.Items {
		for e, condition := range currentPod.Status.Conditions {
			if condition.Reason == "Unschedulable" &&
				condition.Status == "False" && !isScheduled(currentPod.Name) {
				//Check if the pod is already scheduled

				printUnscheduledPods(currentPod, e)

				var memoryInGB = int(currentPod.Spec.Containers[1].Resources.Requests.Memory().Value()) / 1024 / 1024 / 1024
				var numOfCores = int(currentPod.Spec.Containers[1].Resources.Requests.Cpu().Value())

				if(currentPod.Labels["InstanceType"] != "") {
					fmt.Println("INSTANCE TYPE: ", currentPod.Labels["InstanceType"])
					createNewMachineOfType(currentPod.Labels["InstanceType"], 80, currentPod.Name)
				} else {
					go createNewMachine(numOfCores, memoryInGB, currentPod.Name)
				}

				storeWorkflow(currentPod.Name)
			}
		}
	}
}

func printUnscheduledPods(currentPod v12.Pod, e int) {
	fmt.Println("-----------------------------------------------------")
	fmt.Println("REASON:  ", currentPod.Status.Conditions[e].Reason)
	fmt.Println("STATUS:  ", currentPod.Status.Conditions[e].Status)
	fmt.Println("MESSAGE:  ", currentPod.Status.Conditions[e].Message)

	fmt.Println(currentPod.Spec.Containers[0].Resources.Requests.Cpu())
	fmt.Println("CPU: ", currentPod.Spec.Containers[1].Resources.Requests.Cpu().Value())
	fmt.Println("Memory: ", currentPod.Spec.Containers[1].Resources.Requests.Memory().Value())
	fmt.Println("-----------------------------------------------------")
}

func printNodeInfo(nodeinfo NodeAllocatedResources, nodeName string) {
	fmt.Println("-----------------------------------------------------")
	fmt.Println("NODE: ", nodeName)
	fmt.Printf("CPU %f\n", nodeinfo.CPURequestsFraction)
	fmt.Printf("MEMORY %f\n", nodeinfo.MemoryRequestsFraction)
	fmt.Printf("CPU_AVAIL %f\n", nodeinfo.CPURequests)
	fmt.Printf("MEMORY_AVAIL %f\n", nodeinfo.MemoryRequests)
	fmt.Println("-----------------------------------------------------")
}

func printRemoveNodeSkipMessage(node v12.Node, providerId string, internalIp string, found bool) {
	fmt.Printf("Could not remove %s due to one of the following reasons: " +
		"\n   1. The node was not found: %t" +
		"\n   2, The providerId was empty %t" +
		"\n   3. The node was not marked as removable: %t \n",
		node.Name, !found,providerId=="", !isRemovable(providerId, internalIp) )
}
