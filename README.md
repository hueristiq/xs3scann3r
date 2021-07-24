# sigs3scann3r

[![release](https://img.shields.io/github/release/signedsecurity/sigs3scann3r?style=flat&color=0040ff)](https://github.com/signedsecurity/sigs3scann3r/releases) ![maintenance](https://img.shields.io/badge/maintained%3F-yes-0040ff.svg) [![open issues](https://img.shields.io/github/issues-raw/signedsecurity/sigs3scann3r.svg?style=flat&color=0040ff)](https://github.com/signedsecurity/sigs3scann3r/issues?q=is:issue+is:open) [![closed issues](https://img.shields.io/github/issues-closed-raw/signedsecurity/sigs3scann3r.svg?style=flat&color=0040ff)](https://github.com/signedsecurity/sigs3scann3r/issues?q=is:issue+is:closed) [![license](https://img.shields.io/badge/license-MIT-gray.svg?colorB=0040FF)](https://github.com/signedsecurity/sigs3scann3r/blob/master/LICENSE) [![twitter](https://img.shields.io/badge/twitter-@signedsecurity-0040ff.svg)](https://twitter.com/signedsecurity)

sigs3scann3r is tool to scan AWS S3 bucket permissions.

## Resources

* [Features](#features)
* [Installation](#installation)
	* [From Binary](#from-binary)
	* [From source](#from-source)
	* [From github](#from-github)
* [Usage](#usage)
	* [Interpreting Results](#interpreting-results)
* [Contribution](#contribution)

## Features

* Scans all bucket permissions to find misconfigurations

## Installation

#### From Binary

You can download the pre-built binary for your platform from this repository's [releases](https://github.com/signedsecurity/sigs3scann3r/releases/) page, extract, then move it to your `$PATH`and you're ready to go.

#### From Source

sigs3scann3r requires **go1.14+** to install successfully. Run the following command to get the repo

```bash
GO111MODULE=on go get -u -v github.com/signedsecurity/sigs3scann3r/cmd/sigs3scann3r
```

#### From Github

```bash
git clone https://github.com/signedsecurity/sigs3scann3r.git && \
cd sigs3scann3r/cmd/sigs3scann3r/ && \
go build . && \
mv sigs3scann3r /usr/local/bin/ && \
sigs3scann3r -h
```

## Usage

> **NOTE:** To use this tool awscli is required to have been installed and configured.

To display help message for sigs3scann3r use the `-h` flag:

```
$ sigs3scann3r -h

     _           _____                           _____
 ___(_) __ _ ___|___ / ___  ___ __ _ _ __  _ __ |___ / _ __
/ __| |/ _` / __| |_ \/ __|/ __/ _` | '_ \| '_ \  |_ \| '__|
\__ \ | (_| \__ \___) \__ \ (_| (_| | | | | | | |___) | |
|___/_|\__, |___/____/|___/\___\__,_|_| |_|_| |_|____/|_| v1.0.0
       |___/

USAGE:
  sigs3scann3r [OPTIONS]

OPTIONS:
  -iL, --input-list   input buckets list (use `iL -` to read from stdin)
   -c, --concurrency  number of concurrent threads (default: 10)
  -nC, --no-color     no color mode (default: false)
   -v, --verbose      verbose mode

```

sigs3scann3r takes buckets in the format:

* Name - e.g. `flaws.cloud`
* URL style - e.g. `s3://flaws.cloud`
* Path style - e.g `https://s3.amazonaws.com/flaws.cloud`
* Virtual Hosted style - e.g `flaws.cloud.s3.amazonaws.com`

### Interpreting Results

[Possible permissions](https://docs.aws.amazon.com/AmazonS3/latest/userguide/managing-acls.html) for buckets:

* Read - List and view all files
* Write - Write files to bucket
* Read ACP - Read all Access Control Policies attached to bucket
* Write ACP - Write Access Control Policies to bucket
* Full Control - All above permissions

## Contribution

[Issues](https://github.com/signedsecurity/sigs3scann3r/issues) and [Pull Requests](https://github.com/signedsecurity/sigs3scann3r/pulls) are welcome!