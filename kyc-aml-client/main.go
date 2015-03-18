package main

import (
	"os"
	"fmt"
)

func main() {
	
	if len(os.Args) < 2 {
		fmt.Printf("\n"+`usage: kyc-aml-client "a query goes here"`+"\n\n")
		return
	}
	
	client, err := NewKycAmlClient("config.json")
	if err != nil {
		return
	}
	
	err = client.Query(os.Args[1])
	if err != nil {
		return
	}
}