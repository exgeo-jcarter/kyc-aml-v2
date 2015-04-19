package main

import (
	kyc_aml_client_server "./KycAmlClientServer"
)

func main() {
	
	client, err := kyc_aml_client_server.NewKycAmlClientServer("KycAmlClientServer/config.json")
	if err != nil {
		return
	}
	
	err = client.Listen()
	if err != nil {
		return
	}
	
	
}
