# MistralMail

# WIP: don't use this!

MistralMail will be a production-ready, and easy to setup mailserver (MTA/MSA/IMAP).


## Usage

    go run cmd/mistralmail/*.go

It will seed the database with a user with `username@example.com` and `password` as username and password.


## Development

Navigate to the parent folder and execute the following commands so you don't have to do rewrite in the go mod file.

    go work init
    go work use smtp
    go work use imap-backend

Also set `GOPRIVATE` so it fetches the mistralmail repos over SSH:

    go env -w "GOPRIVATE=github.com/mistralmail/*"


## Acknowledgements

* [GoConvey](https://github.com/smartystreets/goconvey)
* [go-maildir](https://github.com/sloonz/go-maildir)


## Authors

Mathias Beke - [denbeke.be](http://denbeke.be)  
Timo Truyts
