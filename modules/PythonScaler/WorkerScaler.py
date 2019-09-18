from juju.model import Model
import ScalerDbConnector
import json

GB = 1024

async def scaleUpWorker(cores, memoryGB, controller):
    if not ScalerDbConnector.testConnection():
        return "could not connect to databse. Will not create a new worker"

    kubernetesWorkerUnitName = 'kubernetes-worker'
    defaultModel = await controller.get_model("default")
    await defaultModel.get_status()
    model = Model()
    await model.connect()

    machine = await model.add_machine(
        constraints={
            'mem': memoryGB * GB,
            'cores': cores #,
            #'root-disk': 30 * GB
        },
        series='bionic',
    )

    print("DEPLOYING MACHINE WITH ID: " + machine.id)

    id = ""
    for application in defaultModel.applications.items():
        if application[1].entity_id is not None and application[1].entity_id == kubernetesWorkerUnitName:
            newUnit = await application[1].add_unit(
                to=machine.id
            )

            id = newUnit[0].entity_id

    ScalerDbConnector.insertCreatedWorker(machine.id, id, "EMPTY", True, True)

    retStr = {
        "unitId": id,
        "machineId": machine.id
    }

    await model.disconnect()
    await controller.disconnect()

    return retStr

async def scaleUpWorkerSpecific(disk, type, controller):
    if not ScalerDbConnector.testConnection():
        return "could not connect to databse. Will not create a new worker"

    kubernetesWorkerUnitName = 'kubernetes-worker'
    defaultModel = await controller.get_model("default")
    await defaultModel.get_status()
    model = Model()
    await model.connect()

    machine = await model.add_machine(
        constraints={
            'root-disk': disk * GB,
            'instance-type': type
        },
        series='bionic',
    )

    print("DEPLOYING MACHINE WITH ID: " + machine.id)

    id = ""
    for application in defaultModel.applications.items():
        if application[1].entity_id is not None and application[1].entity_id == kubernetesWorkerUnitName:
            newUnit = await application[1].add_unit(
                to=machine.id
            )

            id = newUnit[0].entity_id

    ScalerDbConnector.insertCreatedWorker(machine.id, id, "EMPTY", True, True)

    retStr = {
        "unitId": id,
        "machineId": machine.id
    }

    await model.disconnect()
    await controller.disconnect()

    return retStr



def extractAddresses(addresses):
    if (addresses is None):
        return "No Addresses found"

    ips = ""
    for address in addresses:
        ips = ips + ", " + address['value']

    return ips


async def updateMachineMetadata(controller):
    defaultModel = await controller.get_model("default")
    await defaultModel.get_status()
    model = Model()
    await model.connect()

    for machine in defaultModel.machines.items():
        instanceId = machine[1].data['instance-id']
        addresses = machine[1].data['addresses']
        id = machine[0]
        ips = extractAddresses(addresses)
        ScalerDbConnector.updateMachineMetadata(id, ips, instanceId)

    await model.disconnect()
    await controller.disconnect()


async def scaleDownWorkers(controller, num):
    kubernetesWorkerUnitName = 'kubernetes-worker/' + num
    defaultModel = await controller.get_model("default")
    await defaultModel.get_status()
    model = Model()
    await model.connect()

    units = defaultModel.units
    toRemove = units.get(kubernetesWorkerUnitName)
    machineToRemove = toRemove.machine
    machineId = machineToRemove.entity_id

    if ScalerDbConnector.isRemovable(machineId):
        await toRemove.remove()
        await machineToRemove.remove()
        ScalerDbConnector.deleteMachine(machineId)
    else:
        print("Cannot remove machine: " + machineToRemove.entity_id + " because it is marked as non removable")

    await model.disconnect()
    await controller.disconnect()

async def listStatus(controller):
    defaultModel = await controller.get_model("default")
    await defaultModel.get_status()
    model = Model()
    await model.connect()

    retStr = "------- Machines -------"

    for machine in defaultModel.machines:
        retStr = retStr + "\n" + "Machine ID: " + machine

    retStr = retStr + "\n" + "------- Units -------"

    for units in defaultModel.units.items():
        if units[1].machine is not None:
            retStr = retStr + "\n" + "Unit: " + units[1].entity_id + " Is running on machine: " +  units[1].machine.entity_id

    await model.disconnect()
    await controller.disconnect()

    return retStr

async def scaleDownWorkersOnMachineId(controller, machineId):
    if not ScalerDbConnector.isRemovable(machineId):
        return ("Cannot remove machine: " + machineId + " because it is marked as non removable")
    else:
        defaultModel = await controller.get_model("default")
        await defaultModel.get_status()
        model = Model()
        await model.connect()

        for units in defaultModel.units.items():
            if units[1].machine is not None:
                if units[1].machine.entity_id == machineId:
                    unitName =  units[1].entity_id
                    machineName = units[1].machine.entity_id
                    machineToRemove = units[1].machine
                    await units[1].remove()
                    await machineToRemove.remove()
                    ScalerDbConnector.deleteMachine(machineId)
                    await model.disconnect()
                    await controller.disconnect()
                    return "Removed unit: " + unitName + " and machine: " + machineName

    await model.disconnect()
    await controller.disconnect()

    return "fail"
