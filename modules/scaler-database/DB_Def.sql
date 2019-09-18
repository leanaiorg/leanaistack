CREATE DATABASE scalerdb; 

USE scalerdb;

CREATE TABLE ClusterInfo (
    machineId varchar(30),
    unitId varchar(30),
    nodeName varchar(30),
    instanceId varchar(256),
    addresses varchar(256),
    removable boolean,
    active boolean
); 


CREATE TABLE WorkloadInfo (
    workflowName varchar(30),
    machineId varchar(30),
    unitId varchar(30),
    scheduled boolean
); 


insert into ClusterInfo (machineId, unitId, nodeName, removable, active) VALUES  ("2","kubernetes-worker/0", "juju-063677-default-2", False, True) 

INSERT INTO WorkloadInfo (
    workflowName, machineId, unitId, scheduled
) VALUES 
(
    "testWL",
    "10",
    "10",
    True
)

