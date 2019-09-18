import mysql.connector
import os
#scaler-db-mysql.submit-scaler.svc.cluster.local

def getMysqlConnection():
  mydb = mysql.connector.connect(
    host="scaler-db-mysql.submit-scaler.svc.cluster.local",
    #port='3306',
    user=os.environ['SCALER_MYSQL_USER'],
    passwd=os.environ['SCALER_MYSQL_PASSWORD'],
    database="scalerdb"
    )

  return mydb

def testConnection():
  try:
    mydb = mysql.connector.connect(
      host="scaler-db-mysql.submit-scaler.svc.cluster.local",
      #port="3306",
      user=os.environ['SCALER_MYSQL_USER'],
      passwd=os.environ['SCALER_MYSQL_PASSWORD'] ,
      database="scalerdb"
    )

    print(mydb)
    mycursor = mydb.cursor()
    mycursor.execute("SHOW DATABASES")

    count = 0;
    for x in mycursor:
      print(x)
      count = count + 1

    mycursor.close()
    mydb.close()

    if count == 0:
      return False
    else:
      return True
  except:
    return False


def insertCreatedWorker(machineId, unitId, nodeName, removable, active):
  mydb = getMysqlConnection()
  print(mydb)
  mycursor = mydb.cursor()

  sql = 'INSERT INTO ClusterInfo (machineId, unitId, nodeName, removable, active) VALUES (%s, %s, %s, %s, %s)'
  val = (machineId, unitId, nodeName, removable, active)
  mycursor.execute(sql, val)

  mydb.commit()
  mycursor.close()
  mydb.close()

def isRemovable(machineId):
  mydb = getMysqlConnection()
  mycursor = mydb.cursor()
  sql = ("SELECT removable FROM ClusterInfo WHERE machineId = %s")
  val = (machineId,)
  mycursor.execute(sql, val)

  for (removable) in mycursor:
    mycursor.close()
    mydb.close()
    return int(removable[0]) == 1

  mycursor.close()
  mydb.close()
  return False

def deleteMachine(machineId):
  mydb = getMysqlConnection()
  mycursor = mydb.cursor()
  sql = ("DELETE FROM ClusterInfo WHERE machineId = %s")
  val = (machineId,)
  mycursor.execute(sql, val)
  mydb.commit()
  mycursor.close()
  mydb.close()

def listClusterInfo():
  mydb = getMysqlConnection()
  print(mydb)
  mycursor = mydb.cursor()
  mycursor.execute("SELECT * FROM ClusterInfo")
  myresult = mycursor.fetchall()

  ret = ""
  for x in myresult:
    for value in x:
      ret = ret + "   " + str(value)
    ret = ret + "\n"


  mycursor.close()
  mydb.close()
  return ret

def updateMachineMetadata(machineId, addresses, instanceId):
  if exists(machineId):
    addMetadata(machineId, instanceId, addresses)
  else:
    insertCreatedWorker(machineId, "", "", 0, 1)

def exists(machineId):
  mydb = getMysqlConnection()
  mycursor = mydb.cursor()
  sql = ("SELECT machineId FROM ClusterInfo WHERE machineId = %s")
  val = (machineId,)
  mycursor.execute(sql, val)

  for (removable) in mycursor:
    mycursor.close()
    mydb.close()
    return True

  mycursor.close()
  mydb.close()
  return False

def addMetadata(machineId, instanceId, addresses):
  mydb = getMysqlConnection()
  mycursor = mydb.cursor()
  sql = ("UPDATE ClusterInfo SET instanceId=%s, addresses=%s WHERE machineId = %s")

  val = (instanceId, addresses, machineId)
  mycursor.execute(sql, val)

  mydb.commit()
  mycursor.close()
  mydb.close()
