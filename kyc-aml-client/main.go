package main

import (
	kyc_aml_client "./KycAmlClient"
	"os"
	"fmt"
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
	
	num_query_servers := 3
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
	
	go (func() {
		doublemetaphone_train_sdn_res, err := client.QueryDoubleMetaphoneServer("train_sdn", sdn_list)
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
		
		go (func() {
			doublemetaphone_name_res, err := client.QueryDoubleMetaphoneServer("query_name", os.Args[1])
			if err != nil {
				return
			}
			if doublemetaphone_name_res != "{}" {
				fmt.Printf("DoubleMetaphone name results: %s\n", doublemetaphone_name_res)
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
		
		go (func() {
			doublemetaphone_address_res, err := client.QueryDoubleMetaphoneServer("query_address", os.Args[2])
			if err != nil {
				return
			}
			if doublemetaphone_address_res != "{}" {
				fmt.Printf("DoubleMetaphone address results: %s\n", doublemetaphone_address_res)
			}
			wait_for_queries_ch <- 1
		})()
	}
	
	for i := 0; i < num_queries; i++ {
		<- wait_for_queries_ch
	}
}
