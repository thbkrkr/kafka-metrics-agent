# Kafka metrics agent

Agent to collect Kafka metrics and send them to OpenTSDB.

```
docker run --rm \
  krkr/kafka-metrics-agent \
    -kafka-brokers=$KAFKA_BROKERS
    -kafka-jmx-password=$KAFKA_JMX_PASSWORD
    -opentsdb-url=$OPENTSDB_URL
    -opentsdb-token=$OPENTSDB_TOKEN
```