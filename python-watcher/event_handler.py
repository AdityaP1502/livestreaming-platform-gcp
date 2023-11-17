"""Filesystem watcher handler to upload file to google bucket when
a modified event takes place
"""

import platform
import threading
from time import time
from watchdog.events import FileSystemEventHandler, FileSystemEvent, FileSystemMovedEvent
from google.cloud import storage

def upload_blob(bucket_name, source_file_name, destination_blob_name):
        """Uploads a file to the bucket."""
        # The ID of your GCS bucket
        # bucket_name = "your-bucket-name"
        # The path to your file to upload
        # source_file_name = "local/path/to/file"
        # The ID of your GCS object
        # destination_blob_name = "storage-object-name"

        storage_client = storage.Client()
        bucket = storage_client.bucket(bucket_name)
        blob = bucket.blob(destination_blob_name)

        # Optional: set a generation-match precondition to avoid potential race conditions
        # and data corruptions. The request to upload is aborted if the object's
        # generation number does not match your precondition. For a destination
        # object that does not yet exist, set the if_generation_match precondition to 0.
        # If the destination object already exists in your bucket, set instead a
        # generation-match precondition using its generation number.
        blob.upload_from_filename(source_file_name)

        print(
            f"File {source_file_name} uploaded to {destination_blob_name}."
        )


class UploadHandler():
    def __init__(self, max_concurrent_thread = 5) -> None:
        self.num_threads_active = 0
        self.max_concurrent_thread = max_concurrent_thread
        self._threads : list[threading.Thread] = []
        self._terminate = False
        self._watcher_t = threading.Thread(target=self._watcher, daemon=True)
        self._watcher_t.start()
        
    
    def _watcher(self):
        t : threading.Thread
        
        while not self._terminate:
            removed_idx = []
            for i in range(len(self._threads)):
                if not self._threads[i].is_alive:
                    removed_idx.append(i)
                
            for idx in removed_idx:
                del self._threads[idx]
                self.max_concurrent_thread -= 1

    def upload_file(self, bucket_name, source_file_name, destination_blob_name):
        while self.num_threads_active == self.max_concurrent_thread:
            continue

        t = threading.Thread(
            target=upload_blob,
            args=(bucket_name, source_file_name, destination_blob_name),
            daemon=True
        )
        
        t.start()
        self.max_concurrent_thread += 1
        self._threads.append(t)
            
class GoogleStorageHandler(FileSystemEventHandler):
    BUCKET_NAME = "hls-manifest"

    def __init__(self) -> None:
        super().__init__()
        self._upload_handler = UploadHandler()
        
    def on_modified(self, event : FileSystemEvent):
        if not event.is_directory:
            filename = event.src_path.split(
                '\\' if platform.system() == 'Windows' else '/')[-1]

            ext = filename.split('.')[-1]
            
            if ext != 'tmp':
                # upload_blob(
                #     bucket_name=self.BUCKET_NAME,
                #     source_file_name=event.src_path,
                #     destination_blob_name=f'stream/{filename}'
                # )
                self._upload_handler.upload_file(
                    bucket_name=self.BUCKET_NAME,
                    source_file_name=event.src_path,
                    destination_blob_name=f'stream/{filename}'
                )
                 
    def on_moved(self, event : FileSystemMovedEvent):
         if not event.is_directory:
            filename = event.dest_path.split(
                '\\' if platform.system() == 'Windows' else '/')[-1]

            ext = filename.split('.')[-1]
            
            if not ext == 'm3u8':
                return
            
            # upload_blob( 
            #     bucket_name=self.BUCKET_NAME,
            #     source_file_name=event.dest_path,
            #     destination_blob_name=f'stream/{filename}'
            # )
            
            self._upload_handler.upload_file(
                    bucket_name=self.BUCKET_NAME,
                    source_file_name=event.dest_path,
                    destination_blob_name=f'stream/{filename}'
            )