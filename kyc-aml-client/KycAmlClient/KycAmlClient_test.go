package KycAmlClient

import (
	"testing"
	"os"
	"time"
	kyc_aml_data "../../kyc-aml-data/KycAmlData"
	kyc_aml_fuzzy "../../kyc-aml-fuzzy/KycAmlFuzzy"
	kyc_aml_metaphone "../../kyc-aml-metaphone/KycAmlMetaphone"
	"math/rand"
)

var fuzzy_server *kyc_aml_fuzzy.KycAmlFuzzyS
var metaphone_server *kyc_aml_metaphone.KycAmlMetaphoneS
var client *KycAmlClientS
const alphabet = "qwertyuiopasdfghjklzxcvbnm"
//var num_akas int64

func TestMain(m *testing.M) {
	
	// Start a new server.
	data_server, err := kyc_aml_data.NewKycAmlData("../../kyc-aml-data/KycAmlData/config.json")
	if err != nil {
		os.Exit(1)
	}
	
	// Listen for connections.
	go (func() {
		err = data_server.Listen()
		if err != nil {
			os.Exit(2)
		}
	})()
	
	// Start a new server.
	fuzzy_server, err = kyc_aml_fuzzy.NewKycAmlFuzzy("../../kyc-aml-fuzzy/KycAmlFuzzy/config.json")
	if err != nil {
		os.Exit(3)
	}
	
	// Listen for connections.
	go (func() {
		err = fuzzy_server.Listen()
		if err != nil {
			os.Exit(4)
		}
	})()
	
	// Start a new server.
	metaphone_server, err = kyc_aml_metaphone.NewKycAmlMetaphone("../../kyc-aml-metaphone/KycAmlMetaphone/config.json")
	if err != nil {
		os.Exit(5)
	}
	
	// Listen for connections.
	go (func() {
		err = metaphone_server.Listen()
		if err != nil {
			os.Exit(6)
		}
	})()
	
	time.Sleep(1 * time.Second)
	
	client, err = NewKycAmlClient("config.json")
	if err != nil {
		os.Exit(7)
	}
	_ = client
	
	_, err = client.QueryDataServer("load_sdn_list", "")
	if err != nil {
		os.Exit(8)
	}
	
	sdn_list, err := client.QueryDataServer("get_sdn_list", "")
	if err != nil {
		os.Exit(9)
	}
	_ = sdn_list
	
	fuzzy_train_sdn_res, err := client.QueryFuzzyServer("train_sdn", sdn_list)
	if err != nil {
		os.Exit(10)
	}
	_ = fuzzy_train_sdn_res

	metaphone_train_sdn_res, err := client.QueryMetaphoneServer("train_sdn", sdn_list)
	if err != nil {
		os.Exit(11)
	}
	_ = metaphone_train_sdn_res
	
	/*
	for _, sdn_entry := range fuzzy_server.SdnList.SdnEntries {
		for _, _ = range sdn_entry.AkaList.Akas {
			num_akas++
		}
	}
	*/
	
	os.Exit(m.Run())
}

// Fuzzy search a random name from the SDN list with 2 of its characters randomly changed.
func BenchmarkFuzzyNameQuery(b *testing.B) {
	
	for i := 0; i < b.N; i++ {
		
		rand.Seed(time.Now().UnixNano())
		rand_num := rand.Intn(len(fuzzy_server.SdnList.SdnEntries)-1)
		rand_alphabet := 1
	
		first_name := fuzzy_server.SdnList.SdnEntries[rand_num].FirstName
		
		if len(first_name) > 1 {
			rand_first_name := rand.Intn(len(first_name)-1)
			rand_alphabet = rand.Intn(len(alphabet)-1)
			first_name = first_name[:rand_first_name] + string(alphabet[rand_alphabet]) + first_name[rand_first_name+1:]
		}
		
		last_name := fuzzy_server.SdnList.SdnEntries[rand_num].LastName
		
		if len(last_name) > 1 {
			rand_last_name := rand.Intn(len(last_name)-1)
			rand_alphabet = rand.Intn(len(alphabet)-1)
			last_name = last_name[:rand_last_name] + string(alphabet[rand_alphabet]) + last_name[rand_last_name+1:]
		}
	
		_, err := client.QueryFuzzyServer("query_name", first_name+" "+last_name)
		if err != nil {
			b.FailNow()
		}
	}
}

// Fuzzy search a random reverse name from the SDN list with 2 of its characters randomly changed.
func BenchmarkFuzzyRevNameQuery(b *testing.B) {
	
	for i := 0; i < b.N; i++ {
		
		rand.Seed(time.Now().UnixNano())
		rand_num := rand.Intn(len(fuzzy_server.SdnList.SdnEntries)-1)
		rand_alphabet := 0
	
		first_name := fuzzy_server.SdnList.SdnEntries[rand_num].FirstName
		
		if len(first_name) > 1 {
			rand_first_name := rand.Intn(len(first_name)-1)
			rand_alphabet = rand.Intn(len(alphabet)-1)
			first_name = first_name[:rand_first_name] + string(alphabet[rand_alphabet]) + first_name[rand_first_name+1:]
		}
		
		last_name := fuzzy_server.SdnList.SdnEntries[rand_num].LastName
		
		if len(last_name) > 1 {
			rand_last_name := rand.Intn(len(last_name)-1)
			rand_alphabet = rand.Intn(len(alphabet)-1)
			last_name = last_name[:rand_last_name] + string(alphabet[rand_alphabet]) + last_name[rand_last_name+1:]
		}
	
		_, err := client.QueryFuzzyServer("query_name", last_name+" "+first_name)
		if err != nil {
			b.FailNow()
		}
	}
}

// Fuzzy search a random aka from the SDN list with 2 of its characters randomly changed.
func BenchmarkFuzzyAkaQuery(b *testing.B) {
	
	for i := 0; i < b.N; i++ {
		
		rand.Seed(time.Now().UnixNano())
		rand_num := rand.Intn(len(fuzzy_server.SdnList.SdnEntries)-1)
		rand_alphabet := 0
		rand_aka := 0
		first_name := ""
		last_name := ""
	
		sdn_entry := fuzzy_server.SdnList.SdnEntries[rand_num]
		
		if len(sdn_entry.AkaList.Akas) > 1 {
			rand_aka = rand.Intn(len(sdn_entry.AkaList.Akas)-1)
		}
		
		if len(sdn_entry.AkaList.Akas) > 0 {
			first_name = sdn_entry.AkaList.Akas[rand_aka].FirstName
			last_name = sdn_entry.AkaList.Akas[rand_aka].LastName
		}
		
		if len(first_name) > 1 {
			rand_first_name := rand.Intn(len(first_name)-1)
			rand_alphabet = rand.Intn(len(alphabet)-1)
			first_name = first_name[:rand_first_name] + string(alphabet[rand_alphabet]) + first_name[rand_first_name+1:]
		}
		
		if len(last_name) > 1 {
			rand_last_name := rand.Intn(len(last_name)-1)
			rand_alphabet = rand.Intn(len(alphabet)-1)
			last_name = last_name[:rand_last_name] + string(alphabet[rand_alphabet]) + last_name[rand_last_name+1:]
		}
	
		_, err := client.QueryFuzzyServer("query_name", first_name+" "+last_name)
		if err != nil {
			b.FailNow()
		}
	}
}

// Fuzzy search a random reverse aka from the SDN list with 2 of its characters randomly changed.
func BenchmarkFuzzyRevAkaQuery(b *testing.B) {
	
	for i := 0; i < b.N; i++ {
		
		rand.Seed(time.Now().UnixNano())
		rand_num := rand.Intn(len(fuzzy_server.SdnList.SdnEntries)-1)
		rand_alphabet := 0
		rand_aka := 0
		first_name := ""
		last_name := ""
	
		sdn_entry := fuzzy_server.SdnList.SdnEntries[rand_num]
		
		if len(sdn_entry.AkaList.Akas) > 1 {
			rand_aka = rand.Intn(len(sdn_entry.AkaList.Akas)-1)
		}
		
		if len(sdn_entry.AkaList.Akas) > 0 {
			first_name = sdn_entry.AkaList.Akas[rand_aka].FirstName
			last_name = sdn_entry.AkaList.Akas[rand_aka].LastName
		}
		
		if len(first_name) > 1 {
			rand_first_name := rand.Intn(len(first_name)-1)
			rand_alphabet = rand.Intn(len(alphabet)-1)
			first_name = first_name[:rand_first_name] + string(alphabet[rand_alphabet]) + first_name[rand_first_name+1:]
		}
		
		if len(last_name) > 1 {
			rand_last_name := rand.Intn(len(last_name)-1)
			rand_alphabet = rand.Intn(len(alphabet)-1)
			last_name = last_name[:rand_last_name] + string(alphabet[rand_alphabet]) + last_name[rand_last_name+1:]
		}
	
		_, err := client.QueryFuzzyServer("query_name", last_name+" "+first_name)
		if err != nil {
			b.FailNow()
		}
	}
}

// Fuzzy search a random address from the SDN list with 2 of its characters randomly changed.
func BenchmarkFuzzyAddressQuery(b *testing.B) {
	
	for i := 0; i < b.N; i++ {
		
		rand.Seed(time.Now().UnixNano())
		rand_sdn := 0
		rand_alphabet := 0
		rand_addresses := 0
		rand_address := 0
		address1 := ""
		addresses := []kyc_aml_fuzzy.AddressS{}
		
		if len(fuzzy_server.SdnList.SdnEntries) > 1 {
			rand_sdn = rand.Intn(len(fuzzy_server.SdnList.SdnEntries)-1)
		}
		
		if len(fuzzy_server.SdnList.SdnEntries) > 0 {
			addresses = fuzzy_server.SdnList.SdnEntries[rand_sdn].AddressList.Addresses
		}
		
		if len(addresses) > 1 {
			rand_addresses = rand.Intn(len(addresses)-1)
		}
		
		if (len(addresses) > 0) && (len(addresses[rand_addresses].Address1) > 1) {
			rand_address = rand.Intn(len(addresses[rand_addresses].Address1)-1)
		}
		
		if len(addresses) > 0 {
			
			address1 = addresses[rand_addresses].Address1
		}
		
		if len(address1) > 1 {
			rand_address = rand.Intn(len(address1)-1)
			rand_alphabet = rand.Intn(len(alphabet)-1)
			address1 = address1[:rand_address] + string(alphabet[rand_alphabet]) + address1[rand_address+1:]
		}
	
		_, err := client.QueryFuzzyServer("query_address", address1)
		if err != nil {
			b.FailNow()
		}
	}
}
