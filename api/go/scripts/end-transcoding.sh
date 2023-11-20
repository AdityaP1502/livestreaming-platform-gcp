#!/bin/bash
STORAGE_LINK=$1

SOURCE=${BASH_SOURCE[0]}
while [ -L "$SOURCE" ]; do # resolve $SOURCE until the file is no longer a symlink
  DIR=$( cd -P "$( dirname "$SOURCE" )" >/dev/null 2>&1 && pwd )
  SOURCE=$(readlink "$SOURCE")
  [[ $SOURCE != /* ]] && SOURCE=$DIR/$SOURCE # if $SOURCE was a relative symlink, we need to resolve it relative to the path where the symlink file was located
done

DIR=$( cd -P "$( dirname "$SOURCE" )" >/dev/null 2>&1 && pwd )

# create a new directory
TEMP_DIR=$DIR/../../../tmp/$STORAGE_LINK

TRANSCODER_PID=$(cat $TEMP_DIR/transcoder.tmp)
WATCHER_PID=$(cat $TEMP_DIR/watcher.tmp)

echo $TRANSCODER_PID
echo $WATCHER_PID

kill $TRANSCODER_PID
kill $WATCHER_PID
