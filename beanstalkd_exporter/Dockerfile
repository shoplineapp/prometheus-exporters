FROM seitk/beanstalkd_exporter:davidtannock

WORKDIR /

COPY docker-entrypoint.sh .

ENTRYPOINT ["/bin/sh"]
CMD ["docker-entrypoint.sh"]