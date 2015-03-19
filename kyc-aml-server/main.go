/*
	This is for checking names and addresses against a blacklist.
*/

package main

func main() {
	
	// Start a new server.
	server, err := NewKycAmlServer("config.json")
	if err != nil {
		return
	}
	
	// Listen for connections.
	err = server.Listen()
	if err != nil {
		return
	}
}
