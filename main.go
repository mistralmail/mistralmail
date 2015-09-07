package main

import (
	"fmt"
	"log"

	"github.com/gopistolet/gopistolet/mta"
)

func main() {
	fmt.Println("GoPistolet at your service!")

	c := mta.Config{
		Hostname: "localhost",
		Port:     2525,
	}
	mta := mta.NewDefault(c)
	err := mta.ListenAndServe()
	if err != nil {
		log.Println(err)
	}
}
