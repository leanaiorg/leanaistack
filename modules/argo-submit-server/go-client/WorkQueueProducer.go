
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/minio/minio-go"
	"github.com/streadway/amqp"
	"log"
	"math/rand"
	"time"
)

type WorkItem struct {
	FileName string `json:"fileName"`
	S3Uri  string `json:"s3Uri"`
	Bucket string  `json:"bucket"`
}

var ssl = true
var conn *amqp.Connection
var ch *amqp.Channel
var (
	amqpURI = flag.String("amqp", rabbitConnection, "AMQP URI")
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

func init() {
	/*
	flag.Parse()
	initAmqp("default-workqueue")

	 */
}


func initAmqp(queueName string) {
	var err error

	conn, err = amqp.Dial(*amqpURI)
	failOnError(err, "Failed to connect to RabbitMQ")

	ch, err = conn.Channel()
	failOnError(err, "Failed to open a channel")

	q, err := ch.QueueDeclare(
		queueName, // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	failOnError(err, "Failed to declare a queue")
	fmt.Println("Created Queue: ", q.Name)
}

func publishMessage(message string, s3uri string, bucket string, queueName string) {
	workItem := WorkItem{}
	workItem.FileName = message
	workItem.S3Uri = s3uri
	workItem.Bucket = bucket

	payload, err := json.Marshal(workItem)
	failOnError(err, "Failed to marshal JSON")

	err = ch.Publish(
		"",     // exchange
		queueName, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing {
			ContentType: "text/plain",
			Body:        []byte(payload),
		})
	failOnError(err, "Failed to publish a message")
}

func produceJobs(s3Uri string, bucketName string, prefix string, queueName string) {
	minioClient, err := minio.New(s3Uri, accessKey, secretKey, ssl)
	if err != nil {
		fmt.Println(err)
		return
	}

	initAmqp(queueName)
	// Create a done channel to control 'ListObjectsV2' go routine.
	doneCh := make(chan struct{})

	// Indicate to our routine to exit cleanly upon return.
	defer close(doneCh)

	isRecursive := true
	objectCh := minioClient.ListObjectsV2(bucketName, prefix, isRecursive, doneCh)
	for object := range objectCh {
		if object.Err != nil {
			fmt.Println(object.Err)
			return
		}

		fmt.Println("Key: ", object.Key)
		publishMessage(object.Key, s3Uri,bucketName, queueName)
	}
}

func createQueueId(length int) string {
	var seededRand *rand.Rand = rand.New(
		rand.NewSource(time.Now().UnixNano()))

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func generateQueueNameIfNotProvided(workflowQueueInfo WorkQueueInfo) string {
	fmt.Println("Producing S3 work")
	if len(workflowQueueInfo.QueueName) == 0 {
		return createQueueId(10)
	}

	return workflowQueueInfo.QueueName
}

func defineWorkQueue(workflowConfig WorkflowConfig) string {
	var createdWFResponse = ""
	for index := range workflowConfig.WorkflowSteps {
		var workflowQueueInfo = workflowConfig.WorkflowSteps[index].WorkQueueInfo

		if workflowQueueInfo.Enabled {
			//Set queue name for later env variable in the pods
			workflowQueueInfo.QueueName = generateQueueNameIfNotProvided(workflowQueueInfo)

			createdWFResponse = createdWFResponse + "Running on queue: " +
				workflowQueueInfo.QueueName + " "

			workflowConfig.WorkflowSteps[index].WorkQueueInfo.QueueName = workflowQueueInfo.QueueName

			produceJobs(workflowQueueInfo.S3Uri,
				workflowQueueInfo.S3Bucket,
				workflowQueueInfo.Prefix,
				workflowQueueInfo.QueueName)
		}
	}

	return createdWFResponse
}