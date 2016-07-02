j () {
  curl -s -u ${JMX_AUTH} \
    "http://${BROKER_URL}/read/$1" | jq .
}

k() {
  curl -s "localhost:4242/metrics/$1" -w '\n{"status":%{response_code}}' | jq .
}

list() {
  # https://signalfx.com/kafka-monitoring/

  # Bytes In
  kafka.server:name=BytesInPerSec,type=BrokerTopicMetrics
  kafka.server:name=BytesInPerSec,topic=*,type=BrokerTopicMetrics
  # Bytes Out
  kafka.server:name=BytesOutPerSec,type=BrokerTopicMetrics
  kafka.server:name=BytesOutPerSec,topic=*,type=BrokerTopicMetrics
  # Messages In
  kafka.server:name=MessagesInPerSec,type=BrokerTopicMetrics
  kafka.server:name=MessagesInPerSec,topic=*,type=BrokerTopicMetrics

  # * All topic metrics
  #kafka.server:name=*,type=BrokerTopicMetrics
  #kafka.server:name=*,topic=*,type=BrokerTopicMetrics

  # Active Controllers
  kafka.controller:name=ActiveControllerCount,type=KafkaController

  # Request Queue
  kafka.network:name=RequestQueueSize,type=RequestChannel
  kafka.network:name=RequestQueueTimeMs,request=*,type=RequestMetrics

  # Under Replicated Partitions
  kafka.server:name=UnderReplicatedPartitions,type=ReplicaManager
  kafka.server:name=UnderReplicated,partition=*,topic=*,type=Partition

  # Log Flushes
  # Log Flush Time in Ms
  # Log Flush Time in Ms - 95th Percentile
  kafka.log:name=LogFlushRateAndTimeMs,type=LogFlushStats

  # Produce Total Time
  # Produce Total Time - 99th Percentile
  # Produce Total Time - Median
  kafka.network:name=TotalTimeMs,request=Produce,type=RequestMetrics

  # Fetch Consumer Total Time
  # Fetch Consumer Total Time - 99th Percentile
  # Fetch Consumer Total Time - Median
  kafka.network:name=TotalTimeMs,request=FetchConsumer,type=RequestMetrics

  # Fetch Follower Total Time
  # Fetch Follower Total Time  - 99th Percentile
  # Fetch Follower Total Time - Median
  kafka.network:name=TotalTimeMs,request=FetchFollower,type=RequestMetrics
}