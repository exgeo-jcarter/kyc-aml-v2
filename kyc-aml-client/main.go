package main

import (
	kyc_aml_client "./KycAmlClient"
	"os"
	"fmt"
)

func main() {
	
	if len(os.Args) < 2 {
		fmt.Printf("\n"+`usage: kyc-aml-client "a query goes here"`+"\n\n")
		return
	}
	
	client, err := kyc_aml_client.NewKycAmlClient("KycAmlClient/config.json")
	if err != nil {
		return
	}
	
	query_res, err := client.Query(os.Args[1])
	if err != nil {
		return
	}
	fmt.Printf("%s\n", query_res)
}