package main

import (
	"net"
	"encoding/json"
	"log"
	"bufio"
	"io/ioutil"
	"fmt"
)

type kycAmlClientS struct {
	Conf *KycAmlClientConfS
}

type KycAmlClientConfS struct {
	Host 		string	`json:"host,omitempty"`
	Port 		string	`json:"port,omitempty"`
	Protocol 	string	`json:"protocol,omitempty"`
}

func NewKycAmlClient(conf_filename string) (new_kycamlclient *kycAmlClientS, err error) {
	
	new_kycamlclient = &kycAmlClientS{}
	
	err = new_kycamlclient.LoadConf(conf_filename)
	if err != nil {
		return
	}
	
	return
}

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
	
	log.Printf("Config file loaded.")
	return
}

func (this *kycAmlClientS) Query(q string) (err error) {
	
	msg := []byte(`{"action": "query", "value": "`+q+`"}`+"\n")
	
	con, err := net.Dial(this.Conf.Protocol, this.Conf.Host+":"+this.Conf.Port)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	
	resCh := make(chan int)
	
	conbuf := bufio.NewReader(con)
	
	go (func(resCh chan int, conbuf *bufio.Reader) {
		
		res := []byte{}
		
		for {
			
			res, err = conbuf.ReadBytes('\n')
			if len(res) > 0 {
				
				fmt.Printf("%s", res[:len(res)-1])
				resCh <- 1
				break
			}
			
			if err != nil {
				log.Printf("Error: %v", err)
		    	resCh <- 0
		    	break
			}
		}
		
		con.Close()
	})(resCh, conbuf)
	
	// send query to server
	_, err = con.Write(msg)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	
	// Wait for server's response
	<- resCh
	
	return
}
