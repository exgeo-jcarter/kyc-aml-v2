package main

import (
	kyc_aml_client "./KycAmlClient"
	"os"
	"fmt"
//	"time"
)

func main() {
	
	if len(os.Args) < 2 {
		fmt.Printf("\n"+`usage: kyc-aml-client "a name query goes here" ["address or postal code query"]`+"\n\n")
		return
	}
	
	client, err := kyc_aml_client.NewKycAmlClient("KycAmlClient/config.json")
	if err != nil {
		return
	}
	
	_, err = client.QueryDataServer("load_sdn_list", "")
	if err != nil {
		return
	}
	
	sdn_list, err := client.QueryDataServer("get_sdn_list", "")
	if err != nil {
		return
	}
	
	num_query_servers := 2
	wait_for_training_ch := make(chan int)
	
	go (func() {
		fuzzy_train_sdn_res, err := client.QueryFuzzyServer("train_sdn", sdn_list)
		if err != nil {
			return
		}
		_ = fuzzy_train_sdn_res
		wait_for_training_ch <- 1
	})()
	
	go (func() {
		metaphone_train_sdn_res, err := client.QueryMetaphoneServer("train_sdn", sdn_list)
		if err != nil {
			return
		}
		_ = metaphone_train_sdn_res
		wait_for_training_ch <- 1
	})()
	
	/*
	go (func(wait_for_training_ch chan int) {
		
		for ; num_query_servers_trained < num_query_servers; {
			time.Sleep(5 * time.Millisecond)
		}
		
		wait_for_training_ch <- 1
		
	})(wait_for_training_ch)
	*/
	
	for i := 0; i < num_query_servers; i++ {
		<- wait_for_training_ch
	}
	
	num_queries := 0
	wait_for_queries_ch := make(chan int)
	
	if os.Args[1] != "" {
		
		num_queries += num_query_servers
		
		go (func() {
			fuzzy_name_res, err := client.QueryFuzzyServer("query_name", os.Args[1])
			if err != nil {
				return
			}
			if fuzzy_name_res != "{}" {
				fmt.Printf("Fuzzy name results: %s\n", fuzzy_name_res)
			}
			wait_for_queries_ch <- 1
		})()
		
		go (func() {
			metaphone_name_res, err := client.QueryMetaphoneServer("query_name", os.Args[1])
			if err != nil {
				return
			}
			if metaphone_name_res != "{}" {
				fmt.Printf("Metaphone name results: %s\n", metaphone_name_res)
			}
			wait_for_queries_ch <- 1
		})()
	}
	
	if len(os.Args) > 2 {
		
		num_queries += num_query_servers
		
		go (func() {
			fuzzy_address_res, err := client.QueryFuzzyServer("query_address", os.Args[2])
			if err != nil {
				return
			}
			if fuzzy_address_res != "{}" {
				fmt.Printf("Fuzzy address results: %s\n", fuzzy_address_res)
			}
			wait_for_queries_ch <- 1
		})()
		
		go (func() {
			metaphone_address_res, err := client.QueryMetaphoneServer("query_address", os.Args[2])
			if err != nil {
				return
			}
			if metaphone_address_res != "{}" {
				fmt.Printf("Metaphone address results: %s\n", metaphone_address_res)
			}
			wait_for_queries_ch <- 1
		})()
	}
	
	/*
	go (func(wait_for_queries_ch chan int) {
			
		for ; num_query_responses < num_queries; {
			time.Sleep(5 * time.Millisecond)
		}
		
		wait_for_queries_ch <- 1
		
	})(wait_for_queries_ch)
	*/
	
	for i := 0; i < num_queries; i++ {
		<- wait_for_queries_ch
	}
}
