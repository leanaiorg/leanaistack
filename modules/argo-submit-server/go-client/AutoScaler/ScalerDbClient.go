package AutoScaler

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
)

type WorkflowInfo struct {
	WorkflowName string `json:"workflowName"`
	MachineId    string `json:"machineId"`
	UnitId       string `json:"unitId"`
	Scheduled    bool   `json:"scheduled"`
}

type ClusterInfo struct {
	NodeName   string `json:"nodeName"`
	MachineId  string `json:"machineId"`
	UnitId     string `json:"unitId"`
	InstanceId string `json:"instanceId"`
	Addresses  string `json:"addresses"`
	Removable  bool   `json:"removable"`
	Active     bool   `json:"active"`
}

func storeWorkflow(workflowName string) bool {
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

func findMachineId(providerId string, internalIP string) (string, bool) {
	machines := getMachines()
	for machines.Next() {
		var clusterInfo ClusterInfo
		machines.Scan(
			&clusterInfo.MachineId,
			&clusterInfo.UnitId,
			&clusterInfo.NodeName,
			&clusterInfo.InstanceId,
			&clusterInfo.Addresses,
			&clusterInfo.Removable,
			&clusterInfo.Active)

		defer machines.Close()

		if (clusterInfo.InstanceId == "") {
			continue
		}

		if strings.Contains(clusterInfo.Addresses, internalIP) || strings.Contains(providerId, clusterInfo.InstanceId) {
			return clusterInfo.MachineId, true
		}
	}

	return "", false
}

func isRemovable(providerId string, internalIP string) (bool) {
	machines := getMachines()
	for machines.Next() {
		var clusterInfo ClusterInfo
		machines.Scan(
			&clusterInfo.MachineId,
			&clusterInfo.UnitId,
			&clusterInfo.NodeName,
			&clusterInfo.InstanceId,
			&clusterInfo.Addresses,
			&clusterInfo.Removable,
			&clusterInfo.Active)

		defer machines.Close()

		if (clusterInfo.InstanceId == "") {
			fmt.Println(clusterInfo.InstanceId)
			continue
		}

		if strings.Contains(clusterInfo.Addresses, internalIP) || strings.Contains(providerId, clusterInfo.InstanceId) {
			return clusterInfo.Removable
		}
	}

	return false
}

func getMachines() *sql.Rows {
	// Open up our database connection.
	var connectionString = fmt.Sprintf("%s:%s@tcp(%s)/%s", dbRootUser, dbRootPwd, dbHostName, database)
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		log.Print(err.Error())
	}

	defer db.Close()

	machines, err := db.Query("SELECT * FROM ClusterInfo")
	if err != nil {
		log.Println(err.Error()) // proper error handling instead of panic in your app
		return machines
	}

	return machines
}
