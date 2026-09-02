import sys, pathlib
p = pathlib.Path(sys.argv[1])
p.parent.mkdir(parents=True, exist_ok=True)
p.write_bytes(bytes.fromhex(sys.argv[2]))
print('WSL Saved:', sys.argv[1])
