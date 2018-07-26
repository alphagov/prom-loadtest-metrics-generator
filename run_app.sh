#!/bin/bash

get_scrape_config() {
# returns the scrape config for an index value
echo "
  - job_name: load-test-targets-${1}
    metrics_path: \"/metrics/${1}\"
    scheme: https
    scrape_interval: ${2}s
    static_configs:
      - targets: [\"test-prom-loadtest-metrics-generator.cloudapps.digital\"]
"
}

create_scrape_config() {
# Creates scrape config for use in the prometheus-aws-configuration-beta

FILENAME="outputs/scrape-${ENDPOINT_COUNT}-${SCRAPE_INTERVAL}.txt"

SCRAPE_CONFIG=''

for index in $( seq 1 ${ENDPOINT_COUNT} )
do
SCRAPE_CONFIG+=$(get_scrape_config ${index} ${SCRAPE_INTERVAL})
done

cat <<EOF >${FILENAME}
$SCRAPE_CONFIG
EOF

echo "#### Append the contents of ${FILENAME} to the scrape_configs section in terraform/app-ecs-services/templates/prometheus.tpl"

}

start_webserver() {
    go run *.go
}

### Set default behaviour to run the webserver

if [ "$1" ] ; then 
    TASK="$1"
else
    echo "other options available:"
    echo "./run_app.sh -w <ENDPOINT_COUNT> <ROUTE_COUNT>  # default task, run the webserver with optional parameters"
    echo "./run_app.sh -c <ENDPOINT_COUNT> <SCRAPE_INTERVAL>  # create the scrape config with optional parameters"

    TASK="-w"
fi

### Set ENDPOINT_COUNT or ROUTE_COUNT to override environment variables

if [ "$2" ] ; then
    ENDPOINT_COUNT=$2
fi

### Select shell task to run

case "$TASK" in

-c) 
    if [ "$3" ] ; then
        SCRAPE_INTERVAL=$3
    fi
    echo "Create scrape config files with ${ENDPOINT_COUNT} endpoints, ${SCRAPE_INTERVAL}s scrape interval"

    create_scrape_config
;;

-w) 
    if [ "$3" ] ; then
        ROUTE_COUNT=$3
    fi

    echo "Start web server with ${ENDPOINT_COUNT} endpoints, ${ROUTE_COUNT} routes simulated"
    start_webserver
;;

esac
