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
	"fmt"
)

type kycAmlServerS struct {
	Conf *KycAmlServerConfS
	Data *SdnListS
	DataLocked bool
	FuzzyModel *fuzzy.Model 
}

type KycAmlServerConfS struct {
	Host 		string	`json:"host,omitempty"`
	Port 		string	`json:"port,omitempty"`
	Protocol 	string	`json:"protocol,omitempty"`
	DataUrl		string	`json:"data_url,omitempty"`
}

func NewKycAmlServer(conf_filename string) (new_kycamlserver *kycAmlServerS, err error) {
	
	new_kycamlserver = &kycAmlServerS{}
	
	err = new_kycamlserver.LoadConf(conf_filename)
	if err != nil {
		return
	}
	
	err = new_kycamlserver.LoadData(new_kycamlserver.Conf.DataUrl)
	if err != nil {
		return
	}
	
	new_kycamlserver.FuzzyModel = fuzzy.NewModel()
	
	// For testing only, this is not advisable on production
    //new_kycamlserver.FuzzyModel.SetThreshold(1)
    
    // This expands the distance searched, but costs more resources (memory and time). 
    // For spell checking, "2" is typically enough, for query suggestions this can be higher
    new_kycamlserver.FuzzyModel.SetDepth(4)
    
    new_kycamlserver.FuzzyTrain()
	
	return
}

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

func (this *kycAmlServerS) LoadData(url string) (err error) {
	
	this.DataLocked = true
	
	client := http.Client{}
	
	log.Printf("Retrieving dataset. Please wait before performing any queries...")
	
	sdn_res, err := client.Get(url)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	
	log.Printf("Parsing dataset")
	
	d := xml.NewDecoder(sdn_res.Body)
	err = d.Decode(&this.Data)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	
	this.DataLocked = false
	log.Printf("Dataset loaded.")

	return
}

func (this *kycAmlServerS) FuzzyTrain() {
	
	training_set := []string{}
	
	for idx, sdn_entry := range this.Data.SdnEntries {
		_ = idx
		/*
		if idx > 10 {
			break
		}
		*/
		
		training_set = append(training_set, sdn_entry.FirstName, sdn_entry.LastName)
		
		// TODO: train akas and addresses
	}
	
	log.Printf("Training fuzzy search.")
	
	this.FuzzyModel.Train(training_set)
	
	log.Printf("Fuzzy search training complete.")
}

func (this *kycAmlServerS) Listen() (err error) {
	
	l, err := net.Listen(this.Conf.Protocol, this.Conf.Host+":"+this.Conf.Port)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	defer l.Close()
	
	log.Printf("Server listening. You can perform queries now.")
	
	for {
		con, err := l.Accept()
		if err != nil {
			log.Printf("Error: %v", err)
			continue
		}
		
		go this.handleRequest(con)
	}
}

func (this *kycAmlServerS) handleRequest(con net.Conn) {
		
	conbuf := bufio.NewReader(con)
	
	res, err := conbuf.ReadBytes('\n')
	if err != nil {
		log.Printf("Error: %v", err)
		con.Close()
		return
	}
	
	var socketMsg SocketMsgS
	err = json.Unmarshal(res[:len(res)-1], &socketMsg)
	if err != nil {
		log.Printf("Error: %v", err)
		con.Close()
		return
	}
	
	switch socketMsg.Action {
		
	case "load_data":
		err = this.LoadData(this.Conf.DataUrl)
		if err != nil {
			return
		}
		con.Write([]byte(`{"result": "Data reloaded"}`+"\n"))
		
	case "query":
		log.Printf("Running query")
		
		q_result := this.FuzzyModel.SpellCheck(socketMsg.Value)
		
		log.Printf("Query result: %v", q_result)
		con.Write([]byte(fmt.Sprintf(`{"result": "%v"}`+"\n", q_result)))
	}
	
	con.Close()
}
