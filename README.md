GoPistolet
==========

[![Build Status](https://travis-ci.org/gopistolet/gopistolet.svg?branch=master)](https://travis-ci.org/gopistolet/gopistolet)

GoPistolet will be a production-ready, and easy to setup mailserver (MTA/MSA/IMAP).

Status
------

Right now we have implemented the SMTP protocol [RFC 5321](go get github.com/gopistolet/gopistolet) and the MTA part.
The server listens on a socket and saves all incoming messages in a maildir.

Screenshots of the maildir, openened with [Mutt](http://www.mutt.org):

![maildir with mutt](http://denbeke.be/foto/GoPistolet_maildir.png)

![maildir with mutt](http://denbeke.be/foto/GoPistolet_maildir2.png)


Installing
----------

Install GoPistolet:

    $ go get github.com/gopistolet/gopistolet

You need the following packages (look in `.travis.yml` for an up-to-date list):

    $ go get github.com/smartystreets/goconvey/convey
    $ go get github.com/gopistolet/gospf
    $ go get github.com/sloonz/go-maildir
   
    
    
Configuration
-------------

Copy `config.sample.json` to `config.json` and edit the file if you want to change the defaults.


Acknowledgements
-----------------

* [GoConvey](github.com/smartystreets/goconvey/convey)
* [go-maildir](github.com/sloonz/go-maildir)

Authors
-------

Mathias Beke - [denbeke.be](http://denbeke.be)  
Timo Truyts
