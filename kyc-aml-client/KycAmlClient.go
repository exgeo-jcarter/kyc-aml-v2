/*
	This is for checking names and addresses against a blacklist.
*/

package main

import (
	"net"
	"encoding/json"
	"log"
	"bufio"
	"io/ioutil"
	"fmt"
	"strings"
)

type kycAmlClientS struct {
	Conf *KycAmlClientConfS
}

type KycAmlClientConfS struct {
	Host 		string	`json:"host,omitempty"`
	Port 		string	`json:"port,omitempty"`
	Protocol 	string	`json:"protocol,omitempty"`
}

// Initialize the client.
func NewKycAmlClient(conf_filename string) (new_kycamlclient *kycAmlClientS, err error) {
	
	new_kycamlclient = &kycAmlClientS{}
	
	// Load server settings.
	err = new_kycamlclient.LoadConf(conf_filename)
	if err != nil {
		return
	}
	
	return
}

// Load server settings.
func (this *kycAmlClientS) LoadConf(filename string) (err error) {
	
	conf_bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	
	err = json.Unmarshal(conf_bytes, &this.Conf)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	
	return
}

// Query the server to check for string matches against the blacklist.
func (this *kycAmlClientS) Query(q string) (err error) {
	
	// Make the query struct.
	msg_struct := QueryReqS{
		Action: "query",
		Value: strings.ToLower(q),
	}
	
	// Marshal the query struct.
	msg, err := json.Marshal(msg_struct)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	
	// Add newline to the end of the query bytes, so server knows where to stop reading.
	msg = []byte(string(msg)+"\n")
	
	// Connect to the server.
	con, err := net.Dial(this.Conf.Protocol, this.Conf.Host+":"+this.Conf.Port)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	
	// Make a channel that's used to wait for the server's response.
	resCh := make(chan int)
	conbuf := bufio.NewReader(con)
	
	// Start listening for a response from server.
	go (func(resCh chan int, conbuf *bufio.Reader) {
		
		res := []byte{}
		
		for {
			
			// Read server's response until newline.
			res, err = conbuf.ReadBytes('\n')
			if err != nil {
				log.Printf("Error: %v", err)
		    	resCh <- 0
		    	break
			}
			
			// If we gone a non-empty response.
			if len(res) > 0 {
				
				// Parse the server's response into a struct.
				var query_res QueryResS
				err = json.Unmarshal(res[:len(res)-1], &query_res)
				if err != nil {
					log.Printf("Error: %v", err)
					resCh <- 0
					return
				}
				
				// Remove empty strings from the result section of server's response.
				for idx, val := range query_res.Result {
					if (val == "") {
						if idx < (len(query_res.Result)-1) {
							query_res.Result = append(query_res.Result[:idx], query_res.Result[idx+1:]...)
						}
					}
				}
				if len(query_res.Result) > 2 {
					query_res.Result = query_res.Result[2:]
				}
				
				// Remove empty strings from the metaphone_result section of server's response.
				for idx, val := range query_res.MetaphoneResult {
					if (val == "") {
						if idx < (len(query_res.MetaphoneResult)-1) {
							query_res.MetaphoneResult = append(query_res.MetaphoneResult[:idx], query_res.MetaphoneResult[idx+1:]...)
						}
					}
				}
				if len(query_res.MetaphoneResult) > 2 {
					query_res.MetaphoneResult = query_res.MetaphoneResult[2:]
				}
				
				// Marshal the query response.
				query_res_bytes, err := json.Marshal(query_res)
				if err != nil {
					log.Printf("Error: %v", err)
					resCh <- 0
					return
				}
				
				// Output the query response.
				fmt.Printf("%s\n", query_res_bytes)
				
				// Send an int to the channel, meaning we can exit the program now.
				resCh <- 1
				break
			}
		}
		
		// Close connection to the server.
		con.Close()
		
	})(resCh, conbuf)
	
	// Send query to the server.
	_, err = con.Write(msg)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	
	// Wait for server's response.
	<- resCh
	
	return
}
