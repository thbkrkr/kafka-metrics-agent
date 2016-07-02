#!/bin/bash -u

wait_server() {
  sleep 1
}

create_datasource() {
  sleep 3
  echo "create_datasource kafka-metrics"
  curl -u admin:admin \
    -H "Content-Type:application/json" \
    -XPOST localhost:3000/api/datasources -d '
  {
    "name":"kafka-metrics",
    "type":"opentsdb",
    "access":"proxy",
    "jsonData": {
      "tsdbVersion": 2,
      "tsdbResolution": 1
    },
    "url":"'$OPENTSDB_URL'",
    "basicAuth":true,
    "basicAuthUser":"'$OPENTSDB_USER'",
    "basicAuthPassword":"'$OPENTSDB_PASSWORD'"
  }' -v
}

create_dashboard() {
  sleep 4

  for d in $(ls -1 /dashbs); do
    echo "create_dashboard $d"
    curl -u admin:admin \
      -H "Content-Type:application/json" \
      -XPOST localhost:3000/api/dashboards/db/ @/dashbs/$d '
    ' -v
  done
}

create_datasource &
create_dashboard &

exec /run.sh