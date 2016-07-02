build: build-go
	docker build -t krkr/kafka-metrics-agent .

build-go:
	docker run --rm \
		-e GOBIN=/go/bin/ -e CGO_ENABLED=0 -e GOPATH=/go \
		-v $$(pwd):/go/src/github.com/thbkrkr/kafka-metrics-agent\
		-w /go/src/github.com/thbkrkr/kafka-metrics-agent \
		golang:1.6.2-alpine go build

dev:
	@export $(cat .env | xargs) > /dev/null && \
	go build && ./kafka-metrics-agent \
		-kafka-brokers=${KAFKA_BROKERS} \
		-kafka-jmx-password=${KAFKA_JMX_PASSWORD} \
		-opentsdb-url=${OPENTSDB_URL} \
		-opentsdb-token=${OPENTSDB_TOKEN}
