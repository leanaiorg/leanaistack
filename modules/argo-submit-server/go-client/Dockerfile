FROM golang:latest

WORKDIR $GOPATH/src/main
RUN mkdir ~/.kube

ENV GOBIN /go/bin

COPY go.mod .
COPY go.sum .
COPY argo-cli.go .
COPY ProcessWorkflow.go .
COPY WorkQueueProducer.go .
COPY workflowTemplate.json .
COPY AutoScaler AutoScaler/

RUN ls -l

COPY config /root/.kube/config
RUN go build

CMD go run main
