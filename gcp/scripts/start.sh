#!/bin/bash

# place this on the same folder as mediamtx binary
echo $1
STREAM_LINK="rtsp://$VM_IP:8554/$1"
STORAGE_LINK="$1"

JSON_STRING=$( jq -n \
        --arg sl "$STREAM_LINK" \
        --arg stl "$STORAGE_LINK" \
        '{"stream-link": $sl, "storage-link": $stl}' )


echo "$JSON_STRING" > file.json
echo $TRANSCODER_LB_IP
echo $JSON_STRING
curl -X POST -H "Content-Type: application/json"  -d "$JSON_STRING" $TRANSCODER_LB_IP:8080/init
