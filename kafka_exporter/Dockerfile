FROM danielqsj/kafka-exporter

ARG KAFKA_BOOTSTRAP_SERVER=""
ARG KAFKA_API_KEY=""
ARG KAFKA_API_SECRET=""

ENV KAFKA_BOOTSTRAP_SERVER=$KAFKA_BOOTSTRAP_SERVER
ENV KAFKA_API_KEY=$KAFKA_API_KEY
ENV KAFKA_API_SECRET=$KAFKA_API_SECRET

ENTRYPOINT []

CMD kafka_exporter --kafka.server=$KAFKA_BOOTSTRAP_SERVER --sasl.username=$KAFKA_API_KEY --sasl.password=$KAFKA_API_SECRET --sasl.enabled --sasl.handshake --tls.insecure-skip-tls-verify --tls.enabled
