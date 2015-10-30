// This package is a wrapper for log functions
// in order to ensure substitutability with other libraries
package log

import "github.com/Sirupsen/logrus"


func Printf(format string, v ...interface{}) {
	logrus.Printf(format, v...)
}

func Println(v ...interface{}) {
    logrus.Println(v...)
}

func Fatalf(format string, v ...interface{}) {
	logrus.Fatalf(format, v...)
}