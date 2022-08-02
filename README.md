# hqs3scann3r

[![release](https://img.shields.io/github/release/hueristiq/hqs3scann3r?style=flat&color=0040ff)](https://github.com/hueristiq/hqs3scann3r/releases) ![maintenance](https://img.shields.io/badge/maintained%3F-yes-0040ff.svg) [![open issues](https://img.shields.io/github/issues-raw/hueristiq/hqs3scann3r.svg?style=flat&color=0040ff)](https://github.com/hueristiq/hqs3scann3r/issues?q=is:issue+is:open) [![closed issues](https://img.shields.io/github/issues-closed-raw/hueristiq/hqs3scann3r.svg?style=flat&color=0040ff)](https://github.com/hueristiq/hqs3scann3r/issues?q=is:issue+is:closed) [![license](https://img.shields.io/badge/license-MIT-gray.svg?colorB=0040FF)](https://github.com/hueristiq/hqs3scann3r/blob/master/LICENSE) [![twitter](https://img.shields.io/badge/twitter-@itshueristiq-0040ff.svg)](https://twitter.com/itshueristiq)

hqs3scann3r is tool to scan AWS S3 bucket permissions.

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

You can download the pre-built binary for your platform from this repository's [releases](https://github.com/hueristiq/hqs3scann3r/releases/) page, extract, then move it to your `$PATH`and you're ready to go.

#### From Source

hqs3scann3r requires **go1.17+** to install successfully. Run the following command to get the repo

```bash
go install -v github.com/hueristiq/hqs3scann3r/cmd/hqs3scann3r@latest
```

#### From Github

```bash
git clone https://github.com/hueristiq/hqs3scann3r.git && \
cd hqs3scann3r/cmd/hqs3scann3r/ && \
go build . && \
mv hqs3scann3r /usr/local/bin/ && \
hqs3scann3r -h
```

## Usage

> **NOTE:** To use this tool awscli is required to have been installed and configured.

To display help message for hqs3scann3r use the `-h` flag:

```
hqs3scann3r -h
```

```
 _               _____                           _____      
| |__   __ _ ___|___ / ___  ___ __ _ _ __  _ __ |___ / _ __ 
| '_ \ / _` / __| |_ \/ __|/ __/ _` | '_ \| '_ \  |_ \| '__|
| | | | (_| \__ \___) \__ \ (_| (_| | | | | | | |___) | |   
|_| |_|\__, |___/____/|___/\___\__,_|_| |_|_| |_|____/|_| v1.1.0
          |_|

USAGE:
  hqs3scann3r [OPTIONS]

OPTIONS:
   -c, --concurrency  number of concurrent threads (default: 10)
   -d, --dump         location to dump objects
  -iL, --input-list   buckets list (use `-iL -` to read from stdin)
  -nC, --no-color     no color mode (default: false)
   -v, --verbose      verbose mode
```

hqs3scann3r takes buckets in the format:

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

[Issues](https://github.com/hueristiq/hqs3scann3r/issues) and [Pull Requests](https://github.com/hueristiq/hqs3scann3r/pulls) are welcome!