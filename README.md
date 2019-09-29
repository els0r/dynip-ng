# Dynamic IP Updater

> Listens on your external interface and updates your files and Cloudflare DNS records

[![Github Release](https://img.shields.io/github/release/els0r/dynip-ng.svg)](https://github.com/els0r/dynip-ng/releases)
[![GoDoc](https://godoc.org/github.com/els0r/dynip-ng?status.svg)](https://godoc.org/github.com/els0r/dynip-ng/)
[![Go Report Card](https://goreportcard.com/badge/github.com/els0r/dynip-ng)](https://goreportcard.com/report/github.com/els0r/dynip-ng)
[![Build Status](https://cloud.drone.io/api/badges/els0r/dynip-ng/status.svg)](https://cloud.drone.io/els0r/dynip-ng)

## Overview

This tool was born out of a simple problem: not having access to a static IP address. Since configuration for some programs relies on the IP being explicitly specified, it's annoying to adapt configuration whenever the external IP changes.

The same goes for DNS records that are hosted on Cloudflare. If the external IP is listed in one of your zone's records, why log in every time and change it manually?

Cloudflare has an API and `go` provides a templating language. This small tool leverages the two.

## How to install

The easiest way is to run

```bash
go get github.com/els0r/dynip-ng
cd $GOPATH/src/github.com/els0r/dynip-ng
go generate
```

This will fetch the source code, compile it (with the version baked into the binary) and provide a deployable archive with `systemd` files and an example configuration.

## How to configure

Check [the example configuration](./addon/dynip-ng.yml.example) for how to configure the tool.

It can be verified with

```bash
dynip-ng config -c /path/to/config/file
```

## How to run

To start the listener from the command line, run

```bash
dynip-ng run -c /path/to/config/file
```

If you are debian-based and want to run it as a daemon (recommended), copy the `dynip.service` file to your `systemd` files and run

```bash
systemctl enable dynip.service
systemctl start dynip.service
```

## How to deploy

If you want to deploy the TAR archive with the pre-defined directory structure (see [install.sh](./install.sh)), run

```bash
tar xf dynip.tar.bz2 -C / --strip-components=2
```

## Bug Reports

Please use the [issue tracker](https://github.com/els0r/dynip-ng/issues) for bugs and feature requests.

## License

See the [LICENSE](./LICENSE) file for usage conditions.
