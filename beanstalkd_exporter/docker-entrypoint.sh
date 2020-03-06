if [[ -z "$BEANSTALKD_TUBES" ]]; then
  TUBE_ARG="--beanstalkd.allTubes"
else
  TUBE_ARG="--beanstalkd.tubes=$BEANSTALKD_TUBES"
fi

./beanstalkd_exporter --beanstalkd.address="${BEANSTALKD_SERVER}" $TUBE_ARG