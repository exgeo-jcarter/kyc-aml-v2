/*
	This is for checking names and addresses against a blacklist.
*/

package KycAmlData

import (
	"io/ioutil"
	"encoding/json"
	"encoding/xml"
	"net/http"
	"log"
	"net"
	"bufio"
)

type kycAmlDataS struct {
	Conf *KycAmlDataConfS
	SdnList []byte
	SdnListLocked bool
}

// Server details, loaded from config.json
type KycAmlDataConfS struct {
	Host 		string	`json:"host,omitempty"`
	Port 		string	`json:"port,omitempty"`
	Protocol 	string	`json:"protocol,omitempty"`
	SdnListUrl	string	`json:"sdn_list_url,omitempty"`
}

// Load the server.
func NewKycAmlData(conf_filename string) (new_kycamldata *kycAmlDataS, err error) {
	
	new_kycamldata = &kycAmlDataS{}
	
	// Load the server configuration.
	err = new_kycamldata.LoadConf(conf_filename)
	if err != nil {
		return
	}
	
	return
}

// Load the server configuration.
func (this *kycAmlDataS) LoadConf(filename string) (err error) {
	
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
	
	log.Printf("Data server config file loaded.")
	return
}

// Load the blacklist data.
func (this *kycAmlDataS) LoadSdnList(url string) (err error) {
	
	// Lock the dataset.
	this.SdnListLocked = true
	
	client := http.Client{}
	
	log.Printf("Retrieving dataset. Please wait before performing any queries...")
	
	sdn_res, err := client.Get(url)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	
	log.Printf("Parsing XML dataset.")
	
	d := xml.NewDecoder(sdn_res.Body)
	
	var sdn_list_struct *SdnListS
	err = d.Decode(&sdn_list_struct)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	
	log.Printf("Converting dataset to JSON.")
	this.SdnList, err = json.Marshal(sdn_list_struct)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	
	// Unlock the dataset.
	this.SdnListLocked = false
	log.Printf("Dataset loaded. You can perform queries now.")

	return
}

// Listen for new connections.
func (this *kycAmlDataS) Listen() (err error) {
	
	// Listen.
	l, err := net.Listen(this.Conf.Protocol, this.Conf.Host+":"+this.Conf.Port)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	defer l.Close()
	
	log.Printf("Data server listening.")
	
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
func (this *kycAmlDataS) handleRequest(con net.Conn) {
	
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
	case "load_sdn_list":
		
		if len(this.SdnList) != 0 {
			// Respond to client.
			con.Write([]byte(`{"result": "SDN list already loaded."}`+"\n"))
			con.Close()
			return
		}
		
		// Reload blacklist.
		err = this.LoadSdnList(this.Conf.SdnListUrl)
		if err != nil {
			con.Close()
			return
		}
		
		// Respond to client.
		con.Write([]byte(`{"result": "SDN list loaded."}`+"\n"))
		
	case "load_sdn_list_force":
		
		// Reload blacklist.
		err = this.LoadSdnList(this.Conf.SdnListUrl)
		if err != nil {
			con.Close()
			return
		}
		
		// Respond to client.
		con.Write([]byte(`{"result": "SDN list loaded."}`+"\n"))
		
	// If a query was requested.
	case "get_sdn_list":
	
		// Responsd with blacklist data.
		this.SdnList = append(this.SdnList, []byte("\n")...)
		con.Write(this.SdnList)
	}
	
	// Close the client's connection.
	con.Close()
}
