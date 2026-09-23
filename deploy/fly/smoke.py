#!/usr/bin/env python3
"""Disposable end-to-end check of the managed image: Litestream replication to
MinIO, restore onto an empty volume, same-origin dashboard, and backup guard.
Creates only uniquely named resources and removes them at the end."""
import json, os, secrets, subprocess, time, urllib.request, urllib.error

# Build first: docker build -f deploy/fly/Dockerfile -t chaossql-cloud:local .
IMAGE = os.environ.get("CHAOSSQL_CLOUD_IMAGE", "chaossql-cloud:local")
tag = "a4smoke" + secrets.token_hex(4)
net, minio, app_a, app_b = f"{tag}-net", f"{tag}-minio", f"{tag}-a", f"{tag}-b"
vol_a, vol_b = f"{tag}-vol-a", f"{tag}-vol-b"
minio_user, minio_pass = "u" + secrets.token_hex(6), secrets.token_hex(16)
admin = "owner-" + secrets.token_hex(16)

def sh(*args, check=True):
    r = subprocess.run(list(args), capture_output=True, text=True)
    if check and r.returncode != 0:
        raise SystemExit(f"FAILED {' '.join(args)}\n{r.stdout}\n{r.stderr}")
    return r

def http(port, path, token=None, method="GET", body=None):
    req = urllib.request.Request(f"http://127.0.0.1:{port}{path}", method=method, data=body)
    if token:
        req.add_header("Authorization", "Bearer " + token)
    if body is not None:
        req.add_header("Content-Type", "application/json")
    try:
        with urllib.request.urlopen(req, timeout=5) as r:
            return r.status, r.read()
    except urllib.error.HTTPError as e:
        return e.code, e.read()

def wait_healthy(port):
    for _ in range(60):
        try:
            if http(port, "/v1/health")[0] == 200:
                return
        except Exception:
            pass
        time.sleep(1)
    raise SystemExit("server did not become healthy")

def app_env():
    return ["-e", f"CHAOSSQL_ADMIN_TOKEN={admin}", "-e", "LITESTREAM_BUCKET=chaossql",
            "-e", f"LITESTREAM_ENDPOINT=http://{minio}:9000",
            "-e", f"LITESTREAM_ACCESS_KEY_ID={minio_user}", "-e", f"LITESTREAM_SECRET_ACCESS_KEY={minio_pass}",
            "-e", "PUBLIC_URL=http://localhost"]

def start_app(name, vol):
    sh("docker", "run", "-d", "--name", name, "--network", net, "-v", f"{vol}:/data",
       "-p", "127.0.0.1::8080", *app_env(), IMAGE)
    port = sh("docker", "port", name, "8080").stdout.strip().split(":")[-1]
    wait_healthy(port)
    return port

try:
    sh("docker", "network", "create", net)
    sh("docker", "run", "-d", "--name", minio, "--network", net,
       "-e", f"MINIO_ROOT_USER={minio_user}", "-e", f"MINIO_ROOT_PASSWORD={minio_pass}",
       "quay.io/minio/minio:latest", "server", "/data")
    for _ in range(30):
        r = sh("docker", "run", "--rm", "--network", net, "--entrypoint", "sh", "quay.io/minio/mc:latest", "-c",
               f"mc alias set m http://{minio}:9000 {minio_user} {minio_pass} >/dev/null && mc mb -p m/chaossql", check=False)
        if r.returncode == 0:
            break
        time.sleep(1)
    else:
        raise SystemExit("could not create MinIO bucket")

    # Guard: no replication configured -> refuse to start.
    r = sh("docker", "run", "--rm", "-e", f"CHAOSSQL_ADMIN_TOKEN={admin}", IMAGE, check=False)
    assert r.returncode != 0 and "refusing to start without backups" in r.stderr, r
    print("PASS: refuses to start without Litestream configuration")

    port = start_app(app_a, vol_a)
    assert sh("docker", "exec", app_a, "id", "-u").stdout.strip() == "10001"
    status, index = http(port, "/")
    assert status == 200 and b"<html" in index.lower(), status
    assert http(port, "/wasm/chaossql.wasm")[0] == 200
    assert http(port, "/v1/runs")[0] == 401
    print("PASS: nonroot, same-origin dashboard and WASM, API auth")

    out = sh("docker", "exec", "-u", "0", app_a, "chaossql-admin", "org", "create", "--name", "Early Adopter", "--plan", "team").stdout
    owners = sh("docker", "exec", app_a, "sh", "-c", "stat -c '%U' /data/*").stdout.split()
    assert owners and set(owners) == {"chaossql"}, owners
    org_id = [l.split()[-1] for l in out.splitlines() if l.startswith("Organization:")][0]
    owner = [l.split()[-1] for l in out.splitlines() if l.startswith("Owner Token:")][0]
    status, body = http(port, "/v1/organizations/me/tokens", owner, "POST", json.dumps({"name": "CI"}).encode())
    assert status in (200, 201), (status, body)
    member = json.loads(body)["token"]
    status, body = http(port, "/v1/organizations/me/subscription", member)
    assert status == 200 and org_id.encode() in body, (status, body)
    print("PASS: tenant provisioned and member token issued through the API")

    time.sleep(5)  # allow Litestream to ship WAL segments
    # Restore drill while production is running: must not start replication.
    drill = sh("docker", "run", "--rm", "--network", net, *app_env(), "--entrypoint", "chaossql-restore-check", IMAGE).stdout
    assert org_id in drill and "Restore check succeeded" in drill, drill
    print("PASS: root-run admin keeps service ownership; restore drill reads replica without replicating")
    sh("docker", "stop", "-t", "30", app_a)
    logs = sh("docker", "logs", app_a).stderr + sh("docker", "logs", app_a).stdout
    assert "refusing" not in logs

    # Fresh machine, empty volume: must restore from the replica.
    port_b = start_app(app_b, vol_b)
    status, body = http(port_b, "/v1/organizations/me/subscription", member)
    assert status == 200 and org_id.encode() in body, (status, body)
    assert http(port_b, "/v1/organizations/me/subscription", owner)[0] == 200
    listing = sh("docker", "exec", app_b, "chaossql", "server", "org", "list").stdout
    assert org_id in listing and "Early Adopter" in listing
    print("PASS: empty volume restored from replica; tenant and tokens intact")
finally:
    for c in (app_a, app_b, minio):
        sh("docker", "rm", "-f", c, check=False)
    for v in (vol_a, vol_b):
        sh("docker", "volume", "rm", "-f", v, check=False)
    sh("docker", "network", "rm", net, check=False)
    print("Disposable containers, volumes and network removed.")
