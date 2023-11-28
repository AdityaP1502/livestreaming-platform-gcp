project_id="ltkalivestream"
region="asia-southeast2"
max_retries=10
retry_delay=5

check_forwarding_rule_exists() {
    local response
    local project_id="$1"
    local region="$2"
    local forwarding_rule_name="$3"

    response=$(curl -X GET \
        -H "Authorization: Bearer $(gcloud auth print-access-token)" \
        -s -o /dev/null -w "%{http_code}" \
        "https://compute.googleapis.com/compute/v1/projects/$project_id/regions/$region/forwardingRules/$forwarding_rule_name")

    echo "$response"
}

wait_for_forwarding_rule_deletion() {
    local project_id="$1"
    local region="$2"
    local forwarding_rule_name="$3"
    local max_retries="$4"
    local retry_delay="$5"

    local retries=0
    while [ "$retries" -lt "$max_retries" ]; do
        response_code=$(check_forwarding_rule_exists "$project_id" "$region" "$forwarding_rule_name")

        if [ "$response_code" -ne 200 ]; then
            echo "Forwarding rule '$forwarding_rule_name' deleted (HTTP Status Code: $response_code)"
            break
        else
            echo "Retrying... (HTTP Status Code: $response_code)"
            retries=$((retries + 1))
            sleep "$retry_delay"
        fi
    done

    if [ "$retries" -eq "$max_retries" ]; then
        echo "Max retries reached. Forwarding rule '$forwarding_rule_name' was not deleted."
        exit 1
    fi
}

gcloud compute instances stop public-api
gcloud sql instances patch ltkalivestream --activation-policy=NEVER

curl -X DELETE -H "Authorization: Bearer $(gcloud auth print-access-token)" "https://compute.googleapis.com/compute/v1/projects/ltkalivestream/regions/asia-southeast2/forwardingRules/stream-lb-forwarding-rule"
curl -X DELETE -H "Authorization: Bearer $(gcloud auth print-access-token)" "https://compute.googleapis.com/compute/v1/projects/ltkalivestream/regions/asia-southeast2/forwardingRules/transcoder-lb-forwarding-rule"


wait_for_forwarding_rule_deletion "$project_id" "$region" "stream-lb-forwarding-rule" "$max_retries" "$retry_delay"
wait_for_forwarding_rule_deletion "$project_id" "$region" "transcoder-lb-forwarding-rule" "$max_retries" "$retry_delay"


curl -X DELETE -H "Authorization: Bearer $(gcloud auth print-access-token)" "https://compute.googleapis.com/compute/v1/projects/ltkalivestream/regions/asia-southeast2/backendServices/stream-lb"
curl -X DELETE -H "Authorization: Bearer $(gcloud auth print-access-token)" "https://compute.googleapis.com/compute/v1/projects/ltkalivestream/regions/asia-southeast2/backendServices/transcoder-lb"

gcloud compute instance-groups managed delete stream-mig --zone=asia-southeast2-a
gcloud compute instance-groups managed delete transcoder-mig --zone=asia-southeast2-a
