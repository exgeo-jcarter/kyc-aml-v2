/*
	This is for checking names and addresses against a blacklist.
*/

package KycAmlMetaphone

import (
	"io/ioutil"
	"encoding/json"
	"log"
	"net"
	"bufio"
	"strings"
	"github.com/dotcypress/phonetics"
)

type KycAmlMetaphoneS struct {
	Conf 					*KycAmlMetaphoneConfS
	SdnList					*SdnListS
	SdnListMapNames			map[string]string
	SdnListMapRevNames		map[string]string
	SdnListMapAkas			map[string]string
	SdnListMapRevAkas		map[string]string
	SdnListMapAddresses		map[string]string
	SdnListMapPostalCodes 	map[string]string
}

// Server details, loaded from config.json
type KycAmlMetaphoneConfS struct {
	Host 		string	`json:"host,omitempty"`
	Port 		string	`json:"port,omitempty"`
	Protocol 	string	`json:"protocol,omitempty"`
}

// Load the server.
func NewKycAmlMetaphone(conf_filename string) (new_kycamlfuzzy *KycAmlMetaphoneS, err error) {
	
	new_kycamlfuzzy = &KycAmlMetaphoneS{
		SdnListMapNames: 		map[string]string{},
		SdnListMapRevNames: 	map[string]string{},
		SdnListMapAkas: 		map[string]string{},
		SdnListMapRevAkas: 		map[string]string{},
		SdnListMapAddresses:	map[string]string{},
		SdnListMapPostalCodes: 	map[string]string{},
		
	}
	
	// Load the server configuration.
	err = new_kycamlfuzzy.LoadConf(conf_filename)
	if err != nil {
		return
	}
	
	return
}

// Load the server configuration.
func (this *KycAmlMetaphoneS) LoadConf(filename string) (err error) {
	
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
	
	log.Printf("Metaphone server config file loaded.")
	return
}



// Train the fuzzy search with the blacklist data.
func (this *KycAmlMetaphoneS) TrainSdn(sdn_list string) (err error) {
	
	log.Printf("Parsing JSON SDN list.")
	
	err = json.Unmarshal([]byte(sdn_list), &this.SdnList)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	
	log.Printf("Training Metaphone SDN list.")
	
	// Loop over all the entries in the SDN blacklist.
	for _, sdn_entry := range this.SdnList.SdnEntries {
		
		// Names to lowercase.
		firstname_lower := strings.ToLower(sdn_entry.FirstName)
		lastname_lower := strings.ToLower(sdn_entry.LastName)
		name := firstname_lower+" "+lastname_lower
		name_metaphone := strings.ToLower(phonetics.EncodeMetaphone(name))
		this.SdnListMapNames[name_metaphone] = name
		revname := lastname_lower+" "+firstname_lower
		revname_metaphone := strings.ToLower(phonetics.EncodeMetaphone(revname))
		this.SdnListMapRevNames[revname_metaphone] = revname
		
		// Loop over all AKAs.
		for _, aka_list := range sdn_entry.AkaList.Akas {
			
			// AKA names to lowercase.
			aka_firstname_lower := strings.ToLower(aka_list.FirstName)
			aka_lastname_lower := strings.ToLower(aka_list.LastName)
			aka := aka_firstname_lower+" "+aka_lastname_lower
			aka_metaphone := strings.ToLower(phonetics.EncodeMetaphone(aka))
			this.SdnListMapAkas[aka_metaphone] = aka
			revaka := aka_lastname_lower+" "+aka_firstname_lower
			revaka_metaphone := strings.ToLower(phonetics.EncodeMetaphone(revaka))
			this.SdnListMapRevAkas[revaka_metaphone] = revaka
		}
		
		// Loop over all addresses.
		for _, address_list := range sdn_entry.AddressList.Addresses {
			
			// Addresses to lowercase.
			address1_lower := strings.ToLower(address_list.Address1)
			address1_metaphone := strings.ToLower(phonetics.EncodeMetaphone(address1_lower))
			this.SdnListMapAddresses[address1_metaphone] = address1_lower
			
			postalcode_lower := strings.ToLower(address_list.PostalCode)
			postalcode_metaphone := strings.ToLower(phonetics.EncodeMetaphone(postalcode_lower))
			this.SdnListMapPostalCodes[postalcode_metaphone] = postalcode_lower
		}
	}
	
	log.Printf("Metaphone search training complete. You can perform queries now.")
	
	return
}

// Listen for new connections.
func (this *KycAmlMetaphoneS) Listen() (err error) {
	
	// Listen.
	l, err := net.Listen(this.Conf.Protocol, this.Conf.Host+":"+this.Conf.Port)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	defer l.Close()
	
	log.Printf("Metaphone server listening.")
	
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
func (this *KycAmlMetaphoneS) handleRequest(con net.Conn) {
	
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

func (this *KycAmlMetaphoneS) Query(con net.Conn, socketMsg *SocketMsgS) {
	
	// Metaphone query to lowercase.
	metaphone_query := strings.ToLower(socketMsg.Value)
	metaphone_encoded_query := strings.ToLower(phonetics.EncodeMetaphone(metaphone_query))

	name_q_result := []string{}
	revname_q_result := []string{}
	aka_q_result := []string{}
	revaka_q_result := []string{}
	address_q_result := []string{}
	postal_code_q_result := []string{}

	num_queries := 0
	wait_for_query_results_ch := make(chan int)

	switch socketMsg.Action {
	
	// Search for metaphone matches.
	case "query_name":
		num_queries += 4
		
		go (func() {
			name_metaphone_result, ok := this.SdnListMapNames[metaphone_encoded_query]
			if ok && (name_metaphone_result != "") {
				name_q_result = []string{name_metaphone_result,}
			}
			wait_for_query_results_ch <- 1
		})()
		
		go (func() {
			revname_metaphone_result, ok2 := this.SdnListMapRevNames[metaphone_encoded_query]
			if ok2 && (revname_metaphone_result != "") {
				revname_q_result = []string{revname_metaphone_result,}
			}
			wait_for_query_results_ch <- 1
		})()
		
		go (func() {
			aka_metaphone_result, ok3 := this.SdnListMapAkas[metaphone_encoded_query]
			if ok3 && (aka_metaphone_result != "") {
				aka_q_result = []string{aka_metaphone_result,}
			}
			wait_for_query_results_ch <- 1
		})()
		
		go (func() {
			revaka_metaphone_result, ok4 := this.SdnListMapRevAkas[metaphone_encoded_query]
			if ok4 && (revaka_metaphone_result != "") {
				revaka_q_result = []string{revaka_metaphone_result,}
			}
			wait_for_query_results_ch <- 1
		})()
		
	case "query_address":
		num_queries += 2
		
		go (func() {
			address_metaphone_result, ok := this.SdnListMapAddresses[metaphone_encoded_query]
			if ok && (address_metaphone_result != "") {
				address_q_result = []string{address_metaphone_result,}
			}
			wait_for_query_results_ch <- 1
		})()
		
		go (func() {
			postal_code_metaphone_result, ok2 := this.SdnListMapPostalCodes[metaphone_encoded_query]
			if ok2 && (postal_code_metaphone_result != "") {
				postal_code_q_result = []string{postal_code_metaphone_result,}
			}
			wait_for_query_results_ch <- 1
		})()
	}
	
	for i := 0; i < num_queries; i++ {
		<- wait_for_query_results_ch
	}
	
	res_struct := &QueryResS{
		Query: metaphone_query,
		EncodedQuery: metaphone_encoded_query,
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
