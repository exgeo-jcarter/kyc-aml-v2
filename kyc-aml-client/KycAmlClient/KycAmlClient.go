/*
	This is for checking names and addresses against a blacklist.
*/

package KycAmlClient

import (
	"net"
	"encoding/json"
	"log"
	"bufio"
	"io/ioutil"
	"strings"
	//"fmt"
)

type KycAmlClientS struct {
	Conf 	*KycAmlClientConfS
}

type KycAmlClientConfS struct {
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
func NewKycAmlClient(conf_filename string) (new_kycamlclient *KycAmlClientS, err error) {
	
	new_kycamlclient = &KycAmlClientS{}
	
	// Load server settings.
	err = new_kycamlclient.LoadConf(conf_filename)
	if err != nil {
		return
	}
	
	return
}

// Load server settings.
func (this *KycAmlClientS) LoadConf(filename string) (err error) {
	
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

// Query the fuzzy server to check for string matches against the blacklist.
func (this *KycAmlClientS) QueryServer(protocol, host, port, action, value string) (res string, err error) {
	
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

func (this *KycAmlClientS) QueryDataServer(action, value string) (res string, err error) {
	
	res, err = this.QueryServer(this.Conf.DataProtocol, this.Conf.DataHost, this.Conf.DataPort, action, value)
	return
}

func (this *KycAmlClientS) QueryFuzzyServer(action, value string) (res string, err error) {
	
	res, err = this.QueryServer(this.Conf.FuzzyProtocol, this.Conf.FuzzyHost, this.Conf.FuzzyPort, action, value)
	return
}

func (this *KycAmlClientS) QueryMetaphoneServer(action, value string) (res string, err error) {
	
	res, err = this.QueryServer(this.Conf.MetaphoneProtocol, this.Conf.MetaphoneHost, this.Conf.MetaphonePort, action, value)
	return
}

func (this *KycAmlClientS) QueryDoubleMetaphoneServer(action, value string) (res string, err error) {
	
	res, err = this.QueryServer(this.Conf.DoubleMetaphoneProtocol, this.Conf.DoubleMetaphoneHost, this.Conf.DoubleMetaphonePort, action, value)
	return
}

func (this *KycAmlClientS) CalculateRiskScore(q, aq, f_name_res, f_add_res, m_name_res, m_add_res, dm_name_res, dm_add_res string) (score float64, err error) {
	
	num_queries := 0.0
	num_results := 0.0
	
	if q != "" {
		num_queries += 6
	}
	
	if aq != "" {
		num_queries += 4
	}
	
	var f_name_res_json FuzzyQueryResS
	err = json.Unmarshal([]byte(f_name_res), &f_name_res_json)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	if len(f_name_res_json.NameResult) > 0 {
		num_results++
	}
	if len(f_name_res_json.RevNameResult) > 0 {
		num_results++
	}
	if len(f_name_res_json.AkaResult) > 0 {
		num_results++
	}
	if len(f_name_res_json.RevAkaResult) > 0 {
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
	}
	if len(f_add_res_json.PostalCodeResult) > 0 {
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
	}
	if len(m_name_res_json.RevNameResult) > 0 {
		num_results++
	}
	if len(m_name_res_json.AkaResult) > 0 {
		num_results++
	}
	if len(m_name_res_json.RevAkaResult) > 0 {
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
	}
	if len(m_add_res_json.PostalCodeResult) > 0 {
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
	}
	if len(dm_name_res_json.NameResult2) > 0 {
		num_results++
	}
	if len(dm_name_res_json.RevNameResult1) > 0 {
		num_results++
	}
	if len(dm_name_res_json.RevNameResult2) > 0 {
		num_results++
	}
	if len(dm_name_res_json.AkaResult1) > 0 {
		num_results++
	}
	if len(dm_name_res_json.AkaResult2) > 0 {
		num_results++
	}
	if len(dm_name_res_json.RevAkaResult1) > 0 {
		num_results++
	}
	if len(dm_name_res_json.RevAkaResult2) > 0 {
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
	}
	if len(dm_add_res_json.AddressResult2) > 0 {
		num_results++
	}
	if len(dm_add_res_json.PostalCodeResult1) > 0 {
		num_results++
	}
	if len(dm_add_res_json.PostalCodeResult2) > 0 {
		num_results++
	}
	
	score = num_results / num_queries * 100.0
	return
}

// Looks up an entry in the SDN list.
func (this *KycAmlClientS) LookupSdnEntry(f_name_res, f_add_res, m_name_res, m_add_res, dm_name_res, dm_add_res string) (res string, err error) {
	
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

/*
// Calculates amount of risk, based on how close our match was to the original query.
func (this *kycAmlFuzzyS) CalculateRiskScore(q, mq string, res, mres []string) (score float64) {
	
	var q_score float64
	var mq_score float64
	
	// Calculate fuzzy risk.
	for _, val := range res {
	
		var q_score2 float64
	
		if (len(val) > 0) && (len(q) >= len(val)) {
		
			for idx2, _ := range val {
				
				if q[idx2] == val[idx2] {
					q_score2++
				}
			}
			
			q_score2 -= (float64(len(q)) - float64(len(val)))
			q_score2 /= (float64(len(q))) / 100
		
		} else if (len(val) > 0) && (len(q) < len(val)){
			
			for idx2, _ := range q {
				
				if q[idx2] == val[idx2] {
					q_score2++
				}
			}
			
			q_score2 -= (float64(len(val)) - float64(len(q)))
			q_score2 /= (float64(len(val))) / 100
		}
		
		if q_score2 > q_score {
			q_score = q_score2
		}
	}
	
	if len(mres) > 0 {
		mq_score = 100
	}
	
	score = (q_score + mq_score) / 2
	
	return
}
*/
