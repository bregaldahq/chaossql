import os
import sys

# Ensure sdks/python is on sys.path
SDK_DIR = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
if SDK_DIR not in sys.path:
    sys.path.insert(0, SDK_DIR)

pytest_plugins = ["chaossql.pytest_plugin"]
