#!/bin/bash

# place this on the same folder as mediamtx binary
echo $1
STREAM_LINK="rtsp://$VM_IP:8554/$1"
STORAGE_LINK="$1"
TMP_DIR="/usr/local/stream/tmp/$STORAGE_LINK"

mkdir -p $TMP_DIR

JSON_STRING=$( jq -n \
        --arg sl "$STREAM_LINK" \
        --arg stl "$STORAGE_LINK" \
        '{"stream-link": $sl, "storage-link": $stl}' )


echo "$JSON_STRING" > file.json
echo $TRANSCODER_LB_IP
echo $JSON_STRING
curl -X POST -H "Content-Type: application/json"  -d "$JSON_STRING" $TRANSCODER_LB_IP:8080/init | jq '.[].ip' > $TMP_DIR/ip
