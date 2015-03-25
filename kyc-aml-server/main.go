/*
	This is for checking names and addresses against a blacklist.
*/

package main

import (
	kyc_aml_server "./KycAmlServer"
)

func main() {
	
	// Start a new server.
	server, err := kyc_aml_server.NewKycAmlServer("KycAmlServer/config.json")
	if err != nil {
		return
	}
	
	// Listen for connections.
	err = server.Listen()
	if err != nil {
		return
	}
}
