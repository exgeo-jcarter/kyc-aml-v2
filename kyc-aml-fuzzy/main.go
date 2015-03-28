/*
	This is for checking names and addresses against a blacklist.
*/

package main

import (
	kyc_aml_fuzzy "./KycAmlFuzzy"
)

func main() {
	
	// Start a new server.
	server, err := kyc_aml_fuzzy.NewKycAmlFuzzy("KycAmlFuzzy/config.json")
	if err != nil {
		return
	}
	
	// Listen for connections.
	err = server.Listen()
	if err != nil {
		return
	}
}
