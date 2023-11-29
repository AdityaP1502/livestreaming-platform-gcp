#!/bin/bash
# set -e

DB_INSTANCE_NAME="ltkalivestream"

gcloud sql instances patch $DB_INSTANCE_NAME --activation-policy=ALWAYS

while [[ $(gcloud sql instances describe $DB_INSTANCE_NAME --format="value(sqlServerDatabaseEngine.version)") == "SQL_SERVER_2019" ]]; do
    sleep 5
done

echo "Cloud SQL instance is now fully started."


create stream server managed instance groups
gcloud beta compute instance-groups managed create stream-mig \
--project=ltkalivestream \
--base-instance-name=stream-mig \
--size=0 \
--template=projects/ltkalivestream/global/instanceTemplates/stream-server-template \
--zone=asia-southeast2-a \
--list-managed-instances-results=PAGELESS \
--no-force-update-on-repair

gcloud beta compute instance-groups managed set-autoscaling stream-mig \
--project=ltkalivestream \
--zone=asia-southeast2-a \
--cool-down-period=60 \
--max-num-replicas=2 \
--min-num-replicas=0 \
--mode=on \
--target-cpu-utilization=0.6

# create transcoder managed instance gruops
gcloud beta compute instance-groups managed create transcoder-mig \
--project=ltkalivestream \
--base-instance-name=transcoder-mig \
--size=0 \
--template=projects/ltkalivestream/global/instanceTemplates/transcoder-server-template \
--zone=asia-southeast2-a \
--list-managed-instances-results=PAGELESS \
--no-force-update-on-repair \

gcloud beta compute instance-groups managed set-autoscaling transcoder-mig \
--project=ltkalivestream \
--zone=asia-southeast2-a \
--cool-down-period=60 \
--max-num-replicas=2 \
--min-num-replicas=0 \
--mode=on \
--target-cpu-utilization=0.8

# Create the load balancer using rest api
create_forwarding_rule() {
    local project_id="$1"
    local region="$2"
    local backend_service_name="$3"

    local response
    response=$(curl -X POST \
        -H "Authorization: Bearer $(gcloud auth print-access-token)" \
        -H "Content-Type: application/json" \
        -d @$backend_service_name-forwarding-rules.json \
        -s -o /dev/null -w "%{http_code}" \
        "https://compute.googleapis.com/compute/v1/projects/$project_id/regions/$region/forwardingRules")

    # echo $response
    if [ "$response" -eq 200 ]; then
        # The backend service exists
        return 0
    else
        # The backend service does not exist
        return 1
    fi
}

wait_for_forwarding_rule_creation() {
    local project_id="$1"
    local region="$2"
    local backend_service_name="$3"
    local max_retries="$4"
    local retry_delay="$5"

    local retries=0
    while [ "$retries" -lt "$max_retries" ]; do
        create_forwarding_rule "$project_id" "$region" "$backend_service_name"

        if [ $? -eq 0 ]; then
            echo "forwarding rule created successfully"
            break
        else
            retries=$((retries + 1))
            sleep "$retry_delay"
        fi
    done

    if [ "$retries" -eq "$max_retries" ]; then
        echo "Max retries reached. Backend service '$backend_service_name' was not created."
        exit 1
    fi
}

project_id="ltkalivestream"
region="asia-southeast2"
max_retries=10
retry_delay=10


# RTSP load
curl -X POST -H "Authorization: Bearer $(gcloud auth print-access-token)" -H "Content-Type: application/json" -d @rtsp-backend-service.json "https://compute.googleapis.com/compute/v1/projects/ltkalivestream/regions/asia-southeast2/backendServices"
# wait until load balancer is up

# Transcoder load balancer
curl -X POST -H "Authorization: Bearer $(gcloud auth print-access-token)" -H "Content-Type: application/json" -d @transcoder-backend-service.json "https://compute.googleapis.com/compute/v1/projects/ltkalivestream/regions/asia-southeast2/backendServices"
# wait_for_backend_service_creation "$project_id" "$region" "transcoder-lb" "$max_retries" "$retry_delay"

# curl -X POST -H "Authorization: Bearer $(gcloud auth print-access-token)" -H "Content-Type: application/json" -d @rtsp-forwarding-rules.json "https://compute.googleapis.com/compute/v1/projects/ltkalivestream/regions/asia-southeast2/forwardingRules"
# curl -X POST -H "Authorization: Bearer $(gcloud auth print-access-token)" -H "Content-Type: application/json" -d @transcoder-forwarding-rules.json "https://compute.googleapis.com/compute/v1/projects/ltkalivestream/regions/asia-southeast2/forwardingRules"

wait_for_forwarding_rule_creation "$project_id" "$region" "rtsp" "$max_retries" "$retry_delay"
wait_for_forwarding_rule_creation "$project_id" "$region" "transcoder" "$max_retries" "$retry_delay"
