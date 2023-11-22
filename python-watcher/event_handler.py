"""Filesystem watcher handler to upload file to google bucket when
a modified event takes place
"""

import platform
import threading
from multiprocessing import Value

from google.cloud import storage
from watchdog.events import FileSystemEventHandler, FileSystemEvent, FileSystemMovedEvent

def delete_blob(bucket_name, blob_directory):
    client = storage.Client()
    bucket = client.get_bucket(bucket_name)

    blobs = bucket.list_blobs(prefix=blob_directory)

    for blob in blobs:
        blob.delete()

    
def upload_blob(bucket_name, source_file_name, destination_blob_name):
    """Uploads a file to the bucket."""
    # The ID of your GCS bucket
    # bucket_name = "your-bucket-name"
    # The path to your file to upload
    # source_file_name = "local/path/to/file"
    # The ID of your GCS object
    # destination_blob_name = "storage-object-name"
    print(f"[INFO] Received upload job {source_file_name} -> {destination_blob_name}")
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
        f"[INFO] File {source_file_name} uploaded to {destination_blob_name}."
    )

class UploaderThread(threading.Thread):
    def __init__(self, group = None, target = None, name = None, args = ..., kwargs = None, *, daemon = None) -> None:
        super().__init__(group, target, name, args, kwargs, daemon=daemon)
        self.exc : BaseException = None
    
    def join(self, timeout = None) -> None:
        super().join(timeout)
        
        if self.exc is not None:
            raise self.exc
    
    def run(self) -> None:
        self.exc = None
        
        try:
            self.ret = self._target(*self._args, **self._kwargs)
        except BaseException as e:
            self.exc = e

class UploadHandler():
    def __init__(self, max_concurrent_thread = 5) -> None:
        self.num_threads_active = 0
        self.max_concurrent_thread = max_concurrent_thread
        self._threads : list[UploaderThread] = []
        self._terminate = Value("i", 0)
        self._watcher_t = threading.Thread(target=self._watcher, daemon=True)
        self._watcher_t.start()
        print(type(self._terminate))
    
    def _watcher(self):
        
        while True:
            removed_idx = []
            for i, t in enumerate(self._threads):
                if not t.is_alive():
                    removed_idx.append(i)
                    
            removed_idx.reverse()
            
            for idx in removed_idx:
                if self._threads[idx].exc is not None:
                    print(f"[ERROR] Uploading files failed.\n{self._threads[idx].exc}")
                    
                del self._threads[idx]
                self.max_concurrent_thread -= 1
                
            with self._terminate.get_lock():
                if self._terminate.value == 1:
                    break
        
        print("[INFO : Watcher] Joining spawned thread")        
        
        for i, t in enumerate(self._threads):
            print(f"[INFO : Watcher] Joining thread - {i + 1} / {len(self._threads)}")
            
            try:
                t.join(timeout=10)
            except:
                pass
            
        print("[INFO : Watcher] All thread are closed successfully")

    def upload_file(self, bucket_name, source_file_name, destination_blob_name):
        while (self.num_threads_active == self.max_concurrent_thread):
            
            with self._terminate.get_lock():
                if self._terminate.value == 1:
                    return
                
            continue
        
        
        t = UploaderThread(
            target=upload_blob,
            args=(bucket_name, source_file_name, destination_blob_name),
            daemon=True
        )
        
        t.start()
        self.num_threads_active += 1
        self._threads.append(t)

    def stop(self):
        with self._terminate.get_lock():
            self._terminate.value += 1
            
        self._watcher_t.join()
        print("[INFO : Upload Handler] All thread are closed successfully")

class GoogleStorageHandler(FileSystemEventHandler):
    def __init__(self, bucket_name, bucket_path) -> None:
        super().__init__()

        if bucket_name == "":
            raise ValueError("Bucket name is an empty string.")
        
        if bucket_path == "" :
            raise ValueError("Bucket path is an empty string.")

        self._upload_handler = UploadHandler()

        self._bucket_name = bucket_name
        self._bucket_path = bucket_path

        

    def stop_upload_hanlder(self):
        try:
            self._upload_handler.stop()
        except:
            pass
        
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
                    bucket_name=self._bucket_name,
                    source_file_name=event.src_path,
                    destination_blob_name=f'{self._bucket_path}/{filename}' # stream/username/streamid/{filename}
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
                    bucket_name=self._bucket_name,
                    source_file_name=event.dest_path,
                    destination_blob_name=f'{self._bucket_path}/{filename}' # stream/username/streamid/{filename}
            )