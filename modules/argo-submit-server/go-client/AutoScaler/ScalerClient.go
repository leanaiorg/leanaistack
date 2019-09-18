package AutoScaler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)


func callRemoveMachine(machineId string) (bool, string){
	url := fmt.Sprintf(scalerUrl+"/removemachine/%s", machineId)
	fmt.Println(url)
	response, err := http.Get(url)
	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
		return false, fmt.Sprintf("The HTTP request failed with error %s\n", err)
	} else {
		data, _ := ioutil.ReadAll(response.Body)
		fmt.Println(string(data))
		return true, string(data)
	}
}

func createNewMachine(numOfCores int, memoryInGB int, currentPod string) {
	url := fmt.Sprintf(scalerUrl+"/scaleup/%d/%d", numOfCores, memoryInGB)
	fmt.Println(url)
	response, err := http.Get(url)
	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
	} else {
		body, _ := ioutil.ReadAll(response.Body)
		var data ScaleUpResponse
		json.Unmarshal(body, &data)

		fmt.Println("Scheduled: ", currentPod)

		if len(data.MachineId) != 0 && len(data.UnitId) != 0 {
			scheduleFlow(currentPod, data.MachineId, data.UnitId)
		}
	}
}

func createNewMachineOfType(instanceType string, disk int,currentPod string) {

	///scaleuptype/40/m5.4xlarge
	url := fmt.Sprintf(scalerUrl+"/scaleuptype/%d/%s", disk, instanceType)
	fmt.Println(url)
	response, err := http.Get(url)
	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
	} else {
		body, _ := ioutil.ReadAll(response.Body)
		var data ScaleUpResponse
		json.Unmarshal(body, &data)

		fmt.Println("Scheduled: ", currentPod)

		if len(data.MachineId) != 0 && len(data.UnitId) != 0 {
			scheduleFlow(currentPod, data.MachineId, data.UnitId)
		}
	}
}

func updateMachineData() {
	url := fmt.Sprintf(scalerUrl + "/update")

	fmt.Println(url)
	response, err := http.Get(url)
	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
	} else {
		data, _ := ioutil.ReadAll(response.Body)
		fmt.Println(string(data))
	}
}