import os
import http.server
import socketserver
from http.server import SimpleHTTPRequestHandler
PORT = 8000

class Handler(http.server.SimpleHTTPRequestHandler):
    def __init__(self, *args, **kwargs):
        self.extensions_map.update({
            '.m3u8': 'application/x-mpegURL',
            '.ts': 'video/MP2T',
        })
        super().__init__(*args, **kwargs)
    def end_headers(self):
        self.send_header('Access-Control-Allow-Origin', '*')
        SimpleHTTPRequestHandler.end_headers(self)

os.chdir("../")
print(os.getcwd())

httpd = socketserver.TCPServer(("", PORT), Handler)

print(f"Serving on http://localhost:{PORT}")
httpd.serve_forever()

