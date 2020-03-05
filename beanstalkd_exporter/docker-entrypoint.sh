if [[ "${BEANSTALKD_ALLTUBES}" == "true" ]]; then
  echo "ALLTUBE"
  ./beanstalkd_exporter --beanstalkd.address=${BEANSTALKD_SERVER} --beanstalkd.allTubes
elif [[ "${BEANSTALKD_ALLTUBES}" == "false" ]]; then
  echo "SPECIFY_TUBE"
  ./beanstalkd_exporter --beanstalkd.address=${BEANSTALKD_SERVER} --beanstalkd.tubes=${BEANSTALKD_TUBES}
else
  echo "environment variable 'BEANSTALKD_ALLTUBES' not set, must be 'true' or 'false'"
  exit 1
fi