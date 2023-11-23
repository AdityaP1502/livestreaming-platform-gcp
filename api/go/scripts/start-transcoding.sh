#!/bin/bash
STREAM_LINK=$1
STORAGE_LINK=$2

SOURCE=${BASH_SOURCE[0]}
while [ -L "$SOURCE" ]; do # resolve $SOURCE until the file is no longer a symlink
  DIR=$( cd -P "$( dirname "$SOURCE" )" >/dev/null 2>&1 && pwd )
  SOURCE=$(readlink "$SOURCE")
  [[ $SOURCE != /* ]] && SOURCE=$DIR/$SOURCE # if $SOURCE was a relative symlink, we need to resolve it relative to the path where the symlink file was located
done

DIR=$( cd -P "$( dirname "$SOURCE" )" >/dev/null 2>&1 && pwd )

# create a new directory
STREAM_DIR=/usr/local/transcoder/$STORAGE_LINK
LOG_DIR=/usr/local/transcoder/logs/$STORAGE_LINK
TEMP_DIR=/usr/local/transcoder/temp/$STORAGE_LINK

echo "The stream will be placed in $STREAM_DIR"
echo "The log files can be found in $LOG_DIR"

mkdir -p $STREAM_DIR
mkdir -p $LOG_DIR
mkdir -p $LOG_DIR/watcher
mkdir -p $TEMP_DIR

# Run watcher
pushd /usr/local/python-watcher

# echo python3 main.py --input=$STREAM_DIR --output=$STORAGE_LINK

python3 main.py --input=$STREAM_DIR --output=$STORAGE_LINK > $LOG_DIR/watcher/output.log 2>&1 < /dev/null &
WATCHER_PID=$!

# run ffmpeg 
ffmpeg -nostdin -i $STREAM_LINK \
  -vsync 0 \
  -copyts \
  -c:v copy \
  -c:a copy \
  -f hls \
  -hls_time 1 \
  -hls_list_size 3 \
  -hls_segment_type mpegts \
  -hls_segment_filename $STREAM_DIR/%d.ts \
  $STREAM_DIR/index.m3u8 > $LOG_DIR/output.log 2>&1 < /dev/null &

TRANSCODER_PID=$!

echo "Running two process in background with PID:"
echo "[w] $WATCHER_PID [t] $TRANSCODER_PID"

echo $WATCHER_PID > $TEMP_DIR/watcher.tmp
echo $TRANSCODER_PID > $TEMP_DIR/transcoder.tmp

popd


