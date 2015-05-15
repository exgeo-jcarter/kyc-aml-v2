package main

import (
	"os"
	"net"
	"encoding/gob"
	"log"
)

func main() {
	
	if len(os.Args) <= 2 {
		log.Printf("Error: Not enough args")
		return
	}
	
	con, err := net.Dial("tcp", "0.0.0.0:3336")
	if err != nil {
		log.Printf("Error: net.Dial: %v", err)
		return
	}
	
	enc := gob.NewEncoder(con)
	dec := gob.NewDecoder(con)
	
	err = enc.Encode(ClientServerQueryReqS{
		QueryName: os.Args[1],
		QueryAddress: os.Args[2],
	})
	if err != nil {
		log.Printf("Error: Encode: %v", err)
		return
	}
	
	var res *ClientServerQueryResS
	err = dec.Decode(&res)
	if err != nil {
		log.Printf("Error: Decode: %v", err)
		return
	}
	
	log.Printf("res: %+v", res)
}
