# xs3scann3r

![made with go](https://img.shields.io/badge/made%20with-Go-0000FF.svg) [![release](https://img.shields.io/github/release/hueristiq/xs3scann3r?style=flat&color=0000FF)](https://github.com/hueristiq/xs3scann3r/releases) [![license](https://img.shields.io/badge/license-MIT-gray.svg?color=0000FF)](https://github.com/hueristiq/xs3scann3r/blob/master/LICENSE) ![maintenance](https://img.shields.io/badge/maintained%3F-yes-0000FF.svg) [![open issues](https://img.shields.io/github/issues-raw/hueristiq/xs3scann3r.svg?style=flat&color=0000FF)](https://github.com/hueristiq/xs3scann3r/issues?q=is:issue+is:open) [![closed issues](https://img.shields.io/github/issues-closed-raw/hueristiq/xs3scann3r.svg?style=flat&color=0000FF)](https://github.com/hueristiq/xs3scann3r/issues?q=is:issue+is:closed) [![contribution](https://img.shields.io/badge/contributions-welcome-0000FF.svg)](https://github.com/hueristiq/xs3scann3r/blob/master/CONTRIBUTING.md)

`xs3scann3r` is a command-line interface (CLI) utility to scan S3 bucket permissions.

## Resources

* [Features](#features)
* [Installation](#installation)
	* [Install release binaries](#install-release-binaries)
	* [Install source](#install-sources)
		* [`go install ...`](#go-install)
		* [`go build ...` the development Version](#go-build--the-development-version)
* [Usage](#usage)
	* [Interpreting Results](#interpreting-results)
* [Contribution](#contribution)
* [Licensing](#licensing)

## Features

* Scans all bucket permissions to find misconfigurations

## Installation

### Install release binaries

Visit the [releases page](https://github.com/hueristiq/xs3scann3r/releases) and find the appropriate archive for your operating system and architecture. Download the archive from your browser or copy its URL and retrieve it with `wget` or `curl`:

* ...with `wget`:

	```bash
	wget https://github.com/hueristiq/xs3scann3r/releases/download/v<version>/xs3scann3r-<version>-linux-amd64.tar.gz
	```

* ...or, with `curl`:

	```bash
	curl -OL https://github.com/hueristiq/xs3scann3r/releases/download/v<version>/xs3scann3r-<version>-linux-amd64.tar.gz
	```

...then, extract the binary:

```bash
tar xf xs3scann3r-<version>-linux-amd64.tar.gz
```

> **TIP:** The above steps, download and extract, can be combined into a single step with this onliner
> 
> ```bash
> curl -sL https://github.com/hueristiq/xs3scann3r/releases/download/v<version>/xs3scann3r-<version>-linux-amd64.tar.gz | tar -xzv
> ```

**NOTE:** On Windows systems, you should be able to double-click the zip archive to extract the `xs3scann3r` executable.

...move the `xs3scann3r` binary to somewhere in your `PATH`. For example, on GNU/Linux and OS X systems:

```bash
sudo mv xs3scann3r /usr/local/bin/
```

**NOTE:** Windows users can follow [How to: Add Tool Locations to the PATH Environment Variable](https://msdn.microsoft.com/en-us/library/office/ee537574(v=office.14).aspx) in order to add `xs3scann3r` to their `PATH`.

### Install source

Before you install from source, you need to make sure that Go is installed on your system. You can install Go by following the official instructions for your operating system. For this, we will assume that Go is already installed.

#### `go install ...`

```bash
go install -v github.com/hueristiq/xs3scann3r/cmd/xs3scann3r@latest
```

#### `go build ...` the development Version

* Clone the repository

	```bash
	git clone https://github.com/hueristiq/xs3scann3r.git 
	```

* Build the utility

	```bash
	cd xs3scann3r/cmd/xs3scann3r && \
	go build .
	```

* Move the `xs3scann3r` binary to somewhere in your `PATH`. For example, on GNU/Linux and OS X systems:

	```bash
	sudo mv xs3scann3r /usr/local/bin/
	```

	**NOTE:** Windows users can follow [How to: Add Tool Locations to the PATH Environment Variable](https://msdn.microsoft.com/en-us/library/office/ee537574(v=office.14).aspx) in order to add `xs3scann3r` to their `PATH`.


**NOTE:** While the development version is a good way to take a peek at `xs3scann3r`'s latest features before they get released, be aware that it may have bugs. Officially released versions will generally be more stable.

## Usage

> **NOTE:** To use this tool awscli is required to have been installed and configured.

To display help message for xs3scann3r use the `-h` flag:

```
`xs3scann3r` -h
```

help message:

```
          _____                           _____      
__  _____|___ / ___  ___ __ _ _ __  _ __ |___ / _ __ 
\ \/ / __| |_ \/ __|/ __/ _` | '_ \| '_ \  |_ \| '__|
 >  <\__ \___) \__ \ (_| (_| | | | | | | |___) | |   
/_/\_\___/____/|___/\___\__,_|_| |_|_| |_|____/|_| v0.0.0

A CLI utility to scan S3 buckets permissions.

USAGE:
  xs3scann3r [OPTIONS]

INPUT:
  -i, --input         input file (use `-` to get from stdin)

CONFIGURATIONS:
   -c, --concurrency  number of concurrent threads (default: 10)
   -d, --dump         location to dump objects

OUTPUT:
  -m, --monochrome    disable output content coloring
  -v, --verbosity     debug, info, warning, error, fatal or silent (default: info)
```

xs3scann3r takes buckets in the format:

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

[Issues](https://github.com/hueristiq/xs3scann3r/issues) and [Pull Requests](https://github.com/hueristiq/xs3scann3r/pulls) are welcome! Check out the [contribution guidelines.](./CONTRIBUTING.md)

## Licensing

This utility is distributed under the [MIT license](./LICENSE)