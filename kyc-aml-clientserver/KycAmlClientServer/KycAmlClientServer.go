/*
	This is for checking names and addresses against a blacklist.
*/

package KycAmlClientServer

import (
	"net"
	"encoding/json"
	"log"
	"bufio"
	"io/ioutil"
	"strings"
	"fmt"
	"os"
	"time"
	"encoding/gob"
)

type KycAmlClientServerS struct {
	Conf 	*KycAmlClientServerConfS
}

type KycAmlClientServerConfS struct {
	ClientHost 				string	`json:"client_host,omitempty"`
	ClientPort 				string	`json:"client_port,omitempty"`
	ClientProtocol			string	`json:"client_protocol,omitempty"`
	DataHost 				string	`json:"data_host,omitempty"`
	DataPort 				string	`json:"data_port,omitempty"`
	DataProtocol			string	`json:"data_protocol,omitempty"`
	FuzzyHost 				string	`json:"fuzzy_host,omitempty"`
	FuzzyPort 				string	`json:"fuzzy_port,omitempty"`
	FuzzyProtocol			string	`json:"fuzzy_protocol,omitempty"`
	MetaphoneHost 			string	`json:"metaphone_host,omitempty"`
	MetaphonePort 			string	`json:"metaphone_port,omitempty"`
	MetaphoneProtocol		string	`json:"metaphone_protocol,omitempty"`
	DoubleMetaphoneHost 	string	`json:"doublemetaphone_host,omitempty"`
	DoubleMetaphonePort 	string	`json:"doublemetaphone_port,omitempty"`
	DoubleMetaphoneProtocol	string	`json:"doublemetaphone_protocol,omitempty"`
}

// Initialize the client.
func NewKycAmlClientServer(conf_filename string) (new_kycamlclientserver *KycAmlClientServerS, err error) {
	
	new_kycamlclientserver = &KycAmlClientServerS{}
	
	// Load server settings.
	err = new_kycamlclientserver.LoadConf(conf_filename)
	if err != nil {
		return
	}
	
	return
}

// Load server settings.
func (this *KycAmlClientServerS) LoadConf(filename string) (err error) {
	
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
	
	//log.Printf("Client config file loaded.")
	
	return
}

func (this *KycAmlClientServerS) Listen() (err error) {
	
	// Listen.
	l, err := net.Listen(this.Conf.ClientProtocol, this.Conf.ClientHost+":"+this.Conf.ClientPort)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	defer l.Close()
	
	log.Printf("Client server listening.")
	
	for {
		// Accept new connection.
		con, err := l.Accept()
		if err != nil {
			log.Printf("Error: %v", err)
			continue
		}
		
		// Do something with the new connection.
		go this.handleRequest(con)
	}
	
	return
}

// TODO: here
func (this *KycAmlClientServerS) handleRequest(con net.Conn) {
	
	// Gob request
	decoder := gob.NewDecoder(con)
	var socketMsg ClientServerQueryReqS
	err := decoder.Decode(&socketMsg)
	if err != nil {
		log.Printf("Error: %v", err)
		con.Close()
		return
	}
	
	/*	// JSON request
	conbuf := bufio.NewReader(con)
	
	// Read buffer until newline.
	res, err := conbuf.ReadBytes('\n')
	if err != nil {
		log.Printf("Error: %v", err)
		con.Close()
		return
	}
	
	var socketMsg ClientServerQueryReqS
	err = json.Unmarshal(res[:len(res)-1], &socketMsg)
	if err != nil {
		log.Printf("Error: %v", err)
		con.Close()
		return
	}
	*/
	
	if socketMsg.QueryName != "" || socketMsg.QueryAddress != "" {

		_, err = this.QueryDataServer("load_sdn_list", "")
		if err != nil {
			return
		}
		
		sdn_list, err := this.QueryDataServer("get_sdn_list", "")
		if err != nil {
			return
		}
		
		num_query_servers := 3
		wait_for_training_ch := make(chan int)
		
		go (func() {
			fuzzy_train_sdn_res, err := this.QueryFuzzyServer("train_sdn", sdn_list)
			if err != nil {
				return
			}
			_ = fuzzy_train_sdn_res
			wait_for_training_ch <- 1
		})()
		
		go (func() {
			metaphone_train_sdn_res, err := this.QueryMetaphoneServer("train_sdn", sdn_list)
			if err != nil {
				return
			}
			_ = metaphone_train_sdn_res
			wait_for_training_ch <- 1
		})()
		
		go (func() {
			doublemetaphone_train_sdn_res, err := this.QueryDoubleMetaphoneServer("train_sdn", sdn_list)
			if err != nil {
				return
			}
			_ = doublemetaphone_train_sdn_res
			wait_for_training_ch <- 1
		})()
		
		for i := 0; i < num_query_servers; i++ {
			<- wait_for_training_ch
		}
		
		num_queries := 0
		wait_for_queries_ch := make(chan int)
		
		aq := ""
		
		fuzzy_name_res := "{}"
		fuzzy_address_res := "{}"
		metaphone_name_res := "{}"
		metaphone_address_res := "{}"
		doublemetaphone_name_res := "{}"
		doublemetaphone_address_res := "{}"
		
		logfile, err := os.OpenFile("kyc-aml.log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			log.Printf("Error: %v", err)
			return
		}
		defer logfile.Close()
		
		if socketMsg.QueryName != "" {
			
			num_queries += num_query_servers
			
			go (func() {
				fuzzy_name_res, err = this.QueryFuzzyServer("query_name", socketMsg.QueryName)
				if err != nil {
					return
				}
				if fuzzy_name_res != "{}" {
					fmt.Printf("Fuzzy name results: %s\n", fuzzy_name_res)
					
					t := time.Now().UTC().Format("2006-01-02 15:04:05")
					_, err = logfile.WriteString(t + " - Fuzzy name results: " + fuzzy_name_res + "\n")
					if err != nil {
						log.Printf("Error: %v", err)
						return
					}
				}
				wait_for_queries_ch <- 1
			})()
			
			go (func() {
				metaphone_name_res, err = this.QueryMetaphoneServer("query_name", socketMsg.QueryName)
				if err != nil {
					return
				}
				if metaphone_name_res != "{}" {
					fmt.Printf("Metaphone name results: %s\n", metaphone_name_res)
					
					t := time.Now().UTC().Format("2006-01-02 15:04:05")
					_, err = logfile.WriteString(t + " - Metaphone name results: " + metaphone_name_res + "\n")
					if err != nil {
						log.Printf("Error: %v", err)
						return
					}
				}
				wait_for_queries_ch <- 1
			})()
			
			go (func() {
				doublemetaphone_name_res, err = this.QueryDoubleMetaphoneServer("query_name", socketMsg.QueryName)
				if err != nil {
					return
				}
				if doublemetaphone_name_res != "{}" {
					fmt.Printf("DoubleMetaphone name results: %s\n", doublemetaphone_name_res)
					
					t := time.Now().UTC().Format("2006-01-02 15:04:05")
					_, err = logfile.WriteString(t + " - DoubleMetaphone name results: " + doublemetaphone_name_res + "\n")
					if err != nil {
						log.Printf("Error: %v", err)
						return
					}
				}
				wait_for_queries_ch <- 1
			})()
		}
		
		if socketMsg.QueryAddress != "" {
			
			num_queries += num_query_servers
			aq = socketMsg.QueryAddress
			
			go (func() {
				fuzzy_address_res, err = this.QueryFuzzyServer("query_address", socketMsg.QueryAddress)
				if err != nil {
					return
				}
				if fuzzy_address_res != "{}" {
					fmt.Printf("Fuzzy address results: %s\n", fuzzy_address_res)
					
					t := time.Now().UTC().Format("2006-01-02 15:04:05")
					_, err = logfile.WriteString(t + " - Fuzzy address results: " + fuzzy_address_res + "\n")
					if err != nil {
						log.Printf("Error: %v", err)
						return
					}
				}
				wait_for_queries_ch <- 1
			})()
			
			go (func() {
				metaphone_address_res, err = this.QueryMetaphoneServer("query_address", socketMsg.QueryAddress)
				if err != nil {
					return
				}
				if metaphone_address_res != "{}" {
					fmt.Printf("Metaphone address results: %s\n", metaphone_address_res)
					
					t := time.Now().UTC().Format("2006-01-02 15:04:05")
					_, err = logfile.WriteString(t + " - Metaphone address results: " + metaphone_address_res + "\n")
					if err != nil {
						log.Printf("Error: %v", err)
						return
					}
				}
				wait_for_queries_ch <- 1
			})()
			
			go (func() {
				doublemetaphone_address_res, err = this.QueryDoubleMetaphoneServer("query_address", socketMsg.QueryAddress)
				if err != nil {
					return
				}
				if doublemetaphone_address_res != "{}" {
					fmt.Printf("DoubleMetaphone address results: %s\n", doublemetaphone_address_res)
					
					t := time.Now().UTC().Format("2006-01-02 15:04:05")
					_, err = logfile.WriteString(t + " - DoubleMetaphone address results: " + doublemetaphone_address_res + "\n")
					if err != nil {
						log.Printf("Error: %v", err)
						return
					}
				}
				wait_for_queries_ch <- 1
			})()
		}
		
		for i := 0; i < num_queries; i++ {
			<- wait_for_queries_ch
		}
		
		risk_score, err := this.CalculateRiskScore(socketMsg.QueryName, aq, fuzzy_name_res, fuzzy_address_res, metaphone_name_res, metaphone_address_res, doublemetaphone_name_res, doublemetaphone_address_res)
		if err != nil {
			return
		}
		
		fmt.Printf("Risk score: %v\n", risk_score)
		t := time.Now().UTC().Format("2006-01-02 15:04:05")
		_, err = logfile.WriteString(t + " - Risk score: " + fmt.Sprintf("%v", risk_score) + "\n")
		if err != nil {
			log.Printf("Error: %v", err)
			return
		}
		
		sdn_entry_res, err := this.LookupSdnEntry(fuzzy_name_res, fuzzy_address_res, metaphone_name_res, metaphone_address_res, doublemetaphone_name_res, doublemetaphone_address_res)
		if err != nil {
			return
		}
		
		fmt.Printf("SDN entry: %s\n", sdn_entry_res)
		t = time.Now().UTC().Format("2006-01-02 15:04:05")
		_, err = logfile.WriteString(t + " - SDN entry: " + sdn_entry_res + "\n")
		if err != nil {
			log.Printf("Error: %v", err)
			return
		}
		
		t = time.Now().UTC().Format("2006-01-02 15:04:05")
		_, err = logfile.WriteString(t + " ----------\n")
		if err != nil {
			log.Printf("Error: %v", err)
			return
		}
		
		var fuzzy_name_res_struct FuzzyQueryResS
		err = json.Unmarshal([]byte(fuzzy_name_res), &fuzzy_name_res_struct)
		if err != nil {
			log.Printf("Error: %v", err)
			return
		}
		
		var fuzzy_address_res_struct FuzzyQueryResS
		err = json.Unmarshal([]byte(fuzzy_address_res), &fuzzy_address_res_struct)
		if err != nil {
			log.Printf("Error: %v", err)
			return
		}
		
		var metaphone_name_res_struct MetaphoneQueryResS
		err = json.Unmarshal([]byte(metaphone_name_res), &metaphone_name_res_struct)
		if err != nil {
			log.Printf("Error: %v", err)
			return
		}
		
		var metaphone_address_res_struct MetaphoneQueryResS
		err = json.Unmarshal([]byte(metaphone_address_res), &metaphone_address_res_struct)
		if err != nil {
			log.Printf("Error: %v", err)
			return
		}
		
		var doublemetaphone_name_res_struct DoubleMetaphoneQueryResS
		err = json.Unmarshal([]byte(doublemetaphone_name_res), &doublemetaphone_name_res_struct)
		if err != nil {
			log.Printf("Error: %v", err)
			return
		}
		
		var doublemetaphone_address_res_struct DoubleMetaphoneQueryResS
		err = json.Unmarshal([]byte(doublemetaphone_address_res), &doublemetaphone_address_res_struct)
		if err != nil {
			log.Printf("Error: %v", err)
			return
		}
		
		var sdn_entry_res_struct SdnEntryS
		
		if len(sdn_entry_res) > 0 {
			err = json.Unmarshal([]byte(sdn_entry_res), &sdn_entry_res_struct)
			if err != nil {
				log.Printf("Error: %v", err)
				return
			}
		}
		
		matches := []string{}
		
		if len(fuzzy_name_res_struct.NameResult) > 0 {
			matches = append(matches, fuzzy_name_res_struct.NameResult...)
		}
		
		if len(fuzzy_name_res_struct.RevNameResult) > 0 {
			matches = append(matches, fuzzy_name_res_struct.RevNameResult...)
		}
		
		if len(fuzzy_name_res_struct.AkaResult) > 0 {
			matches = append(matches, fuzzy_name_res_struct.AkaResult...)
		}
		
		if len(fuzzy_name_res_struct.RevAkaResult) > 0 {
			matches = append(matches, fuzzy_name_res_struct.RevAkaResult...)
		}
		
		if len(fuzzy_address_res_struct.AddressResult) > 0 {
			matches = append(matches, fuzzy_address_res_struct.AddressResult...)
		}
		
		if len(fuzzy_address_res_struct.PostalCodeResult) > 0 {
			matches = append(matches, fuzzy_address_res_struct.PostalCodeResult...)
		}
		
		
		if len(metaphone_name_res_struct.NameResult) > 0 {
			matches = append(matches, metaphone_name_res_struct.NameResult...)
		}
		
		if len(metaphone_name_res_struct.RevNameResult) > 0 {
			matches = append(matches, metaphone_name_res_struct.RevNameResult...)
		}
		
		if len(metaphone_name_res_struct.AkaResult) > 0 {
			matches = append(matches, metaphone_name_res_struct.AkaResult...)
		}
		
		if len(metaphone_name_res_struct.RevAkaResult) > 0 {
			matches = append(matches, metaphone_name_res_struct.RevAkaResult...)
		}
		
		if len(metaphone_address_res_struct.AddressResult) > 0 {
			matches = append(matches, metaphone_address_res_struct.AddressResult...)
		}
		
		if len(metaphone_address_res_struct.PostalCodeResult) > 0 {
			matches = append(matches, metaphone_address_res_struct.PostalCodeResult...)
		}
		
		
		if len(doublemetaphone_name_res_struct.NameResult1) > 0 {
			matches = append(matches, doublemetaphone_name_res_struct.NameResult1...)
		}
		
		if len(doublemetaphone_name_res_struct.NameResult2) > 0 {
			matches = append(matches, doublemetaphone_name_res_struct.NameResult2...)
		}
		
		if len(doublemetaphone_name_res_struct.RevNameResult1) > 0 {
			matches = append(matches, doublemetaphone_name_res_struct.RevNameResult1...)
		}
		
		if len(doublemetaphone_name_res_struct.RevNameResult2) > 0 {
			matches = append(matches, doublemetaphone_name_res_struct.RevNameResult2...)
		}
		
		if len(doublemetaphone_name_res_struct.AkaResult1) > 0 {
			matches = append(matches, doublemetaphone_name_res_struct.AkaResult1...)
		}
		
		if len(doublemetaphone_name_res_struct.AkaResult2) > 0 {
			matches = append(matches, doublemetaphone_name_res_struct.AkaResult2...)
		}
		
		if len(doublemetaphone_name_res_struct.RevAkaResult1) > 0 {
			matches = append(matches, doublemetaphone_name_res_struct.RevAkaResult1...)
		}
		
		if len(doublemetaphone_name_res_struct.RevAkaResult2) > 0 {
			matches = append(matches, doublemetaphone_name_res_struct.RevAkaResult2...)
		}
		
		if len(doublemetaphone_address_res_struct.AddressResult1) > 0 {
			matches = append(matches, doublemetaphone_address_res_struct.AddressResult1...)
		}
		
		if len(doublemetaphone_address_res_struct.AddressResult2) > 0 {
			matches = append(matches, doublemetaphone_address_res_struct.AddressResult2...)
		}
		
		if len(doublemetaphone_address_res_struct.PostalCodeResult1) > 0 {
			matches = append(matches, doublemetaphone_address_res_struct.PostalCodeResult1...)
		}
		
		if len(doublemetaphone_address_res_struct.PostalCodeResult2) > 0 {
			matches = append(matches, doublemetaphone_address_res_struct.PostalCodeResult2...)
		}
		
		
		msg := ClientServerQueryResS{
			FuzzyName: 				fuzzy_name_res_struct,
			FuzzyAddress: 			fuzzy_address_res_struct,
			MetaphoneName: 			metaphone_name_res_struct,
			MetaphoneAddress: 		metaphone_address_res_struct,
			DoubleMetaphoneName: 	doublemetaphone_name_res_struct,
			DoubleMetaphoneAddress: doublemetaphone_address_res_struct,
			SdnEntry:				sdn_entry_res_struct,
			RiskScore:				risk_score,
			Matches:				matches,
		}
	
		
		// Gob response
		encoder := gob.NewEncoder(con)
		msg_p := &msg
		encoder.Encode(msg_p)
	
		/*	// JSON response
		msg_bytes, err := json.Marshal(msg)
		if err != nil {
			log.Printf("Error: %v", err)
			return
		}
		
		msg_bytes = append(msg_bytes, []byte("\n")...)
		
		_, err = con.Write(msg_bytes)
		if err != nil {
			log.Printf("Error: %v", err)
			return
		}
		*/
	}
	
	// Close the client's connection.
	con.Close()
}

// Query the fuzzy server to check for string matches against the blacklist.
func (this *KycAmlClientServerS) QueryServer(protocol, host, port, action, value string) (res string, err error) {
	
	// Make the query struct.
	msg_struct := QueryReqS{
		Action: action,
		Value: strings.ToLower(value),
	}
	
	// Marshal the query struct.
	msg, err := json.Marshal(msg_struct)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	
	// Add newline to the end of the query bytes, so server knows where to stop reading.
	msg = append(msg, []byte("\n")...)
	
	// Connect to the server.
	con, err := net.Dial(protocol, host+":"+port)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	
	// Make a channel that's used to wait for the server's response.
	resCh := make(chan int)
	conbuf := bufio.NewReader(con)
	
	// Start listening for a response from server.
	go (func(resCh chan int, conbuf *bufio.Reader) {
		
		for {
			
			// Read server's response until newline.
			res_bytes, err := conbuf.ReadBytes('\n')
			if err != nil {
				log.Printf("Error: %v", err)
		    	resCh <- 0
		    	break
			}
			
			res = string(res_bytes)
			
			// If we gone a non-empty response.
			if len(res) > 0 {
				
				res = res[:len(res)-1]
				
				// Output the query response.
				//fmt.Printf("%s\n", res[:len(res)-1])
				
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

func (this *KycAmlClientServerS) QueryDataServer(action, value string) (res string, err error) {
	
	res, err = this.QueryServer(this.Conf.DataProtocol, this.Conf.DataHost, this.Conf.DataPort, action, value)
	return
}

func (this *KycAmlClientServerS) QueryFuzzyServer(action, value string) (res string, err error) {
	
	res, err = this.QueryServer(this.Conf.FuzzyProtocol, this.Conf.FuzzyHost, this.Conf.FuzzyPort, action, value)
	return
}

func (this *KycAmlClientServerS) QueryMetaphoneServer(action, value string) (res string, err error) {
	
	res, err = this.QueryServer(this.Conf.MetaphoneProtocol, this.Conf.MetaphoneHost, this.Conf.MetaphonePort, action, value)
	return
}

func (this *KycAmlClientServerS) QueryDoubleMetaphoneServer(action, value string) (res string, err error) {
	
	res, err = this.QueryServer(this.Conf.DoubleMetaphoneProtocol, this.Conf.DoubleMetaphoneHost, this.Conf.DoubleMetaphonePort, action, value)
	return
}

func (this *KycAmlClientServerS) CalculateRiskScore(q, aq, f_name_res, f_add_res, m_name_res, m_add_res, dm_name_res, dm_add_res string) (score float64, err error) {
	
	num_queries := 0.0
	num_results := 0.0
	
	//if q != "" {
		num_queries += 6
	//}
	
	//if aq != "" {
	//	num_queries += 4
	//}
	
	var f_name_res_json FuzzyQueryResS
	err = json.Unmarshal([]byte(f_name_res), &f_name_res_json)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	if len(f_name_res_json.NameResult) > 0 {
		num_results++
	} else if len(f_name_res_json.RevNameResult) > 0 {
		num_results++
	} else if len(f_name_res_json.AkaResult) > 0 {
		num_results++
	} else if len(f_name_res_json.RevAkaResult) > 0 {
		num_results++
	}
	
	var f_add_res_json FuzzyQueryResS
	err = json.Unmarshal([]byte(f_add_res), &f_add_res_json)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	if len(f_add_res_json.AddressResult) > 0 {
		num_results++
	} else if len(f_add_res_json.PostalCodeResult) > 0 {
		num_results++
	}
	
	var m_name_res_json MetaphoneQueryResS
	err = json.Unmarshal([]byte(m_name_res), &m_name_res_json)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	if len(m_name_res_json.NameResult) > 0 {
		num_results++
	} else if len(m_name_res_json.RevNameResult) > 0 {
		num_results++
	} else if len(m_name_res_json.AkaResult) > 0 {
		num_results++
	} else if len(m_name_res_json.RevAkaResult) > 0 {
		num_results++
	}
	
	var m_add_res_json MetaphoneQueryResS
	err = json.Unmarshal([]byte(m_add_res), &m_add_res_json)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	if len(m_add_res_json.AddressResult) > 0 {
		num_results++
	} else if len(m_add_res_json.PostalCodeResult) > 0 {
		num_results++
	}
	
	var dm_name_res_json DoubleMetaphoneQueryResS
	err = json.Unmarshal([]byte(dm_name_res), &dm_name_res_json)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	if len(dm_name_res_json.NameResult1) > 0 {
		num_results++
	} else if len(dm_name_res_json.NameResult2) > 0 {
		num_results++
	} else if len(dm_name_res_json.RevNameResult1) > 0 {
		num_results++
	} else if len(dm_name_res_json.RevNameResult2) > 0 {
		num_results++
	} else if len(dm_name_res_json.AkaResult1) > 0 {
		num_results++
	} else if len(dm_name_res_json.AkaResult2) > 0 {
		num_results++
	} else if len(dm_name_res_json.RevAkaResult1) > 0 {
		num_results++
	} else if len(dm_name_res_json.RevAkaResult2) > 0 {
		num_results++
	}
	
	var dm_add_res_json DoubleMetaphoneQueryResS
	err = json.Unmarshal([]byte(dm_add_res), &dm_add_res_json)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	if len(dm_add_res_json.AddressResult1) > 0 {
		num_results++
	} else if len(dm_add_res_json.AddressResult2) > 0 {
		num_results++
	} else if len(dm_add_res_json.PostalCodeResult1) > 0 {
		num_results++
	} else if len(dm_add_res_json.PostalCodeResult2) > 0 {
		num_results++
	}
	
	score = num_results / num_queries * 100.0
	return
}

// Looks up an entry in the SDN list.
func (this *KycAmlClientServerS) LookupSdnEntry(f_name_res, f_add_res, m_name_res, m_add_res, dm_name_res, dm_add_res string) (res string, err error) {
	
	var f_name_res_json FuzzyQueryResS
	err = json.Unmarshal([]byte(f_name_res), &f_name_res_json)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	
	var f_add_res_json FuzzyQueryResS
	err = json.Unmarshal([]byte(f_add_res), &f_add_res_json)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	
	var m_name_res_json MetaphoneQueryResS
	err = json.Unmarshal([]byte(m_name_res), &m_name_res_json)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	
	var m_add_res_json MetaphoneQueryResS
	err = json.Unmarshal([]byte(m_add_res), &m_add_res_json)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	
	var dm_name_res_json DoubleMetaphoneQueryResS
	err = json.Unmarshal([]byte(dm_name_res), &dm_name_res_json)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	
	var dm_add_res_json DoubleMetaphoneQueryResS
	err = json.Unmarshal([]byte(dm_add_res), &dm_add_res_json)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	
	var sdn_entry_res string
	
	if len(f_name_res_json.NameResult) > 0 {
	
		sdn_entry_res, err = this.QueryMetaphoneServer("lookup_sdn_entry", f_name_res_json.NameResult[0])
		if err != nil {
			return
		}
		res = sdn_entry_res
		return
	}
	
	if len(f_name_res_json.RevNameResult) > 0 {
		
		sdn_entry_res, err = this.QueryMetaphoneServer("lookup_sdn_entry", f_name_res_json.RevNameResult[0])
		if err != nil {
			return
		}
		res = sdn_entry_res
		return
	}
	
	if len(f_name_res_json.AkaResult) > 0 {
		
		sdn_entry_res, err = this.QueryMetaphoneServer("lookup_sdn_entry", f_name_res_json.AkaResult[0])
		if err != nil {
			return
		}
		res = sdn_entry_res
		return
	}
	
	if len(f_name_res_json.RevAkaResult) > 0 {
		
		sdn_entry_res, err = this.QueryMetaphoneServer("lookup_sdn_entry", f_name_res_json.RevAkaResult[0])
		if err != nil {
			return
		}
		res = sdn_entry_res
		return
	}
	
	if len(f_add_res_json.AddressResult) > 0 {
		
		sdn_entry_res, err = this.QueryMetaphoneServer("lookup_sdn_entry", f_add_res_json.AddressResult[0])
		if err != nil {
			return
		}
		res = sdn_entry_res
		return
	}
	
	if len(f_add_res_json.PostalCodeResult) > 0 {
		
		sdn_entry_res, err = this.QueryMetaphoneServer("lookup_sdn_entry", f_add_res_json.PostalCodeResult[0])
		if err != nil {
			return
		}
		res = sdn_entry_res
		return
	}
	
	if len(m_name_res_json.NameResult) > 0 {
	
		sdn_entry_res, err = this.QueryMetaphoneServer("lookup_sdn_entry", m_name_res_json.NameResult[0])
		if err != nil {
			return
		}
		res = sdn_entry_res
		return
	}
	
	if len(m_name_res_json.RevNameResult) > 0 {
		
		sdn_entry_res, err = this.QueryMetaphoneServer("lookup_sdn_entry", m_name_res_json.RevNameResult[0])
		if err != nil {
			return
		}
		res = sdn_entry_res
		return
	}
	
	if len(m_name_res_json.AkaResult) > 0 {
		
		sdn_entry_res, err = this.QueryMetaphoneServer("lookup_sdn_entry", m_name_res_json.AkaResult[0])
		if err != nil {
			return
		}
		res = sdn_entry_res
		return
	}
	
	if len(m_name_res_json.RevAkaResult) > 0 {
		
		sdn_entry_res, err = this.QueryMetaphoneServer("lookup_sdn_entry", m_name_res_json.RevAkaResult[0])
		if err != nil {
			return
		}
		res = sdn_entry_res
		return
	}
	
	if len(m_add_res_json.AddressResult) > 0 {
		
		sdn_entry_res, err = this.QueryMetaphoneServer("lookup_sdn_entry", m_add_res_json.AddressResult[0])
		if err != nil {
			return
		}
		res = sdn_entry_res
		return
	}
	
	if len(m_add_res_json.PostalCodeResult) > 0 {
		
		sdn_entry_res, err = this.QueryMetaphoneServer("lookup_sdn_entry", m_add_res_json.PostalCodeResult[0])
		if err != nil {
			return
		}
		res = sdn_entry_res
		return
	}
	
	if len(dm_name_res_json.NameResult1) > 0 {
	
		sdn_entry_res, err = this.QueryMetaphoneServer("lookup_sdn_entry", dm_name_res_json.NameResult1[0])
		if err != nil {
			return
		}
		res = sdn_entry_res
		return
	}
	
	if len(dm_name_res_json.NameResult2) > 0 {
	
		sdn_entry_res, err = this.QueryMetaphoneServer("lookup_sdn_entry", dm_name_res_json.NameResult2[0])
		if err != nil {
			return
		}
		res = sdn_entry_res
		return
	}
	
	if len(dm_name_res_json.RevNameResult1) > 0 {
		
		sdn_entry_res, err = this.QueryMetaphoneServer("lookup_sdn_entry", dm_name_res_json.RevNameResult1[0])
		if err != nil {
			return
		}
		res = sdn_entry_res
		return
	}
	
	if len(dm_name_res_json.RevNameResult2) > 0 {
		
		sdn_entry_res, err = this.QueryMetaphoneServer("lookup_sdn_entry", dm_name_res_json.RevNameResult2[0])
		if err != nil {
			return
		}
		res = sdn_entry_res
		return
	}
	
	if len(dm_name_res_json.AkaResult1) > 0 {
		
		sdn_entry_res, err = this.QueryMetaphoneServer("lookup_sdn_entry", dm_name_res_json.AkaResult1[0])
		if err != nil {
			return
		}
		res = sdn_entry_res
		return
	}
	
	if len(dm_name_res_json.AkaResult2) > 0 {
		
		sdn_entry_res, err = this.QueryMetaphoneServer("lookup_sdn_entry", dm_name_res_json.AkaResult2[0])
		if err != nil {
			return
		}
		res = sdn_entry_res
		return
	}
	
	if len(dm_name_res_json.RevAkaResult1) > 0 {
		
		sdn_entry_res, err = this.QueryMetaphoneServer("lookup_sdn_entry", dm_name_res_json.RevAkaResult1[0])
		if err != nil {
			return
		}
		res = sdn_entry_res
		return
	}
	
	if len(dm_name_res_json.RevAkaResult2) > 0 {
		
		sdn_entry_res, err = this.QueryMetaphoneServer("lookup_sdn_entry", dm_name_res_json.RevAkaResult2[0])
		if err != nil {
			return
		}
		res = sdn_entry_res
		return
	}
	
	if len(dm_add_res_json.AddressResult1) > 0 {
		
		sdn_entry_res, err = this.QueryMetaphoneServer("lookup_sdn_entry", dm_add_res_json.AddressResult1[0])
		if err != nil {
			return
		}
		res = sdn_entry_res
		return
	}
	
	if len(dm_add_res_json.AddressResult2) > 0 {
		
		sdn_entry_res, err = this.QueryMetaphoneServer("lookup_sdn_entry", dm_add_res_json.AddressResult2[0])
		if err != nil {
			return
		}
		res = sdn_entry_res
		return
	}
	
	if len(dm_add_res_json.PostalCodeResult1) > 0 {
		
		sdn_entry_res, err = this.QueryMetaphoneServer("lookup_sdn_entry", dm_add_res_json.PostalCodeResult1[0])
		if err != nil {
			return
		}
		res = sdn_entry_res
		return
	}
	
	if len(dm_add_res_json.PostalCodeResult2) > 0 {
		
		sdn_entry_res, err = this.QueryMetaphoneServer("lookup_sdn_entry", dm_add_res_json.PostalCodeResult2[0])
		if err != nil {
			return
		}
		res = sdn_entry_res
		return
	}
	
	return
}
