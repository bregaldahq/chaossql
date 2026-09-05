#!/usr/bin/env python3
"""
HTTP Server Verification Tool for ChaosSQL Web Portal (v1.2.0)
Tests ephemeral Python HTTP server and validates HTTP 200 OK + Content-Types
for:
  - /
  - /docs-data.js
  - /app.js
  - /assets/style.css
"""

import subprocess
import time
import sys
import urllib.request
import urllib.error
from pathlib import Path

PORT = 8087
BASE_URL = f"http://127.0.0.1:{PORT}"
SITE_DIR = Path(__file__).resolve().parent.parent / "site"

ENDPOINTS = [
    ("/", "text/html"),
    ("/docs-data.js", "javascript"),
    ("/app.js", "javascript"),
    ("/assets/style.css", "text/css"),
]

def main() -> int:
    print("=" * 65)
    print(f"  ChaosSQL v1.2.0 — Local HTTP Server QA Verification (Port {PORT})")
    print("=" * 65)

    cmd = [sys.executable, "-m", "http.server", str(PORT), "--directory", str(SITE_DIR)]
    print(f"Starting server: {' '.join(cmd)}")
    server_proc = subprocess.Popen(cmd, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)

    try:
        # Wait for server to become responsive
        max_attempts = 20
        connected = False
        for i in range(max_attempts):
            try:
                with urllib.request.urlopen(f"{BASE_URL}/", timeout=1) as resp:
                    if resp.status == 200:
                        connected = True
                        break
            except Exception:
                time.sleep(0.2)

        if not connected:
            print("[FAIL] Server failed to start and respond on port", PORT, file=sys.stderr)
            return 1

        print(f"[OK] Server is listening on {BASE_URL}\n")

        all_passed = True
        curl_outputs = {}

        for path, expected_content_type in ENDPOINTS:
            url = f"{BASE_URL}{path}"
            print(f"--- Testing {path} via curl -I ---")
            
            # Run curl -I
            curl_res = subprocess.run(["curl", "-I", "-s", url], capture_output=True, text=True)
            raw_header = curl_res.stdout.strip()
            curl_outputs[path] = raw_header
            print(raw_header)

            # Validate HTTP 200 OK
            first_line = raw_header.splitlines()[0] if raw_header.splitlines() else ""
            if "200 OK" not in first_line:
                print(f"[FAIL] {path} did not return 200 OK: {first_line}", file=sys.stderr)
                all_passed = False
                continue

            # Validate Content-Type
            content_type_line = ""
            for line in raw_header.splitlines():
                if line.lower().startswith("content-type:"):
                    content_type_line = line
                    break

            if expected_content_type.lower() not in content_type_line.lower():
                print(f"[FAIL] {path} expected content-type '{expected_content_type}', got '{content_type_line}'", file=sys.stderr)
                all_passed = False
            else:
                print(f"[PASS] {path} -> HTTP 200 OK | {content_type_line}\n")

        if all_passed:
            print("=" * 65)
            print("  ALL HTTP ENDPOINT CHECKS PASSED (4/4 endpoints OK)")
            print("=" * 65)
            return 0
        else:
            print("=" * 65)
            print("  SOME HTTP CHECKS FAILED", file=sys.stderr)
            print("=" * 65)
            return 1

    finally:
        server_proc.terminate()
        try:
            server_proc.wait(timeout=2)
        except subprocess.TimeoutExpired:
            server_proc.kill()
        print("Server stopped cleanly.")

if __name__ == "__main__":
    sys.exit(main())
