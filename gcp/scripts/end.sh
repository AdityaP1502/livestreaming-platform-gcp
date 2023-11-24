#!/bin/bash
echo $1
STREAM_LINK="rtsp://$VM_IP:8554/$1"
STORAGE_LINK="$1"
TMP_DIR="/usr/local/stream/tmp/$STORAGE_LINK"

JSON_STRING=$( jq -n \
        --arg sl "$STREAM_LINK" \
        --arg stl "$STORAGE_LINK" \
        '{"stream-link": $sl, "storage-link": $stl}' )

TRANSCODER_IP=$(cat $TMP_DIR/ip)

curl -X POST -H "Content-Type: application/json"  -d "$JSON_STRING" $TRANSCODER_IP:8080/end > /dev/null
