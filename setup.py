import setuptools

with open("README.md", "r", encoding="utf-8") as fh:
    long_description = fh.read()

setuptools.setup(
    name="wework",
    version="0.1.0",
    author="David Mohl",
    author_email="git@d.sh",
    description="A CLI tool for booking WeWork spaces",
    long_description=long_description,
    long_description_content_type="text/markdown",
    url="https://github.com/dvcrn/wework",
    packages=setuptools.find_packages(),
    classifiers=[
        "Programming Language :: Python :: 3",
        "License :: OSI Approved :: MIT License",
        "Operating System :: OS Independent",
    ],
    python_requires=">=3.6",
    install_requires=[
        "requests>=2.26.0",
        "argparse>=1.4.0",
        "beautifulsoup4>=4.12.3",
        "icalendar>=5.0.0",
    ],
    entry_points={
        "console_scripts": [
            "wework=cli:main",
        ],
    },
)
