project_id="ltkalivestream"
region="asia-southeast2"
max_retries=10
retry_delay=5

check_resource_exists() {
    local response
    local project_id="$1"
    local region="$2"
    local resource="$3"
    local resource_name="$4"

    response=$(curl -X GET \
        -H "Authorization: Bearer $(gcloud auth print-access-token)" \
        -s -o /dev/null -w "%{http_code}" \
        "https://compute.googleapis.com/compute/v1/projects/$project_id/regions/$region/$resource/$resource_name")

    echo "$response"
}

wait_for_resource_deletion() {
    local project_id="$1"
    local region="$2"
    local resource="$3"
    local resource_name="$4"
    local max_retries="$5"
    local retry_delay="$6"

    local retries=0
    while [ "$retries" -lt "$max_retries" ]; do
        response_code=$(check_resource_exists "$project_id" "$region" "$resource" "$resource_name")

        if [ "$response_code" -ne 200 ]; then
            echo "'$resource $resource_name' deleted (HTTP Status Code: $response_code)"
            break
        else
            echo "Retrying... (HTTP Status Code: $response_code)"
            retries=$((retries + 1))
            sleep "$retry_delay"
        fi
    done

    if [ "$retries" -eq "$max_retries" ]; then
        echo "Max retries reached. '$resource $resource_name' was not deleted."
        exit 1
    fi
}

gcloud compute instances stop public-api
gcloud sql instances patch ltkalivestream --activation-policy=NEVER

curl -X DELETE -H "Authorization: Bearer $(gcloud auth print-access-token)" "https://compute.googleapis.com/compute/v1/projects/ltkalivestream/regions/asia-southeast2/forwardingRules/stream-lb-forwarding-rule"
curl -X DELETE -H "Authorization: Bearer $(gcloud auth print-access-token)" "https://compute.googleapis.com/compute/v1/projects/ltkalivestream/regions/asia-southeast2/forwardingRules/transcoder-lb-forwarding-rule"


wait_for_resource_deletion "$project_id" "$region" "forwardingRules" "stream-lb-forwarding-rule" "$max_retries" "$retry_delay"
wait_for_resource_deletion "$project_id" "$region" "forwardingRules" "transcoder-lb-forwarding-rule" "$max_retries" "$retry_delay"


curl -X DELETE -H "Authorization: Bearer $(gcloud auth print-access-token)" "https://compute.googleapis.com/compute/v1/projects/ltkalivestream/regions/asia-southeast2/backendServices/stream-lb"
curl -X DELETE -H "Authorization: Bearer $(gcloud auth print-access-token)" "https://compute.googleapis.com/compute/v1/projects/ltkalivestream/regions/asia-southeast2/backendServices/transcoder-lb"

wait_for_resource_deletion "$project_id" "$region" "backendServices" "stream-lb" "$max_retries" "$retry_delay"
wait_for_resource_deletion "$project_id" "$region" "backendServices" "transcoder-lb" "$max_retries" "$retry_delay"

gcloud compute instance-groups managed delete stream-mig --zone=asia-southeast2-a
gcloud compute instance-groups managed delete transcoder-mig --zone=asia-southeast2-a
