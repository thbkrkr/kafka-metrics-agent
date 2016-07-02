FROM alpine:3.4

RUN apk --no-cache add ca-certificates

COPY kafka-metrics-collector /kafka-metrics-collector

ENTRYPOINT ["/kafka-metrics-collector"]