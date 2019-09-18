from juju.controller import Controller
import WorkerScaler
import ScalerDbConnector
from juju import loop
from aiohttp import web
import os

controller_endpoint = os.environ['JUJU_CONTROLLER_ENDPOINT']
username = os.environ['JUJU_USERNAME']
password = os.environ['JUJU_PASSWORD']
# `juju show-controller` will display the certificate
cacert = os.environ['JUJU_CACERT']

controller = Controller()




def listStatus():
    return loop.run(WorkerScaler.listStatus(controller))

async def listStatus(request):
    #name = request.match_info.get('name', "Anonymous")
    await controller.connect(
        controller_endpoint,
        username,
        password,
        cacert,
    )
    ret = await WorkerScaler.listStatus(controller)
    #await controller.connection().close()

    ret = ret + "\n" + ScalerDbConnector.listClusterInfo()

    return web.Response(text=ret)

async def scaleUp(request):
    cores = int(request.match_info.get('cores', "Anonymous"))
    memory = int(request.match_info.get('memory', "Anonymous"))
    await controller.connect(
        controller_endpoint,
        username,
        password,
        cacert,
    )
    ret = await WorkerScaler.scaleUpWorker(cores,memory,controller)
    #await controller.connection().close()
    web.Response()
    return web.json_response(ret)

async def scaleUpSpecific(request):
    disk = int(request.match_info.get('disk', "Anonymous"))
    type = (request.match_info.get('type', "Anonymous"))
    await controller.connect(
        controller_endpoint,
        username,
        password,
        cacert,
    )
    ret = await WorkerScaler.scaleUpWorkerSpecific(disk, type, controller)
    #await controller.connection().close()
    web.Response()
    return web.json_response(ret)

async def scaleDown(request):
    id = request.match_info.get('machineid', "Anonymous")
    await controller.connect(
        controller_endpoint,
        username,
        password,
        cacert,
    )
    ret = await WorkerScaler.scaleDownWorkers(controller,id)
    #await controller.connection().close()
    return web.Response(text=ret)

async def removeMachine(request):
    id = request.match_info.get('machineid', "Anonymous")
    await controller.connect(
        controller_endpoint,
        username,
        password,
        cacert,
    )
    ret = await WorkerScaler.scaleDownWorkersOnMachineId(controller,id)
    #await controller.connection().close()
    return web.Response(text=ret)

async def updateMetaData(request):
    await controller.connect(
        controller_endpoint,
        username,
        password,
        cacert,
    )
    await WorkerScaler.updateMachineMetadata(controller)
    ret = "complete"
    #await controller.connection().close()
    return web.Response(text=ret)

app = web.Application()
app.add_routes([web.get('/status', listStatus),
                web.get('/scaleup/{cores}/{memory}', scaleUp),
                web.get('/scaleuptype/{disk}/{type}', scaleUpSpecific),
                web.get('/removemachine/{machineid}', removeMachine),
                web.get('/update', updateMetaData),
                web.get('/scaledown/{machineid}', scaleDown)])

print("READY")
print(controller_endpoint)


web.run_app(app)
