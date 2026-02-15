FROM redhat/ubi8-minimal AS builder

RUN microdnf update && \
    microdnf install jq gettext unzip curl && \
    microdnf clean all

FROM confluentinc/cp-kafka-connect-base:7.9.1

# last from https://www.confluent.io/hub/debezium/debezium-connector-postgresql
ENV POSTGRESQL_CONNECTOR_VERSION=3.1.2
# last from https://www.confluent.io/hub/confluentinc/kafka-connect-elasticsearch
ENV ELASTICSEARCH_CONNECTOR_VERSION=15.1.1
ENV CONFLUENT_HUB_DIR=/usr/share/confluent-hub-components

COPY --from=builder /usr/lib64/libjq.so.1 /usr/lib64/libjq.so.1
COPY --from=builder /usr/lib64/libonig.so.5 /usr/lib64/libonig.so.5
COPY --from=builder /usr/bin/jq /usr/bin/jq
COPY --from=builder /usr/bin/envsubst /usr/bin/envsubst
COPY --from=builder /usr/bin/curl /usr/bin/curl
COPY --from=builder /usr/bin/unzip /usr/bin/unzip

RUN confluent-hub install --no-prompt debezium/debezium-connector-postgresql:${POSTGRESQL_CONNECTOR_VERSION} --component-dir ${CONFLUENT_HUB_DIR}
RUN confluent-hub install --no-prompt confluentinc/kafka-connect-elasticsearch:${ELASTICSEARCH_CONNECTOR_VERSION} --component-dir ${CONFLUENT_HUB_DIR}

COPY deployments/kafka-connect/probe.py /home/appuser/
