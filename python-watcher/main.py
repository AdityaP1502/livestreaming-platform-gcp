import sys
import time
import logging
import getopt
import signal
import shutil

from watchdog.observers import Observer
from watchdog.events import LoggingEventHandler
from event_handler import GoogleStorageHandler, delete_blob

BUCKET_NAME="hls-stream-belajar-1-404607"

def sigterm_handler(signal, frame):
    print("[INFO] Received SIGTERM signal")
    sys.exit(0)

signal.signal(signalnum=signal.SIGTERM, handler=sigterm_handler)

if __name__ == "__main__":
    bucket_path=''
    path=''

    logging.basicConfig(level=logging.INFO,
                        format='%(asctime)s - %(message)s',
                        datefmt='%Y-%m-%d %H:%M:%S')

    # path = sys.argv[1] if len(sys.argv) > 1 else '.'
    # bucket_path = sys.argv[2] if len(sys.argv) > 2 else ''

    try:
        opts, args = getopt.getopt(sys.argv[1:], "hi:o:", ["input=", "output="])
    except getopt.GetoptError:
        print('script.py -i <inputfile> -o <outputfile> -v')
        sys.exit(2)

    print(opts)
    for opt, arg in opts:
        if opt == '-h':
            print('script.py -i <inputfile> -o <outputfile> -v')
            sys.exit()

        elif opt in ("-i", "--input"):
            path = arg

        elif opt in ("-o", "--output"):
            bucket_path = arg

    print(f'path:{path}')
    print(f'bucket_path:{bucket_path}')
    print(f'bucket_name:{BUCKET_NAME}')

    event_handler = LoggingEventHandler() if bucket_path == '' else \
        GoogleStorageHandler(bucket_name=BUCKET_NAME, bucket_path=bucket_path) 
    
    observer = Observer()
    observer.schedule(event_handler, path, recursive=True)
    observer.start()
    try:
        while True:
            time.sleep(1)
    finally:
        print("[INFO : Main] CTRL + C is pressed.")

        if type(event_handler).__name__ == GoogleStorageHandler.__name__:
            event_handler.stop_upload_hanlder()

        observer.stop()
        observer.join()

        # delete all the stream
        print(f"[INFO] Removing {path} directory and all of its contents") 
        shutil.rmtree(path, ignore_errors=True)

        # delete any stream in the bucket
        delete_blob(bucket_name=BUCKET_NAME, blob_directory=bucket_path)
        
