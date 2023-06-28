# GoPistolet

[![Build Status](https://travis-ci.org/gopistolet/gopistolet.svg?branch=master)](https://travis-ci.org/gopistolet/gopistolet)

GoPistolet will be a production-ready, and easy to setup mailserver (MTA/MSA/IMAP).


## Usage

    go run cmd/gopistolet/*.go

It will seed the database with a user with `username` and `password` as username and password.


## Development

Navigate to the parent folder and execute the following commands so you don't have to do rewrite in the go mod file.

    go work init
    go work use smtp
    go work use imap-backend

Also set `GOPRIVATE` so it fetches the gopistolet repos over SSH:

    go env -w "GOPRIVATE=github.com/gopistolet/*"


## Acknowledgements

* [GoConvey](https://github.com/smartystreets/goconvey)
* [go-maildir](https://github.com/sloonz/go-maildir)


## Authors

Mathias Beke - [denbeke.be](http://denbeke.be)  
Timo Truyts
