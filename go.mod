module github.com/gopistolet/gopistolet

go 1.15

require (
	github.com/gopistolet/gospf v0.0.0-20160422193406-a58dd1fcbf50
	github.com/gopistolet/imap-backend v0.0.0-master
	github.com/gopistolet/smtp v0.0.0-20220206164535-7d177c8d6ca1
	github.com/sirupsen/logrus v1.8.1
	github.com/smartystreets/goconvey v1.6.4

)

replace github.com/gopistolet/imap-backend v0.0.0-master => ../imap-backend
replace github.com/gopistolet/smtp v0.0.0-master => ../smtp
