from setuptools import setup, find_packages

setup(
    name="chaossql",
    version="1.4.0",
    packages=find_packages(),
    entry_points={
        "pytest11": [
            "chaossql = chaossql.pytest_plugin",
        ],
    },
)
