#!/bin/bash -eu

query() {
  curl -s \
    -u ${OPENTSDB_READ_TOKEN} \
    -H 'Content-Type: application/json' \
    -XPOST ${OPENTSDB_URL}/api/query -d '{
      "start":0,
      "queries":[{
          "metric":"kafka.server:name=BytesOutPerSec,type=BrokerTopicMetrics.MeanRate",
          "aggregator":"avg",
          "downsample":"20s-avg",
          "tags":{}
      }]
  }'
}

suggest() {
  curl -s \
    -u ${OPENTSDB_READ_TOKEN} \
    -H 'Content-Type: application/json' \
    -XGET "${OPENTSDB_URL}/api/suggest?type=metrics&q=kafka&max=1000"
}

suggest
#query