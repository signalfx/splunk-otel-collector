#!/usr/bin/env python3
"""Minimal HTTP server that responds with a fixed status code and empty JSON body.

Usage: python3 mock-http-server.py <status_code> <port>
"""
import http.server, sys

status = int(sys.argv[1])
port = int(sys.argv[2])

class Handler(http.server.BaseHTTPRequestHandler):
    def do_POST(self):
        self.send_response(status)
        self.send_header("Content-Type", "application/json")
        self.end_headers()
        self.wfile.write(b"{}")

    def log_message(self, *a):
        pass

http.server.HTTPServer(("127.0.0.1", port), Handler).serve_forever()
