/*
	This is for checking names and addresses against a blacklist.
*/

package main

import (
	"io/ioutil"
	"encoding/json"
	"encoding/xml"
	"net/http"
	"log"
	"net"
	"bufio"
	"github.com/sajari/fuzzy"
	"strings"
	"github.com/dotcypress/phonetics"
)

type kycAmlServerS struct {
	Conf *KycAmlServerConfS
	Data *SdnListS
	DataLocked bool
	FuzzyModel *fuzzy.Model 
}

// Server details, loaded from config.json
type KycAmlServerConfS struct {
	Host 		string	`json:"host,omitempty"`
	Port 		string	`json:"port,omitempty"`
	Protocol 	string	`json:"protocol,omitempty"`
	DataUrl		string	`json:"data_url,omitempty"`
}

// Load the server.
func NewKycAmlServer(conf_filename string) (new_kycamlserver *kycAmlServerS, err error) {
	
	new_kycamlserver = &kycAmlServerS{}
	
	// Load the server configuration.
	err = new_kycamlserver.LoadConf(conf_filename)
	if err != nil {
		return
	}
	
	// Load the blacklist data.
	err = new_kycamlserver.LoadData(new_kycamlserver.Conf.DataUrl)
	if err != nil {
		return
	}
	
	// Set up the fuzzy search model.
	new_kycamlserver.FuzzyModel = fuzzy.NewModel()
    new_kycamlserver.FuzzyModel.SetThreshold(1)
    new_kycamlserver.FuzzyModel.SetDepth(2)
    new_kycamlserver.FuzzyModel.SetUseAutocomplete(false)
    
    // Train fuzzy search on the blacklist data.
    new_kycamlserver.FuzzyTrain()
	
	return
}

// Load the server configuration.
func (this *kycAmlServerS) LoadConf(filename string) (err error) {
	
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
	
	log.Printf("Config file loaded.")
	return
}

// Load the blacklist data.
func (this *kycAmlServerS) LoadData(url string) (err error) {
	
	// Lock the dataset.
	this.DataLocked = true
	
	client := http.Client{}
	
	log.Printf("Retrieving dataset. Please wait before performing any queries...")
	
	sdn_res, err := client.Get(url)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	
	log.Printf("Parsing dataset.")
	
	d := xml.NewDecoder(sdn_res.Body)
	err = d.Decode(&this.Data)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	
	// Unlock the dataset.
	this.DataLocked = false
	log.Printf("Dataset loaded.")

	return
}

// Train the fuzzy search with the blacklist data.
func (this *kycAmlServerS) FuzzyTrain() {
	
	training_set := []string{}
	
	// Loop over all the entries in the SDN blacklist.
	for _, sdn_entry := range this.Data.SdnEntries {
		
		// Names to lowercase.
		firstname_lower := strings.ToLower(sdn_entry.FirstName)
		lastname_lower := strings.ToLower(sdn_entry.LastName)
		
		// Metaphone-encoded names.
		name_metaphone := strings.ToLower(phonetics.EncodeMetaphone(firstname_lower+" "+lastname_lower))
		revname_metaphone := strings.ToLower(phonetics.EncodeMetaphone(lastname_lower+" "+firstname_lower))
		
		// Add names to training set.
		training_set = append(training_set, firstname_lower+" "+lastname_lower, lastname_lower+" "+firstname_lower, name_metaphone, revname_metaphone)
		
		// Loop over all AKAs.
		for _, aka_list := range sdn_entry.AkaList.Akas {
			
			// AKA names to lowercase.
			aka_firstname_lower := strings.ToLower(aka_list.FirstName)
			aka_lastname_lower := strings.ToLower(aka_list.LastName)
			
			// Metaphone-encoded AKA names.
			aka_name_metaphone := strings.ToLower(phonetics.EncodeMetaphone(aka_firstname_lower+" "+aka_lastname_lower))
			aka_revname_metaphone := strings.ToLower(phonetics.EncodeMetaphone(aka_lastname_lower+" "+aka_firstname_lower))
			
			// Add AKA names to training set.
			training_set = append(training_set, aka_firstname_lower+" "+aka_lastname_lower, aka_lastname_lower+" "+aka_firstname_lower, aka_name_metaphone, aka_revname_metaphone)
		}
		
		// Loop over all addresses.
		for _, address_list := range sdn_entry.AddressList.Addresses {
			
			// Addresses to lowercase.
			address1_lower := strings.ToLower(address_list.Address1)
			postalcode_lower := strings.ToLower(address_list.PostalCode)
			
			// Metaphone-encoded addresses.
			address1_metaphone := strings.ToLower(phonetics.EncodeMetaphone(address1_lower))
			postalcode_metaphone := strings.ToLower(phonetics.EncodeMetaphone(postalcode_lower))
			
			// Add addresses to training set.
			training_set = append(training_set, address1_lower, address1_metaphone, postalcode_lower, postalcode_metaphone)
		}
		
	}
	
	log.Printf("Training fuzzy search.")
	
	// Train fuzzy search using the training set.
	this.FuzzyModel.Train(training_set)
	
	log.Printf("Fuzzy search training complete.")
}

// Listen for new connections.
func (this *kycAmlServerS) Listen() (err error) {
	
	// Listen.
	l, err := net.Listen(this.Conf.Protocol, this.Conf.Host+":"+this.Conf.Port)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	defer l.Close()
	
	log.Printf("Server listening. You can perform queries now.")
	
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
}

// Handle network requests.
func (this *kycAmlServerS) handleRequest(con net.Conn) {
	
	conbuf := bufio.NewReader(con)
	
	// Read buffer until newline.
	res, err := conbuf.ReadBytes('\n')
	if err != nil {
		log.Printf("Error: %v", err)
		con.Close()
		return
	}
	
	// Parse request into a struct.
	var socketMsg SocketMsgS
	err = json.Unmarshal(res[:len(res)-1], &socketMsg)
	if err != nil {
		log.Printf("Error: %v", err)
		con.Close()
		return
	}
	
	// Check which action was requested.
	switch socketMsg.Action {
		
	// If data reload was requested.
	case "load_data":
		
		// Reload blacklist.
		err = this.LoadData(this.Conf.DataUrl)
		if err != nil {
			return
		}
		
		// Train fuzzy search with blacklist data.
		this.FuzzyTrain()
		
		// Respond to client.
		con.Write([]byte(`{"result": "Data reloaded"}`+"\n"))
		
	// If a query was requested.
	case "query":
	
		this.Query(con, &socketMsg)
	}
	
	// Close the client's connection.
	con.Close()
}

func (this *kycAmlServerS) Query(con net.Conn, socketMsg *SocketMsgS) {
	
	// Fuzzy query to lowercase.
	fuzzy_query := strings.ToLower(socketMsg.Value)

	// Metaphone-encoded query to lowercase.
	metaphone_query := strings.ToLower(phonetics.EncodeMetaphone(strings.ToLower(socketMsg.Value)))
	
	// Search for fuzzy matches.
	q_result := this.FuzzyModel.Suggestions(strings.ToLower(socketMsg.Value), false)
	
	// Search for metaphone matches.
	q_m_result := this.FuzzyModel.Suggestions(metaphone_query, false)
	
	// If no fuzzy matches, remove empty strings from result set.
	empty := true
	for _, val := range q_result {
		if val != "" {
			empty = false
			break
		}
	}
	if empty {
		q_result = []string{}
	}
	
	// If we get a match, remove first 10 elements of array because they are always empty for some reason.
	if len(q_result) > 0 {
		q_result = q_result[10:]
	}
	
	// If no metaphone matches, remove empty strings from result set.
	empty = true
	for _, val := range q_m_result {
		if val != "" {
			empty = false
			break
		}
	}
	if empty {
		q_m_result = []string{}
	}
	
	// If we get a match, remove first 10 elements of array because they are always empty for some reason.
	if len(q_m_result) > 0 {
		q_m_result = q_m_result[10:]
	}
	
	// Create response struct.
	q_result_struct := QueryResS{
		Query: socketMsg.Value,
		MetaphoneQuery: metaphone_query,
		Result: q_result,
		MetaphoneResult: q_m_result,
		RiskScore: this.CalculateRiskScore(fuzzy_query, metaphone_query, q_result, q_m_result),
	}
	
	// Marshal response.
	q_result_json, err := json.Marshal(q_result_struct)
	if err != nil {
		log.Printf("Error: %v", err)
		con.Close()
		return
	}
	
	// Send results to client.
	log.Printf("Query result: %s", q_result_json)
	con.Write([]byte(string(q_result_json)+"\n"))
}

// Calculates amount of risk, based on how close q is to res, and how close mq is to mres.
func (this *kycAmlServerS) CalculateRiskScore(q, mq string, res, mres []string) (score float64) {
	
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
			
			q_score2 += (float64(len(q)) - float64(len(val)))
			q_score2 /= (float64(len(q))) / 100
		
		} else if (len(val) > 0) && (len(q) < len(val)){
			
			for idx2, _ := range q {
				
				if q[idx2] == val[idx2] {
					q_score2++
				}
			}
			
			q_score2 += (float64(len(val)) - float64(len(q)))
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
