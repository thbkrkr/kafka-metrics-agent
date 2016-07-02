FROM grafana/grafana

RUN apt-get update && apt-get -y install curl

COPY dashbs /dashbs
COPY docker-entrypoint.sh /docker-entrypoint.sh

ENTRYPOINT ["/docker-entrypoint.sh"]