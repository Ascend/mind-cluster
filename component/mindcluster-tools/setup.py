from setuptools import setup, find_packages
from mindcluster_tools import __version__


install_requires = [
    "bitarray",
    "netifaces"
]


setup(
    name="mindcluster_tools",
    version=__version__,
    packages=find_packages(),
    description="",
    author="mindcluster",
    python_requires=">=3.6",
    install_requires=install_requires,
    entry_points={
        "console_scripts": [
            "mindcluster-tools=mindcluster_tools.tools_parser:main",
        ]
    }
)