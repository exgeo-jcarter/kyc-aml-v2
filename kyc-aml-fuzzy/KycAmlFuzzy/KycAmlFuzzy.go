/*
	This is for checking names and addresses against a blacklist.
*/

package KycAmlFuzzy

import (
	"io/ioutil"
	"encoding/json"
	"log"
	"net"
	"bufio"
	"github.com/sajari/fuzzy"
	"strings"
)

type KycAmlFuzzyS struct {
	Conf 					*KycAmlFuzzyConfS
	SdnList					*SdnListS
	SdnListLocked			bool
	FuzzyModelNames 		*fuzzy.Model
	FuzzyModelRevNames 		*fuzzy.Model
	FuzzyModelAkas	 		*fuzzy.Model
	FuzzyModelRevAkas 		*fuzzy.Model
	FuzzyModelAddresses		*fuzzy.Model
	FuzzyModelPostalCodes	*fuzzy.Model
}

// Server details, loaded from config.json
type KycAmlFuzzyConfS struct {
	Host 		string	`json:"host,omitempty"`
	Port 		string	`json:"port,omitempty"`
	Protocol 	string	`json:"protocol,omitempty"`
}

// Load the server.
func NewKycAmlFuzzy(conf_filename string) (new_kycamlfuzzy *KycAmlFuzzyS, err error) {
	
	new_kycamlfuzzy = &KycAmlFuzzyS{}
	
	// Load the server configuration.
	err = new_kycamlfuzzy.LoadConf(conf_filename)
	if err != nil {
		return
	}
	
	// Set up the fuzzy search model.
	new_kycamlfuzzy.FuzzyModelNames = fuzzy.NewModel()
    new_kycamlfuzzy.FuzzyModelNames.SetThreshold(1)
    new_kycamlfuzzy.FuzzyModelNames.SetDepth(2)
    new_kycamlfuzzy.FuzzyModelNames.SetUseAutocomplete(false)
    
    new_kycamlfuzzy.FuzzyModelRevNames = fuzzy.NewModel()
    new_kycamlfuzzy.FuzzyModelRevNames.SetThreshold(1)
    new_kycamlfuzzy.FuzzyModelRevNames.SetDepth(2)
    new_kycamlfuzzy.FuzzyModelRevNames.SetUseAutocomplete(false)
    
    new_kycamlfuzzy.FuzzyModelAkas = fuzzy.NewModel()
    new_kycamlfuzzy.FuzzyModelAkas.SetThreshold(1)
    new_kycamlfuzzy.FuzzyModelAkas.SetDepth(2)
    new_kycamlfuzzy.FuzzyModelAkas.SetUseAutocomplete(false)
    
    new_kycamlfuzzy.FuzzyModelRevAkas = fuzzy.NewModel()
    new_kycamlfuzzy.FuzzyModelRevAkas.SetThreshold(1)
    new_kycamlfuzzy.FuzzyModelRevAkas.SetDepth(2)
    new_kycamlfuzzy.FuzzyModelRevAkas.SetUseAutocomplete(false)
    
    new_kycamlfuzzy.FuzzyModelAddresses = fuzzy.NewModel()
    new_kycamlfuzzy.FuzzyModelAddresses.SetThreshold(1)
    new_kycamlfuzzy.FuzzyModelAddresses.SetDepth(2)
    new_kycamlfuzzy.FuzzyModelAddresses.SetUseAutocomplete(false)
    
    new_kycamlfuzzy.FuzzyModelPostalCodes = fuzzy.NewModel()
    new_kycamlfuzzy.FuzzyModelPostalCodes.SetThreshold(1)
    new_kycamlfuzzy.FuzzyModelPostalCodes.SetDepth(2)
    new_kycamlfuzzy.FuzzyModelPostalCodes.SetUseAutocomplete(false)
	
	return
}

// Load the server configuration.
func (this *KycAmlFuzzyS) LoadConf(filename string) (err error) {
	
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
	
	log.Printf("Fuzzy server config file loaded.")
	return
}



// Train the fuzzy search with the blacklist data.
func (this *KycAmlFuzzyS) TrainSdn(sdn_list string) (err error) {
	
	log.Printf("Parsing JSON SDN list.")
	
	err = json.Unmarshal([]byte(sdn_list), &this.SdnList)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	
	training_set_names := []string{}
	training_set_revnames := []string{}
	training_set_akas := []string{}
	training_set_revakas := []string{}
	training_set_addresses := []string{}
	training_set_postal_codes := []string{}
	
	// Loop over all the entries in the SDN blacklist.
	for _, sdn_entry := range this.SdnList.SdnEntries {
		
		// Names to lowercase.
		firstname_lower := strings.ToLower(sdn_entry.FirstName)
		lastname_lower := strings.ToLower(sdn_entry.LastName)
		
		// Add names to training sets.
		training_set_names = append(training_set_names, firstname_lower+" "+lastname_lower)
		training_set_revnames = append(training_set_revnames, lastname_lower+" "+firstname_lower)
		
		// Loop over all AKAs.
		for _, aka_list := range sdn_entry.AkaList.Akas {
			
			// AKA names to lowercase.
			aka_firstname_lower := strings.ToLower(aka_list.FirstName)
			aka_lastname_lower := strings.ToLower(aka_list.LastName)
			
			// Add AKA names to training set.
			training_set_akas = append(training_set_akas, aka_firstname_lower+" "+aka_lastname_lower)
			training_set_revakas = append(training_set_revakas, aka_lastname_lower+" "+aka_firstname_lower)
		}
		
		// Loop over all addresses.
		for _, address_list := range sdn_entry.AddressList.Addresses {
			
			// Addresses to lowercase.
			address1_lower := strings.ToLower(address_list.Address1)
			postalcode_lower := strings.ToLower(address_list.PostalCode)
			
			// Add addresses to training set.
			training_set_addresses = append(training_set_addresses, address1_lower)
			training_set_postal_codes = append(training_set_postal_codes, postalcode_lower)
		}
	}
	
	num_trainers := 6
	wait_for_training_ch := make(chan int)
	
	// Train fuzzy search using the training sets.
	go (func() {
		log.Printf("Training fuzzy names.")
		this.FuzzyModelNames.Train(training_set_names)
		log.Printf("Training fuzzy names complete.")
		wait_for_training_ch <- 1
	})()
	
	go (func() {
		log.Printf("Training fuzzy reverse names.")
		this.FuzzyModelRevNames.Train(training_set_revnames)
		log.Printf("Training fuzzy reverse names complete.")
		wait_for_training_ch <- 1
	})()
	
	go (func() {
		log.Printf("Training fuzzy akas.")
		this.FuzzyModelAkas.Train(training_set_akas)
		log.Printf("Training fuzzy akas complete.")
		wait_for_training_ch <- 1
	})()
	
	go (func() {
		log.Printf("Training fuzzy reverse akas.")
		this.FuzzyModelRevAkas.Train(training_set_revakas)
		log.Printf("Training fuzzy reverse akas complete.")
		wait_for_training_ch <- 1
	})()
	
	go (func() {
		log.Printf("Training fuzzy addresses.")
		this.FuzzyModelAddresses.Train(training_set_addresses)
		log.Printf("Training fuzzy addresses complete.")
		wait_for_training_ch <- 1
	})()
	
	go (func() {
		log.Printf("Training fuzzy postal codes.")
		this.FuzzyModelPostalCodes.Train(training_set_postal_codes)
		log.Printf("Training fuzzy postal codes complete.")
		wait_for_training_ch <- 1
	})()
	
	for i := 0; i < num_trainers; i++ {
		<- wait_for_training_ch
	}
	
	log.Printf("Fuzzy search training complete. You can perform queries now.")
	
	return
}

// Listen for new connections.
func (this *KycAmlFuzzyS) Listen() (err error) {
	
	// Listen.
	l, err := net.Listen(this.Conf.Protocol, this.Conf.Host+":"+this.Conf.Port)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	defer l.Close()
	
	log.Printf("Fuzzy server listening.")
	
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
func (this *KycAmlFuzzyS) handleRequest(con net.Conn) {
	
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
	case "train_sdn":
		
		// If we've already trained, don't train again.
		if this.SdnList != nil {
			con.Write([]byte(`{"result": "Already trained."}`+"\n"))
			con.Close()
			return
		}
		
		// Train fuzzy search with blacklist data.
		err = this.TrainSdn(socketMsg.Value)
		if err != nil {
			con.Close()
			return
		}
		
		// Respond to client.
		con.Write([]byte(`{"result": "Training SDN list complete."}`+"\n"))
		
	// If data reload was forced.
	case "train_sdn_force":
		
		// Train fuzzy search with blacklist data.
		err = this.TrainSdn(socketMsg.Value)
		if err != nil {
			con.Close()
			return
		}
		
		// Respond to client.
		con.Write([]byte(`{"result": "Training SDN list complete."}`+"\n"))
		
	// If a query was requested, query all fuzzy models.
	default:
		this.Query(con, &socketMsg)
	
	}
	
	// Close the client's connection.
	con.Close()
}

func (this *KycAmlFuzzyS) Query(con net.Conn, socketMsg *SocketMsgS) {
	
	// Fuzzy query to lowercase.
	fuzzy_query := strings.ToLower(socketMsg.Value)

	name_q_result := []string{}
	revname_q_result := []string{}
	aka_q_result := []string{}
	revaka_q_result := []string{}
	address_q_result := []string{}
	postal_code_q_result := []string{}

	num_queries := 0
	wait_for_query_results_ch := make(chan int)

	switch socketMsg.Action {
	
	// Search for fuzzy matches.
	case "query_name":
		num_queries += 4
	
		go (func() {
			name_q_result = this.FuzzyModelNames.Suggestions(fuzzy_query, false)
			
			if len(name_q_result) > 10 {
				name_q_result = name_q_result[10:]
			} else if len(name_q_result) == 10 {
				name_q_result = []string{}
			}
			
			wait_for_query_results_ch <- 1
		})()
		
		go (func() {
			revname_q_result = this.FuzzyModelRevNames.Suggestions(fuzzy_query, false)
			
			if len(revname_q_result) > 10 {
				revname_q_result = revname_q_result[10:]
			} else if len(revname_q_result) == 10 {
				revname_q_result = []string{}
			}
			
			wait_for_query_results_ch <- 1
		})()
		
		go (func() {
			aka_q_result = this.FuzzyModelAkas.Suggestions(fuzzy_query, false)
			
			if len(aka_q_result) > 10 {
				aka_q_result = aka_q_result[10:]
			} else if len(aka_q_result) == 10 {
				aka_q_result = []string{}
			}
			
			wait_for_query_results_ch <- 1
		})()
		
		go (func() {
			revaka_q_result = this.FuzzyModelRevAkas.Suggestions(fuzzy_query, false)
			
			if len(revaka_q_result) > 10 {
				revaka_q_result = revaka_q_result[10:]
			} else if len(revaka_q_result) == 10 {
				revaka_q_result = []string{}
			}
			
			wait_for_query_results_ch <- 1
		})()
		
	case "query_address":
		num_queries += 2
	
		go (func() {
			address_q_result = this.FuzzyModelAddresses.Suggestions(fuzzy_query, false)
			
			if len(address_q_result) > 10 {
				address_q_result = address_q_result[10:]
			} else if len(address_q_result) == 10 {
				address_q_result = []string{}
			}
			
			wait_for_query_results_ch <- 1
		})()
		
		go (func() {
			postal_code_q_result = this.FuzzyModelPostalCodes.Suggestions(fuzzy_query, false)
			
			if len(postal_code_q_result) > 10 {
				postal_code_q_result = postal_code_q_result[10:]
			} else if len(postal_code_q_result) == 10 {
				postal_code_q_result = []string{}
			}
			
			wait_for_query_results_ch <- 1
		})()
	}
	
	for i := 0; i < num_queries; i++ {
		<- wait_for_query_results_ch
	}
	
	res_struct := &QueryResS{
		Query: fuzzy_query,
		NameResult: name_q_result,
		RevNameResult: revname_q_result,
		AkaResult: aka_q_result,
		RevAkaResult: revaka_q_result,
		AddressResult: address_q_result,
		PostalCodeResult: postal_code_q_result,
	}
	
	res_bytes, err := json.Marshal(res_struct)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	
	res_bytes = append(res_bytes, []byte("\n")...)
	con.Write(res_bytes)
}
