package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"io/ioutil"
	v12 "k8s.io/api/core/v1"
	"log"
	"net/http"
	"strings"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type ScaleUpResponse struct {
	UnitId    string `json:"unitId"`
	MachineId string `json:"machineId"`
}

type WorkflowInfo struct {
	WorkflowName string `json:"workflowName"`
	MachineId    string `json:"machineId"`
	UnitId       string `json:"unitId"`
	Scheduled    bool   `json:"scheduled"`
}

func autoscaler(){
	for {
		fmt.Println("Checking for unutilized nodes")
		go RemoveUnutilizedNodes()
		fmt.Println("Unutilized nodes checked. Checking for unscheduled work at: ", time.Now().Add(1 * time.Minute))
		time.Sleep(1 * time.Minute)
		fmt.Println("Checking for scheduled work")
		go checkForUnscheduledPods()
		fmt.Println("Pending work checked. Checking for unutilized nodes at: ", time.Now().Add(1 * time.Minute))
		time.Sleep(1 * time.Minute)
	}
}

func RemoveUnutilizedNodes() {
	clientset, err := kubernetes.NewForConfig(GlobalConfig)
	if err != nil {
		panic(err.Error())
	}

	nodes, err := clientset.CoreV1().Nodes().List(v1.ListOptions{})
	for e := range nodes.Items {
		podList, err := getNodePods(clientset, nodes.Items[e])
		nodeinfo, err := getNodeAllocatedResources(nodes.Items[e], podList)
		if(err != nil) {
			fmt.Println(err)
		}

		printNodeInfo(nodeinfo, nodes.Items[e].Name)

		//And workers > 1 MAKE THIS ASYNC!!!
		if(len(nodes.Items) > 1&& nodeinfo.CPURequestsFraction < 10.0 && nodeinfo.MemoryRequestsFraction < 10.0) {
			fmt.Println("REMOVING NODE: ", nodes.Items[e].Name)
			fmt.Println("Machine ID: ",  strings.Split(nodes.Items[e].Name, "-")[3])
			url := fmt.Sprintf(scalerUrl +"/removemachine/%s", strings.Split(nodes.Items[e].Name, "-")[3])

			fmt.Println(url)
			response, err := http.Get(url)
			if err != nil {
				fmt.Printf("The HTTP request failed with error %s\n", err)
			} else {
				data, _ := ioutil.ReadAll(response.Body)
				fmt.Println(string(data))
			}

		}
	}

	res, err := clientset.CoreV1().ResourceQuotas(Namespace).List((v1.ListOptions{}))

	for e := range res.Items {
		fmt.Println("RESOURCE: ", res.Items[e].Name)
	}
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


func checkForUnscheduledPods() {
	clientset, err := kubernetes.NewForConfig(GlobalConfig)
	if err != nil {
		panic(err.Error())
	}

	podClient := clientset.CoreV1().Pods(Namespace)
	podList, err := podClient.List(v1.ListOptions{})

	for _, currentPod := range podList.Items {
		for e := range currentPod.Status.Conditions {
			if (currentPod.Status.Conditions[e].Reason == "Unschedulable" &&
				currentPod.Status.Conditions[e].Status == "False" && !isScheduled(currentPod.Name)) {
				//Check if the pod is already scheduled

				printUnscheduledPods(currentPod, e)

				var memoryInGB = int(currentPod.Spec.Containers[1].Resources.Requests.Memory().Value()) / 1024 / 1024 / 1024
				var numOfCores = int(currentPod.Spec.Containers[1].Resources.Requests.Cpu().Value())
				go createNewMachine(numOfCores, memoryInGB, currentPod.Name)
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

func createNewMachine(numOfCores int, memoryInGB int, currentPod string) {
	url := fmt.Sprintf(scalerUrl +"/scaleup/%d/%d", numOfCores, memoryInGB)
	processingWorkflow(currentPod)
	fmt.Println(url)
	response, err := http.Get(url)
	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
	} else {
		body, _ := ioutil.ReadAll(response.Body)
		var data ScaleUpResponse
		json.Unmarshal(body, &data)

		fmt.Println("____________________________________________")
		fmt.Println("Scheduling Workflow %s on: ", currentPod)
		fmt.Println(string(data.MachineId))
		fmt.Println(string(data.UnitId))
		fmt.Println("____________________________________________")

		if (len(data.MachineId) != 0 && len(data.UnitId) != 0) {
			scheduleFlow(currentPod, data.MachineId, data.UnitId)
		}
	}
}

func processingWorkflow(workflowName string) bool {
	// Open up our database connection.
	var connectionString = fmt.Sprintf("%s:%s@tcp(%s)/%s", dbRootUser, dbRootPwd, dbHostName, database)
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		log.Print(err.Error())
	}
	defer db.Close()

	insert, err := db.Query("INSERT INTO WorkloadInfo (workflowName, machineId, unitId, scheduled) VALUES (?,?, ?, ?)", workflowName, "", "", true)
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
		return false
	}

	defer insert.Close()

	return true
}

func scheduleFlow(workflowName string, machineId string, unitId string) bool {
	// Open up our database connection.
	var connectionString = fmt.Sprintf("%s:%s@tcp(%s)/%s", dbRootUser, dbRootPwd, dbHostName, database)
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		log.Print(err.Error())
	}
	defer db.Close()

	insert, err := db.Query("UPDATE WorkloadInfo SET machineId=?, unitId=?, scheduled=? WHERE  workflowName = ?", machineId, unitId, true, workflowName)
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
		return false
	}

	defer insert.Close()

	return true
}

func isScheduled(workflowName string) bool {
	// Open up our database connection.
	var connectionString = fmt.Sprintf("%s:%s@tcp(%s)/%s", dbRootUser, dbRootPwd, dbHostName, database)
	db, err := sql.Open("mysql", connectionString)

	// if there is an error opening the connection, handle it
	if err != nil {
		log.Print(err.Error())
	}
	defer db.Close()

	var workflowInfo WorkflowInfo
	err = db.QueryRow("SELECT * FROM WorkloadInfo where workflowName = ?", workflowName).Scan(
		&workflowInfo.WorkflowName, &workflowInfo.MachineId, &workflowInfo.UnitId, &workflowInfo.Scheduled)
	if err != nil {
		log.Println(err.Error()) // proper error handling instead of panic in your app
		return false
	}

	fmt.Println("WF: ", workflowInfo.WorkflowName)
	fmt.Println("Sched: ", workflowInfo.Scheduled)

	return workflowInfo.Scheduled
}