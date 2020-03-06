FROM seitk/beanstalkd_exporter:davidtannock

WORKDIR /

ARG BEANSTALKD_SERVER=""
ARG BEANSTALKD_TUBES=""

ENV BEANSTALKD_SERVER=$BEANSTALKD_SERVER
ENV BEANSTALKD_TUBES=$BEANSTALKD_TUBES

COPY docker-entrypoint.sh .

ENTRYPOINT ["/bin/sh"]
CMD ["docker-entrypoint.sh"]