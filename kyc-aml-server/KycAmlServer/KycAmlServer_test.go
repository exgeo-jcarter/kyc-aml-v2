package KycAmlServer

import (
	"testing"
	"os"
	kyc_aml_client "../../kyc-aml-client/KycAmlClient"
	"encoding/json"
	"math/rand"
	"time"
)

var server *kycAmlServerS

func TestMain(m *testing.M) { 
	
	// Start a new server.
	var err error
	server, err = NewKycAmlServer("config.json")
	if err != nil {
		os.Exit(-1)
	}
	
	// Listen for connections.
	go (func() {
		err = server.Listen()
		if err != nil {
			os.Exit(-1)
		}
	})()

	result := m.Run()
	
	os.Exit(result)
}

/*
func TestQueryRemoveFirstChar(t *testing.T) {
	
	t.Logf("Building test set: All entries in SDN list with their first character removed.")
	
	items := []string{}
	
	for _, sdn_entry := range server.Data.SdnEntries {
		
		name := sdn_entry.FirstName+" "+sdn_entry.LastName
		
		if len(name) > 0 {
			items = append(items, name[1:])
		}
		
		for _, aka := range sdn_entry.AkaList.Akas {
			
			aka_name := aka.FirstName+" "+aka.LastName
			
			if len(aka_name) > 0 {
				items = append(items, aka_name[1:])
			}
		}
		
		for _, address := range sdn_entry.AddressList.Addresses { 
			
			if len(address.Address1) > 0 {
				items = append(items, address.Address1[1:])
			}
		}
	}
	
	// Start a new client.
	client, err := kyc_aml_client.NewKycAmlClient("../../kyc-aml-client/KycAmlClient/config.json")
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	
	t.Logf("Querying...")
	
	num_misses := float64(0)
	var query_res *QueryResS
	
	for _, item := range items {
		res, err := client.Query(item)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
		
		err = json.Unmarshal(res, &query_res)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
		
		if query_res.RiskScore == 0 {
			num_misses++
			//t.Logf("miss: %v", item)
		}
	}
	
	miss_percent := (num_misses / float64(len(items))) * 100.0
	threshold := 30.0
	
	t.Logf("Querying complete. Num misses: %v / %v : %.2f%%", num_misses, len(items), miss_percent)
	if (miss_percent >= threshold) {
		t.Fatalf("Failed due to too many misses.")
	}
	
	t.Logf("Test passed because misses were below threshold of %v%%", threshold)
}

func TestQueryRemoveLastChar(t *testing.T) {
	
	t.Logf("Building test set: All entries in SDN list with their last character removed.")
	
	items := []string{}
	
	for _, sdn_entry := range server.Data.SdnEntries {
		
		name := sdn_entry.FirstName+" "+sdn_entry.LastName
		
		if len(name) > 0 {
			items = append(items, name[:len(name)-1])
		}
		
		for _, aka := range sdn_entry.AkaList.Akas {
			
			aka_name := aka.FirstName+" "+aka.LastName
			
			if len(aka_name) > 0 {
				items = append(items, aka_name[:len(aka_name)-1])
			}
		}
		
		for _, address := range sdn_entry.AddressList.Addresses { 
			
			if len(address.Address1) > 0 {
				items = append(items, address.Address1[:len(address.Address1)-1])
			}
		}
	}
	
	// Start a new client.
	client, err := kyc_aml_client.NewKycAmlClient("../../kyc-aml-client/KycAmlClient/config.json")
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	
	t.Logf("Querying...")
	num_misses := float64(0)
	var query_res *QueryResS
	
	for _, item := range items {
		
		res, err := client.Query(item)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
		
		err = json.Unmarshal(res, &query_res)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
		
		if query_res.RiskScore == 0 {
			num_misses++
			//t.Logf("miss: %v", item)
		}
	}
	
	miss_percent := (num_misses / float64(len(items))) * 100.0
	threshold := 30.0
	
	t.Logf("Querying complete. Num misses: %v / %v : %.2f%%", num_misses, len(items), miss_percent)
	if (miss_percent >= threshold) {
		t.Fatalf("Failed due to too many misses.")
	}
	
	t.Logf("Test passed because misses were below threshold of %v%%", threshold)
}

// Remove a random character. Runs n times.
func TestQueryRemoveRandomChar(t *testing.T) {
	
	n := 1
	threshold := 30.0
	failed := false
	results := []float64{}
	
	t.Logf("All entries in SDN list with a random character removed. Querying %v times...", n)
	
	for i := 0; i < n; i++ {
		
		items := []string{}
		
		rand.Seed(time.Now().UnixNano())
		
		for _, sdn_entry := range server.Data.SdnEntries {
			
			name := sdn_entry.FirstName+" "+sdn_entry.LastName
			
			if len(name) > 0 {
				if len(name) > 1 {
					rand1 := rand.Intn(len(name)-1)
					name = name[:rand1] + name[rand1+1:]
				}
				items = append(items, name)
			}
			
			for _, aka := range sdn_entry.AkaList.Akas {
				
				aka_name := aka.FirstName+" "+aka.LastName
				
				if len(aka_name) > 0 {
					if len(aka_name) > 1 {
						rand2 := rand.Intn(len(aka_name)-1)
						aka_name = aka_name[:rand2] + aka_name[rand2+1:]
					}
					items = append(items, aka_name)
				}
			}
			
			for _, address := range sdn_entry.AddressList.Addresses { 
				
				if len(address.Address1) > 0 {
					if len(address.Address1) > 1 {
						rand3 := rand.Intn(len(address.Address1)-1)
						address.Address1 = address.Address1[:rand3] + address.Address1[rand3+1:]
					}
					items = append(items, address.Address1)
				}
			}
		}
		
		// Start a new client.
		client, err := kyc_aml_client.NewKycAmlClient("../../kyc-aml-client/KycAmlClient/config.json")
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
		
		num_misses := float64(0)
		var query_res *QueryResS
		
		for _, item := range items {
			
			res, err := client.Query(item)
			if err != nil {
				t.Fatalf("Error: %v", err)
			}
			
			err = json.Unmarshal(res, &query_res)
			if err != nil {
				t.Fatalf("Error: %v", err)
			}
			
			if query_res.RiskScore == 0 {
				num_misses++
				//t.Logf("miss: %v", item)
			}
		}
		
		miss_percent := (num_misses / float64(len(items))) * 100.0
		results = append(results, miss_percent)
		
		if (miss_percent >= threshold) {
			failed = true
		}
	}
	
	t.Logf("Querying finished. Miss results(%%): %v", results)
	
	if (failed) {
		t.Fatalf("Failed due to too many misses.")
	}
	
	t.Logf("Test passed because misses were below threshold of %v%%", threshold)
}

// Remove a random character, but not the first. Runs n times.
func TestQueryRemoveRandomCharNotFirst(t *testing.T) {
	
	n := 1
	threshold := 30.0
	failed := false
	results := []float64{}
	
	t.Logf("All entries in SDN list with a random character removed, but not the first. Querying %v times...", n)
	
	for i := 0; i < n; i++ {
		
		items := []string{}
		
		rand.Seed(time.Now().UnixNano())
		
		for _, sdn_entry := range server.Data.SdnEntries {
			
			name := sdn_entry.FirstName+" "+sdn_entry.LastName
			
			if len(name) > 0 {
				if len(name) > 1 {
					rand1 := rand.Intn(len(name)-1)
					
					if rand1 == 0 {
						rand1++ 
					}
					
					name = name[:rand1] + name[rand1+1:]
				}
				items = append(items, name)
			}
			
			for _, aka := range sdn_entry.AkaList.Akas {
				
				aka_name := aka.FirstName+" "+aka.LastName
				
				if len(aka_name) > 0 {
					if len(aka_name) > 1 {
						rand2 := rand.Intn(len(aka_name)-1)
						
						if rand2 == 0 {
							rand2++ 
						}
						
						aka_name = aka_name[:rand2] + aka_name[rand2+1:]
					}
					items = append(items, aka_name)
				}
			}
			
			for _, address := range sdn_entry.AddressList.Addresses { 
				
				if len(address.Address1) > 0 {
					if len(address.Address1) > 1 {
						rand3 := rand.Intn(len(address.Address1)-1)
						
						if rand3 == 0 {
							rand3++ 
						}
						
						address.Address1 = address.Address1[:rand3] + address.Address1[rand3+1:]
					}
					items = append(items, address.Address1)
				}
			}
		}
		
		// Start a new client.
		client, err := kyc_aml_client.NewKycAmlClient("../../kyc-aml-client/KycAmlClient/config.json")
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
		
		num_misses := float64(0)
		var query_res *QueryResS
		
		for _, item := range items {
			
			res, err := client.Query(item)
			if err != nil {
				t.Fatalf("Error: %v", err)
			}
			
			err = json.Unmarshal(res, &query_res)
			if err != nil {
				t.Fatalf("Error: %v", err)
			}
			
			if query_res.RiskScore == 0 {
				num_misses++
				//t.Logf("miss: %v", item)
			}
		}
		
		miss_percent := (num_misses / float64(len(items))) * 100.0
		results = append(results, miss_percent)
		
		if (miss_percent >= threshold) {
			failed = true
		}
	}
	
	t.Logf("Querying finished. Miss results(%%): %v", results)
	
	if (failed) {
		t.Fatalf("Failed due to too many misses.")
	}
	
	t.Logf("Test passed because misses were below threshold of %v%%", threshold)
}

// Change a random character. Runs n times.
func TestQueryChangeRandomChar(t *testing.T) {
	
	n := 1
	threshold := 30.0
	failed := false
	results := []float64{}
	new_chars := "QWERTYUIOPASDFGHJKLZXCVBNMmnbvcxzlkjhgfdsapoiuytrewqMNBVCXZLKJHGFDSAPOIUYTREWQqwertyuiopasdfghjklzxcvbnmlkjhgfdsamnbvcxzpoiuytrewqasdfghjklzxcvbnmqwertyuiop"
	
	t.Logf("All entries in SDN list with a random character changed. Querying %v times...", n)
	
	for i := 0; i < n; i++ {
		
		items := []string{}
		
		rand.Seed(time.Now().UnixNano())
		
		rand4 := rand.Intn(len(new_chars)-2)
		
		for _, sdn_entry := range server.Data.SdnEntries {
			
			name := sdn_entry.FirstName+" "+sdn_entry.LastName
			
			if len(name) > 0 {
				if len(name) > 1 {
					rand1 := rand.Intn(len(name)-1)
					name = name[:rand1] + string(new_chars[rand4:rand4+1]) + name[rand1+1:]
				}
				items = append(items, name)
			}
			
			for _, aka := range sdn_entry.AkaList.Akas {
				
				aka_name := aka.FirstName+" "+aka.LastName
				
				if len(aka_name) > 0 {
					if len(aka_name) > 1 {
						rand2 := rand.Intn(len(aka_name)-1)
						aka_name = aka_name[:rand2] + string(new_chars[rand4:rand4+1]) + aka_name[rand2+1:]
					}
					items = append(items, aka_name)
				}
			}
			
			for _, address := range sdn_entry.AddressList.Addresses { 
				
				if len(address.Address1) > 0 {
					if len(address.Address1) > 1 {
						rand3 := rand.Intn(len(address.Address1)-1)
						address.Address1 = address.Address1[:rand3] + string(new_chars[rand4:rand4+1]) + address.Address1[rand3+1:]
					}
					items = append(items, address.Address1)
				}
			}
		}
		
		// Start a new client.
		client, err := kyc_aml_client.NewKycAmlClient("../../kyc-aml-client/KycAmlClient/config.json")
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
		
		num_misses := float64(0)
		var query_res *QueryResS
		
		for _, item := range items {
			
			res, err := client.Query(item)
			if err != nil {
				t.Fatalf("Error: %v", err)
			}
			
			err = json.Unmarshal(res, &query_res)
			if err != nil {
				t.Fatalf("Error: %v", err)
			}
			
			if query_res.RiskScore == 0 {
				num_misses++
				//t.Logf("miss: %v", item)
			}
		}
		
		miss_percent := (num_misses / float64(len(items))) * 100.0
		results = append(results, miss_percent)
		
		if (miss_percent >= threshold) {
			failed = true
		}
	}
	
	t.Logf("Querying finished. Miss results(%%): %v", results)
	
	if (failed) {
		t.Fatalf("Failed due to too many misses.")
	}
	
	t.Logf("Test passed because misses were below threshold of %v%%", threshold)
}

// Change two random characters. Runs n times.
func TestQueryChangeTwoRandomChars(t *testing.T) {
	
	n := 1
	threshold := 30.0
	failed := false
	results := []float64{}
	new_chars := "QWERTYUIOPASDFGHJKLZXCVBNMmnbvcxzlkjhgfdsapoiuytrewqMNBVCXZLKJHGFDSAPOIUYTREWQqwertyuiopasdfghjklzxcvbnmlkjhgfdsamnbvcxzpoiuytrewqasdfghjklzxcvbnmqwertyuiop"
	
	t.Logf("All entries in SDN list with two random characters changed. Querying %v times...", n)
	
	for i := 0; i < n; i++ {
		
		items := []string{}
		
		rand.Seed(time.Now().UnixNano())
		
		rand4 := rand.Intn(len(new_chars)-2)
		rand5 := rand.Intn(len(new_chars)-2)
		
		for _, sdn_entry := range server.Data.SdnEntries {
			
			name := sdn_entry.FirstName+" "+sdn_entry.LastName
			
			if len(name) > 0 {
				if len(name) > 1 {
					rand1 := rand.Intn(len(name)-1)
					rand11 := rand.Intn(len(name)-1)
					name = name[:rand1] + string(new_chars[rand4:rand4+1]) + name[rand1+1:]
					name = name[:rand11] + string(new_chars[rand5:rand5+1]) + name[rand11+1:]
				}
				items = append(items, name)
			}
			
			for _, aka := range sdn_entry.AkaList.Akas {
				
				aka_name := aka.FirstName+" "+aka.LastName
				
				if len(aka_name) > 0 {
					if len(aka_name) > 1 {
						rand2 := rand.Intn(len(aka_name)-1)
						rand22 := rand.Intn(len(aka_name)-1)
						aka_name = aka_name[:rand2] + string(new_chars[rand4:rand4+1]) + aka_name[rand2+1:]
						aka_name = aka_name[:rand22] + string(new_chars[rand5:rand5+1]) + aka_name[rand22+1:]
					}
					items = append(items, aka_name)
				}
			}
			
			for _, address := range sdn_entry.AddressList.Addresses { 
				
				if len(address.Address1) > 0 {
					if len(address.Address1) > 1 {
						rand3 := rand.Intn(len(address.Address1)-1)
						rand33 := rand.Intn(len(address.Address1)-1)
						address.Address1 = address.Address1[:rand3] + string(new_chars[rand4:rand4+1]) + address.Address1[rand3+1:]
						address.Address1 = address.Address1[:rand33] + string(new_chars[rand5:rand5+1]) + address.Address1[rand33+1:]
					}
					items = append(items, address.Address1)
				}
			}
		}
		
		// Start a new client.
		client, err := kyc_aml_client.NewKycAmlClient("../../kyc-aml-client/KycAmlClient/config.json")
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
		
		num_misses := float64(0)
		var query_res *QueryResS
		
		for _, item := range items {
			
			res, err := client.Query(item)
			if err != nil {
				t.Fatalf("Error: %v", err)
			}
			
			err = json.Unmarshal(res, &query_res)
			if err != nil {
				t.Fatalf("Error: %v", err)
			}
			
			if query_res.RiskScore == 0 {
				num_misses++
				//t.Logf("miss: %v", item)
			}
		}
		
		miss_percent := (num_misses / float64(len(items))) * 100.0
		results = append(results, miss_percent)
		
		if (miss_percent >= threshold) {
			failed = true
		}
	}
	
	t.Logf("Querying finished. Miss results(%%): %v", results)
	
	if (failed) {
		t.Fatalf("Failed due to too many misses.")
	}
	
	t.Logf("Test passed because misses were below threshold of %v%%", threshold)
}
*/

// Change three random characters. Runs n times.
func TestQueryChangeThreeRandomChars(t *testing.T) {
	
	n := 1
	threshold := 30.0
	failed := false
	results := []float64{}
	new_chars := "QWERTYUIOPASDFGHJKLZXCVBNMmnbvcxzlkjhgfdsapoiuytrewqMNBVCXZLKJHGFDSAPOIUYTREWQqwertyuiopasdfghjklzxcvbnmlkjhgfdsamnbvcxzpoiuytrewqasdfghjklzxcvbnmqwertyuiop"
	
	t.Logf("All entries in SDN list with three random characters changed. Querying %v times...", n)
	
	for i := 0; i < n; i++ {
		
		items := []string{}
		
		rand.Seed(time.Now().UnixNano())
		
		rand4 := rand.Intn(len(new_chars)-2)
		rand5 := rand.Intn(len(new_chars)-2)
		rand6 := rand.Intn(len(new_chars)-2)
		
		for _, sdn_entry := range server.Data.SdnEntries {
			
			name := sdn_entry.FirstName+" "+sdn_entry.LastName
			
			if len(name) > 0 {
				if len(name) > 1 {
					rand1 := rand.Intn(len(name)-1)
					rand11 := rand.Intn(len(name)-1)
					rand111 := rand.Intn(len(name)-1)
					name = name[:rand1] + string(new_chars[rand4:rand4+1]) + name[rand1+1:]
					name = name[:rand11] + string(new_chars[rand5:rand5+1]) + name[rand11+1:]
					name = name[:rand111] + string(new_chars[rand6:rand6+1]) + name[rand111+1:]
				}
				items = append(items, name)
			}
			
			for _, aka := range sdn_entry.AkaList.Akas {
				
				aka_name := aka.FirstName+" "+aka.LastName
				
				if len(aka_name) > 0 {
					if len(aka_name) > 1 {
						rand2 := rand.Intn(len(aka_name)-1)
						rand22 := rand.Intn(len(aka_name)-1)
						rand222 := rand.Intn(len(aka_name)-1)
						aka_name = aka_name[:rand2] + string(new_chars[rand4:rand4+1]) + aka_name[rand2+1:]
						aka_name = aka_name[:rand22] + string(new_chars[rand5:rand5+1]) + aka_name[rand22+1:]
						aka_name = aka_name[:rand222] + string(new_chars[rand6:rand6+1]) + aka_name[rand222+1:]
					}
					items = append(items, aka_name)
				}
			}
			
			for _, address := range sdn_entry.AddressList.Addresses { 
				
				if len(address.Address1) > 0 {
					if len(address.Address1) > 1 {
						rand3 := rand.Intn(len(address.Address1)-1)
						rand33 := rand.Intn(len(address.Address1)-1)
						rand333 := rand.Intn(len(address.Address1)-1)
						address.Address1 = address.Address1[:rand3] + string(new_chars[rand4:rand4+1]) + address.Address1[rand3+1:]
						address.Address1 = address.Address1[:rand33] + string(new_chars[rand5:rand5+1]) + address.Address1[rand33+1:]
						address.Address1 = address.Address1[:rand333] + string(new_chars[rand6:rand6+1]) + address.Address1[rand333+1:]
					}
					items = append(items, address.Address1)
				}
			}
		}
		
		// Start a new client.
		client, err := kyc_aml_client.NewKycAmlClient("../../kyc-aml-client/KycAmlClient/config.json")
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
		
		num_misses := float64(0)
		var query_res *QueryResS
		
		for _, item := range items {
			
			res, err := client.Query(item)
			if err != nil {
				t.Fatalf("Error: %v", err)
			}
			
			err = json.Unmarshal(res, &query_res)
			if err != nil {
				t.Fatalf("Error: %v", err)
			}
			
			if query_res.RiskScore == 0 {
				num_misses++
				//t.Logf("miss: %v", item)
			}
		}
		
		miss_percent := (num_misses / float64(len(items))) * 100.0
		results = append(results, miss_percent)
		
		if (miss_percent >= threshold) {
			failed = true
		}
	}
	
	t.Logf("Querying finished. Miss results(%%): %v", results)
	
	if (failed) {
		t.Fatalf("Failed due to too many misses.")
	}
	
	t.Logf("Test passed because misses were below threshold of %v%%", threshold)
}
