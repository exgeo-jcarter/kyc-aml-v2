package KycAmlClient

import (
	"testing"
	"os"
	"time"
	"math/rand"
	"encoding/json"
	"os/exec"
	"log"
	kyc_aml_data "../../kyc-aml-data/KycAmlData"
	kyc_aml_fuzzy "../../kyc-aml-fuzzy/KycAmlFuzzy"
	kyc_aml_metaphone "../../kyc-aml-metaphone/KycAmlMetaphone"
)

var fuzzy_server *kyc_aml_fuzzy.KycAmlFuzzyS
var metaphone_server *kyc_aml_metaphone.KycAmlMetaphoneS
var client *KycAmlClientS
const alphabet = "qwertyuiopasdfghjklzxcvbnm"

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
	
	
	// Start a new external server.
	cmd := exec.Command("../../kyc-aml-doublemetaphone/build/kyc-aml-doublemetaphone", "../../kyc-aml-doublemetaphone/config.json")
	err = cmd.Start()
	if err != nil {
		log.Printf("Error: %v", err)
		os.Exit(12)
	}
	
	// Listen for connections.
	go (func() {
		err = metaphone_server.Listen()
		if err != nil {
			cmd.Process.Kill()
			os.Exit(6)
		}
	})()
	
	time.Sleep(1 * time.Second)
	
	client, err = NewKycAmlClient("config.json")
	if err != nil {
		cmd.Process.Kill()
		os.Exit(7)
	}
	_ = client
	
	_, err = client.QueryDataServer("load_sdn_list", "")
	if err != nil {
		cmd.Process.Kill()
		os.Exit(8)
	}
	
	sdn_list, err := client.QueryDataServer("get_sdn_list", "")
	if err != nil {
		cmd.Process.Kill()
		os.Exit(9)
	}
	_ = sdn_list
	
	// TODO: Comment this out for way faster startup if you're not using fuzzy.
	fuzzy_train_sdn_res, err := client.QueryFuzzyServer("train_sdn", sdn_list)
	if err != nil {
		os.Exit(10)
	}
	_ = fuzzy_train_sdn_res

	metaphone_train_sdn_res, err := client.QueryMetaphoneServer("train_sdn", sdn_list)
	if err != nil {
		cmd.Process.Kill()
		os.Exit(11)
	}
	_ = metaphone_train_sdn_res
	
	doublemetaphone_train_sdn_res, err := client.QueryDoubleMetaphoneServer("train_sdn", sdn_list)
	if err != nil {
		cmd.Process.Kill()
		os.Exit(11)
	}
	_ = doublemetaphone_train_sdn_res
	
	result := m.Run()
	cmd.Process.Kill()	// Kill double metaphone server.
	os.Exit(result)
}

// --- Tests ---

// Fuzzy search all names from the SDN list with 2 of their characters randomly changed, 
// and record the miss rate.
func TestFuzzyNameQueryTwoCharsChanged(t *testing.T) {
	
	n := 1
	threshold := 1.0
	
	rand_alphabet := 0
	rand_pos := 0
	first_name := ""
	last_name := ""
	num_entries := len(fuzzy_server.SdnList.SdnEntries)
	miss_rates := []float64{}
	failed := false
	failed_miss_rate := 0.0
	
	t.Logf("Num iterations: %v", n)
	
	for i := 0; i < n; i++ {
	
		rand.Seed(time.Now().UnixNano())
		num_misses := 0
	
		for _, sdn_entry := range fuzzy_server.SdnList.SdnEntries {
			
			first_name = sdn_entry.FirstName
			last_name = sdn_entry.LastName
			
			name := first_name+" "+last_name
			
			if (len(name) == 0) {
				continue
			}
			
			if len(name) > 1 {
				// 1
				rand_alphabet = rand.Intn(len(alphabet)-1)
				rand_pos = rand.Intn(len(name)-1) 
				name = name[:rand_pos] + string(alphabet[rand_alphabet]) + name[rand_pos+1:]
				
				// 2
				rand_alphabet = rand.Intn(len(alphabet)-1)
				rand_pos = rand.Intn(len(name)-1) 
				name = name[:rand_pos] + string(alphabet[rand_alphabet]) + name[rand_pos+1:]
			}
			
			//t.Logf("Querying: %v", name)
			
			res, err := client.QueryFuzzyServer("query_name", name)
			if err != nil {
				t.FailNow()
			}
			
			//t.Logf("Result: %v", res)
			
			var res_struct *FuzzyQueryResS
			err = json.Unmarshal([]byte(res), &res_struct)
			if err != nil {
				t.Fatalf("Error: %v", err)
			}
			
			if (len(res_struct.NameResult) == 0) &&
			   (len(res_struct.RevNameResult) == 0) &&
			   (len(res_struct.AkaResult) == 0) &&
			   (len(res_struct.RevAkaResult) == 0) &&
			   (len(res_struct.AddressResult) == 0) &&
			   (len(res_struct.PostalCodeResult) == 0) {
			   	
			   	num_misses++
		    }
		}
		
		miss_rate := float64(num_misses) / float64(num_entries) * 100.0
		
		miss_rates = append(miss_rates, miss_rate)
		
		if miss_rate > threshold {
			failed = true
			failed_miss_rate = miss_rate
		}
	}
	
	avg_miss_rate := 0.0
	for _, val := range miss_rates {
		avg_miss_rate += val
	}
	avg_miss_rate = avg_miss_rate / float64(len(miss_rates))
	
	t.Logf("Miss rates: %+v", miss_rates)
	t.Logf("Average miss rate: %v%%", avg_miss_rate)
	
	if failed {
		t.Fatalf("Test failed: Miss rate above threshold: %v%% > %v%%", failed_miss_rate, threshold)
	}
	
	t.Logf("Test passed: Miss rates below threshold: %v%%", threshold)
}

// Fuzzy search all names from the SDN list with 3 of their characters randomly changed, 
// and record the miss rate.
func TestFuzzyNameQueryThreeCharsChanged(t *testing.T) {
	
	n := 1
	threshold := 80.0
	
	rand_alphabet := 0
	rand_pos := 0
	first_name := ""
	last_name := ""
	num_entries := len(fuzzy_server.SdnList.SdnEntries)
	miss_rates := []float64{}
	failed := false
	failed_miss_rate := 0.0
	
	t.Logf("Num iterations: %v", n)
	
	for i := 0; i < n; i++ {
	
		rand.Seed(time.Now().UnixNano())
		num_misses := 0
	
		for _, sdn_entry := range fuzzy_server.SdnList.SdnEntries {
			
			first_name = sdn_entry.FirstName
			last_name = sdn_entry.LastName
			
			name := first_name+" "+last_name
			
			if (len(name) == 0) {
				continue
			}
			
			if len(name) > 1 {
				// 1
				rand_alphabet = rand.Intn(len(alphabet)-1)
				rand_pos = rand.Intn(len(name)-1) 
				name = name[:rand_pos] + string(alphabet[rand_alphabet]) + name[rand_pos+1:]
				
				// 2
				rand_alphabet = rand.Intn(len(alphabet)-1)
				rand_pos = rand.Intn(len(name)-1) 
				name = name[:rand_pos] + string(alphabet[rand_alphabet]) + name[rand_pos+1:]
				
				// 3
				rand_alphabet = rand.Intn(len(alphabet)-1)
				rand_pos = rand.Intn(len(name)-1) 
				name = name[:rand_pos] + string(alphabet[rand_alphabet]) + name[rand_pos+1:]
			}
			
			//t.Logf("Querying: %v", name)
			
			res, err := client.QueryFuzzyServer("query_name", name)
			if err != nil {
				t.FailNow()
			}
			
			//t.Logf("Result: %v", res)
			
			var res_struct *FuzzyQueryResS
			err = json.Unmarshal([]byte(res), &res_struct)
			if err != nil {
				t.Fatalf("Error: %v", err)
			}
			
			if (len(res_struct.NameResult) == 0) &&
			   (len(res_struct.RevNameResult) == 0) &&
			   (len(res_struct.AkaResult) == 0) &&
			   (len(res_struct.RevAkaResult) == 0) &&
			   (len(res_struct.AddressResult) == 0) &&
			   (len(res_struct.PostalCodeResult) == 0) {
			   	
			   	num_misses++
		    }
		}
		
		miss_rate := float64(num_misses) / float64(num_entries) * 100.0
		
		miss_rates = append(miss_rates, miss_rate)
		
		if miss_rate > threshold {
			failed = true
			failed_miss_rate = miss_rate
		}
	}
	
	avg_miss_rate := 0.0
	for _, val := range miss_rates {
		avg_miss_rate += val
	}
	avg_miss_rate = avg_miss_rate / float64(len(miss_rates))
	
	t.Logf("Miss rates: %+v", miss_rates)
	t.Logf("Average miss rate: %v%%", avg_miss_rate)
	
	if failed {
		t.Fatalf("Test failed: Miss rate above threshold: %v%% > %v%%", failed_miss_rate, threshold)
	}
	
	t.Logf("Test passed: Miss rates below threshold: %v%%", threshold)
}

// Metaphone search all names from the SDN list with 2 of their characters randomly doubled, 
// and record the miss rate.
func TestMetaphoneNameQueryTwoCharsDoubled(t *testing.T) {
	
	n := 3
	threshold := 5.0
	
	rand_pos := 0
	first_name := ""
	last_name := ""
	num_entries := len(metaphone_server.SdnList.SdnEntries)
	miss_rates := []float64{}
	failed := false
	failed_miss_rate := 0.0
	
	t.Logf("Num iterations: %v", n)
	
	for i := 0; i < n; i++ {
	
		rand.Seed(time.Now().UnixNano())
		num_misses := 0
	
		for _, sdn_entry := range metaphone_server.SdnList.SdnEntries {
			
			first_name = sdn_entry.FirstName
			last_name = sdn_entry.LastName
			
			if (len(first_name) == 0) && (len(last_name) == 0) {
				continue
			}
			
			if len(first_name) > 1 {
				rand_pos = rand.Intn(len(first_name)-1) 
				first_name = first_name[:rand_pos] + string(first_name[rand_pos]) +string(first_name[rand_pos]) + first_name[rand_pos+1:]
			}
			
			if len(last_name) > 1 {
				rand_pos = rand.Intn(len(last_name)-1) 
				last_name = last_name[:rand_pos] + string(last_name[rand_pos]) + string(last_name[rand_pos]) + last_name[rand_pos+1:]
			}
			
			//t.Logf("Querying: %v %v", first_name, last_name)
			
			res, err := client.QueryMetaphoneServer("query_name", first_name+" "+last_name)
			if err != nil {
				t.FailNow()
			}
			
			//t.Logf("Result: %v", res)
			
			var res_struct *MetaphoneQueryResS
			err = json.Unmarshal([]byte(res), &res_struct)
			if err != nil {
				t.Fatalf("Error: %v", err)
			}
			
			if (len(res_struct.NameResult) == 0) &&
			   (len(res_struct.RevNameResult) == 0) &&
			   (len(res_struct.AkaResult) == 0) &&
			   (len(res_struct.RevAkaResult) == 0) &&
			   (len(res_struct.AddressResult) == 0) &&
			   (len(res_struct.PostalCodeResult) == 0) {
			   	
			   	num_misses++
		    }
		}
		
		miss_rate := float64(num_misses) / float64(num_entries) * 100.0
		
		miss_rates = append(miss_rates, miss_rate)
		
		if miss_rate > threshold {
			failed = true
			failed_miss_rate = miss_rate
		}
	}
	
	avg_miss_rate := 0.0
	for _, val := range miss_rates {
		avg_miss_rate += val
	}
	avg_miss_rate = avg_miss_rate / float64(len(miss_rates))
	
	t.Logf("Miss rates: %+v", miss_rates)
	t.Logf("Average miss rate: %v%%", avg_miss_rate)
	
	if failed {
		t.Fatalf("Test failed: Miss rate above threshold: %v%% > %v%%", failed_miss_rate, threshold)
	}
	
	t.Logf("Test passed: Miss rates below threshold: %v%%", threshold)
}

// Metaphone search all names from the SDN list with 2 of their characters randomly tripled, 
// and record the miss rate.
func TestMetaphoneNameQueryTwoCharsTripled(t *testing.T) {
	
	n := 3
	threshold := 5.0
	
	rand_pos := 0
	first_name := ""
	last_name := ""
	num_entries := len(metaphone_server.SdnList.SdnEntries)
	miss_rates := []float64{}
	failed := false
	failed_miss_rate := 0.0
	
	t.Logf("Num iterations: %v", n)
	
	for i := 0; i < n; i++ {
	
		rand.Seed(time.Now().UnixNano())
		num_misses := 0
	
		for _, sdn_entry := range metaphone_server.SdnList.SdnEntries {
			
			first_name = sdn_entry.FirstName
			last_name = sdn_entry.LastName
			
			if (len(first_name) == 0) && (len(last_name) == 0) {
				continue
			}
			
			if len(first_name) > 1 {
				rand_pos = rand.Intn(len(first_name)-1) 
				first_name = first_name[:rand_pos] + string(first_name[rand_pos]) + string(first_name[rand_pos]) +string(first_name[rand_pos]) + first_name[rand_pos+1:]
			}
			
			if len(last_name) > 1 {
				rand_pos = rand.Intn(len(last_name)-1) 
				last_name = last_name[:rand_pos] + string(last_name[rand_pos]) + string(last_name[rand_pos]) + string(last_name[rand_pos]) + last_name[rand_pos+1:]
			}
			
			//t.Logf("Querying: %v %v", first_name, last_name)
			
			res, err := client.QueryMetaphoneServer("query_name", first_name+" "+last_name)
			if err != nil {
				t.FailNow()
			}
			
			//t.Logf("Result: %v", res)
			
			var res_struct *MetaphoneQueryResS
			err = json.Unmarshal([]byte(res), &res_struct)
			if err != nil {
				t.Fatalf("Error: %v", err)
			}
			
			if (len(res_struct.NameResult) == 0) &&
			   (len(res_struct.RevNameResult) == 0) &&
			   (len(res_struct.AkaResult) == 0) &&
			   (len(res_struct.RevAkaResult) == 0) &&
			   (len(res_struct.AddressResult) == 0) &&
			   (len(res_struct.PostalCodeResult) == 0) {
			   	
			   	num_misses++
		    }
		}
		
		miss_rate := float64(num_misses) / float64(num_entries) * 100.0
		
		miss_rates = append(miss_rates, miss_rate)
		
		if miss_rate > threshold {
			failed = true
			failed_miss_rate = miss_rate
		}
	}
	
	avg_miss_rate := 0.0
	for _, val := range miss_rates {
		avg_miss_rate += val
	}
	avg_miss_rate = avg_miss_rate / float64(len(miss_rates))
	
	t.Logf("Miss rates: %+v", miss_rates)
	t.Logf("Average miss rate: %v%%", avg_miss_rate)
	
	if failed {
		t.Fatalf("Test failed: Miss rate above threshold: %v%% > %v%%", failed_miss_rate, threshold)
	}
	
	t.Logf("Test passed: Miss rates below threshold: %v%%", threshold)
}

// Metaphone search all names from the SDN list with 2 of their characters randomly changed, 
// and record the miss rate.
func TestMetaphoneNameQueryTwoCharsChanged(t *testing.T) {
	
	n := 3
	threshold := 90.0
	
	rand_alphabet := 0
	rand_pos := 0
	first_name := ""
	last_name := ""
	num_entries := len(metaphone_server.SdnList.SdnEntries)
	miss_rates := []float64{}
	failed := false
	failed_miss_rate := 0.0
	
	t.Logf("Num iterations: %v", n)
	
	for i := 0; i < n; i++ {
	
		rand.Seed(time.Now().UnixNano())
		num_misses := 0
	
		for _, sdn_entry := range metaphone_server.SdnList.SdnEntries {
			
			first_name = sdn_entry.FirstName
			last_name = sdn_entry.LastName
			
			if (len(first_name) == 0) && (len(last_name) == 0) {
				continue
			}
			
			if len(first_name) > 1 {
				rand_alphabet = rand.Intn(len(alphabet)-1)
				rand_pos = rand.Intn(len(first_name)-1) 
				first_name = first_name[:rand_pos] + string(alphabet[rand_alphabet]) + first_name[rand_pos+1:]
			}
			
			if len(last_name) > 1 {
				rand_alphabet = rand.Intn(len(alphabet)-1)
				rand_pos = rand.Intn(len(last_name)-1) 
				last_name = last_name[:rand_pos] + string(alphabet[rand_alphabet]) + last_name[rand_pos+1:]
			}
			
			//t.Logf("Querying: %v %v", first_name, last_name)
			
			res, err := client.QueryMetaphoneServer("query_name", first_name+" "+last_name)
			if err != nil {
				t.FailNow()
			}
			
			//t.Logf("Result: %v", res)
			
			var res_struct *MetaphoneQueryResS
			err = json.Unmarshal([]byte(res), &res_struct)
			if err != nil {
				t.Fatalf("Error: %v", err)
			}
			
			if (len(res_struct.NameResult) == 0) &&
			   (len(res_struct.RevNameResult) == 0) &&
			   (len(res_struct.AkaResult) == 0) &&
			   (len(res_struct.RevAkaResult) == 0) &&
			   (len(res_struct.AddressResult) == 0) &&
			   (len(res_struct.PostalCodeResult) == 0) {
			   	
			   	num_misses++
		    }
		}
		
		miss_rate := float64(num_misses) / float64(num_entries) * 100.0
		
		miss_rates = append(miss_rates, miss_rate)
		
		if miss_rate > threshold {
			failed = true
			failed_miss_rate = miss_rate
		}
	}
	
	avg_miss_rate := 0.0
	for _, val := range miss_rates {
		avg_miss_rate += val
	}
	avg_miss_rate = avg_miss_rate / float64(len(miss_rates))
	
	t.Logf("Miss rates: %+v", miss_rates)
	t.Logf("Average miss rate: %v%%", avg_miss_rate)
	
	if failed {
		t.Fatalf("Test failed: Miss rate above threshold: %v%% > %v%%", failed_miss_rate, threshold)
	}
	
	t.Logf("Test passed: Miss rates below threshold: %v%%", threshold)
}

// DoubleMetaphone search all names from the SDN list with 2 of their characters randomly doubled, 
// and record the miss rate.
func TestDoubleMetaphoneNameQueryTwoCharsDoubled(t *testing.T) {
	
	n := 3
	threshold := 10.0
	
	rand_pos := 0
	first_name := ""
	last_name := ""
	num_entries := len(metaphone_server.SdnList.SdnEntries)
	miss_rates := []float64{}
	failed := false
	failed_miss_rate := 0.0
	
	t.Logf("Num iterations: %v", n)
	
	for i := 0; i < n; i++ {
	
		rand.Seed(time.Now().UnixNano())
		num_misses := 0
	
		for _, sdn_entry := range metaphone_server.SdnList.SdnEntries {
			
			first_name = sdn_entry.FirstName
			last_name = sdn_entry.LastName
			
			if (len(first_name) == 0) && (len(last_name) == 0) {
				continue
			}
			
			if len(first_name) > 1 {
				rand_pos = rand.Intn(len(first_name)-1) 
				first_name = first_name[:rand_pos] + string(first_name[rand_pos]) +string(first_name[rand_pos]) + first_name[rand_pos+1:]
			}
			
			if len(last_name) > 1 {
				rand_pos = rand.Intn(len(last_name)-1) 
				last_name = last_name[:rand_pos] + string(last_name[rand_pos]) + string(last_name[rand_pos]) + last_name[rand_pos+1:]
			}
			
			//t.Logf("Querying: %v %v", first_name, last_name)
			
			res, err := client.QueryDoubleMetaphoneServer("query_name", first_name+" "+last_name)
			if err != nil {
				t.FailNow()
			}
			
			//t.Logf("Result: %v", res)
			
			var res_struct *DoubleMetaphoneQueryResS
			err = json.Unmarshal([]byte(res), &res_struct)
			if err != nil {
				t.Fatalf("Error: %v", err)
			}
			
			if (len(res_struct.NameResult1) == 0) &&
			   (len(res_struct.NameResult2) == 0) &&
			   (len(res_struct.RevNameResult1) == 0) &&
			   (len(res_struct.RevNameResult2) == 0) &&
			   (len(res_struct.AkaResult1) == 0) &&
			   (len(res_struct.AkaResult2) == 0) &&
			   (len(res_struct.RevAkaResult1) == 0) &&
			   (len(res_struct.RevAkaResult2) == 0) &&
			   (len(res_struct.AddressResult1) == 0) &&
			   (len(res_struct.AddressResult2) == 0) &&
			   (len(res_struct.PostalCodeResult1) == 0) &&
			   (len(res_struct.PostalCodeResult2) == 0) {
			   	
			   	num_misses++
		    }
		}
		
		miss_rate := float64(num_misses) / float64(num_entries) * 100.0
		
		miss_rates = append(miss_rates, miss_rate)
		
		if miss_rate > threshold {
			failed = true
			failed_miss_rate = miss_rate
		}
	}
	
	avg_miss_rate := 0.0
	for _, val := range miss_rates {
		avg_miss_rate += val
	}
	avg_miss_rate = avg_miss_rate / float64(len(miss_rates))
	
	t.Logf("Miss rates: %+v", miss_rates)
	t.Logf("Average miss rate: %v%%", avg_miss_rate)
	
	if failed {
		t.Fatalf("Test failed: Miss rate above threshold: %v%% > %v%%", failed_miss_rate, threshold)
	}
	
	t.Logf("Test passed: Miss rates below threshold: %v%%", threshold)
}

func TestDoubleMetaphoneNameQueryOneCharChanged(t *testing.T) {
	
	// Uses metaphone_server's SDN list, so metaphone_server must be trained.
	
	n := 3
	threshold := 80.0
	
	rand_alphabet := 0
	rand_pos := 0
	first_name := ""
	last_name := ""
	num_entries := len(metaphone_server.SdnList.SdnEntries)
	miss_rates := []float64{}
	failed := false
	failed_miss_rate := 0.0
	
	t.Logf("Num iterations: %v", n)
	
	for i := 0; i < n; i++ {
	
		rand.Seed(time.Now().UnixNano())
		num_misses := 0
	
		for idx, sdn_entry := range metaphone_server.SdnList.SdnEntries {
			_ = idx
			
			first_name = sdn_entry.FirstName
			last_name = sdn_entry.LastName
			
			name := first_name+" "+last_name
			
			if (len(name) == 0) {
				continue
			}
			
			if len(name) > 1 {
				// 1
				rand_alphabet = rand.Intn(len(alphabet)-1)
				rand_pos = rand.Intn(len(name)-1) 
				name = name[:rand_pos] + string(alphabet[rand_alphabet]) + name[rand_pos+1:]
			}
			
			//t.Logf("Querying: %v", name)
			
			res, err := client.QueryDoubleMetaphoneServer("query_name", name)
			if err != nil {
				t.FailNow()
			}
			
			//t.Logf("Result: %v", res)
			
			var res_struct *DoubleMetaphoneQueryResS
			err = json.Unmarshal([]byte(res), &res_struct)
			if err != nil {
				t.Fatalf("Error: %v", err)
			}
			
			if (len(res_struct.NameResult1) == 0) &&
			   (len(res_struct.NameResult2) == 0) &&
			   (len(res_struct.RevNameResult1) == 0) &&
			   (len(res_struct.RevNameResult2) == 0) &&
			   (len(res_struct.AkaResult1) == 0) &&
			   (len(res_struct.AkaResult2) == 0) &&
			   (len(res_struct.RevAkaResult1) == 0) &&
			   (len(res_struct.RevAkaResult2) == 0) &&
			   (len(res_struct.AddressResult1) == 0) &&
			   (len(res_struct.AddressResult2) == 0) &&
			   (len(res_struct.PostalCodeResult1) == 0) &&
			   (len(res_struct.PostalCodeResult2) == 0) {
			   	
			   	num_misses++
		    }
		}
		
		miss_rate := float64(num_misses) / float64(num_entries) * 100.0
		
		miss_rates = append(miss_rates, miss_rate)
		
		if miss_rate > threshold {
			failed = true
			failed_miss_rate = miss_rate
		}
	}
	
	avg_miss_rate := 0.0
	for _, val := range miss_rates {
		avg_miss_rate += val
	}
	avg_miss_rate = avg_miss_rate / float64(len(miss_rates))
	
	t.Logf("Miss rates: %+v", miss_rates)
	t.Logf("Average miss rate: %v%%", avg_miss_rate)
	
	if failed {
		t.Fatalf("Test failed: Miss rate above threshold: %v%% > %v%%", failed_miss_rate, threshold)
	}
	
	t.Logf("Test passed: Miss rates below threshold: %v%%", threshold)
}

func TestDoubleMetaphoneNameQueryTwoCharsChanged(t *testing.T) {
	
	// Uses metaphone_server's SDN list, so metaphone_server must be trained.
	
	n := 3
	threshold := 92.0
	
	rand_alphabet := 0
	rand_pos := 0
	first_name := ""
	last_name := ""
	num_entries := len(metaphone_server.SdnList.SdnEntries)
	miss_rates := []float64{}
	failed := false
	failed_miss_rate := 0.0
	
	t.Logf("Num iterations: %v", n)
	
	for i := 0; i < n; i++ {
	
		rand.Seed(time.Now().UnixNano())
		num_misses := 0
	
		for idx, sdn_entry := range metaphone_server.SdnList.SdnEntries {
			_ = idx
			
			first_name = sdn_entry.FirstName
			last_name = sdn_entry.LastName
			
			name := first_name+" "+last_name
			
			if (len(name) == 0) {
				continue
			}
			
			if len(name) > 1 {
				// 1
				rand_alphabet = rand.Intn(len(alphabet)-1)
				rand_pos = rand.Intn(len(name)-1) 
				name = name[:rand_pos] + string(alphabet[rand_alphabet]) + name[rand_pos+1:]
				
				// 2
				rand_alphabet = rand.Intn(len(alphabet)-1)
				rand_pos = rand.Intn(len(name)-1) 
				name = name[:rand_pos] + string(alphabet[rand_alphabet]) + name[rand_pos+1:]
			}
			
			//t.Logf("Querying: %v", name)
			
			res, err := client.QueryDoubleMetaphoneServer("query_name", name)
			if err != nil {
				t.FailNow()
			}
			
			//t.Logf("Result: %v", res)
			
			var res_struct *DoubleMetaphoneQueryResS
			err = json.Unmarshal([]byte(res), &res_struct)
			if err != nil {
				t.Fatalf("Error: %v", err)
			}
			
			if (len(res_struct.NameResult1) == 0) &&
			   (len(res_struct.NameResult2) == 0) &&
			   (len(res_struct.RevNameResult1) == 0) &&
			   (len(res_struct.RevNameResult2) == 0) &&
			   (len(res_struct.AkaResult1) == 0) &&
			   (len(res_struct.AkaResult2) == 0) &&
			   (len(res_struct.RevAkaResult1) == 0) &&
			   (len(res_struct.RevAkaResult2) == 0) &&
			   (len(res_struct.AddressResult1) == 0) &&
			   (len(res_struct.AddressResult2) == 0) &&
			   (len(res_struct.PostalCodeResult1) == 0) &&
			   (len(res_struct.PostalCodeResult2) == 0) {
			   	
			   	num_misses++
		    }
		}
		
		miss_rate := float64(num_misses) / float64(num_entries) * 100.0
		
		miss_rates = append(miss_rates, miss_rate)
		
		if miss_rate > threshold {
			failed = true
			failed_miss_rate = miss_rate
		}
	}
	
	avg_miss_rate := 0.0
	for _, val := range miss_rates {
		avg_miss_rate += val
	}
	avg_miss_rate = avg_miss_rate / float64(len(miss_rates))
	
	t.Logf("Miss rates: %+v", miss_rates)
	t.Logf("Average miss rate: %v%%", avg_miss_rate)
	
	if failed {
		t.Fatalf("Test failed: Miss rate above threshold: %v%% > %v%%", failed_miss_rate, threshold)
	}
	
	t.Logf("Test passed: Miss rates below threshold: %v%%", threshold)
}

// --- End Tests ---


// --- Benchmarks ---

// Fuzzy search a random name from the SDN list with 2 of its characters randomly changed.
func BenchmarkFuzzyNameQuery(b *testing.B) {
	
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
		
		//b.Logf("Querying: %v %v", first_name, last_name)
	
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
		
		//b.Logf("Querying: %v %v", last_name, first_name)
	
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
	
		//b.Logf("Querying: %v %v", first_name, last_name)
		
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
	
		//b.Logf("Querying: %v %v", last_name, first_name)
	
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
		
		//b.Logf("Querying: %v", address1)
	
		_, err := client.QueryFuzzyServer("query_address", address1)
		if err != nil {
			b.FailNow()
		}
	}
}

// Fuzzy search a random postal code from the SDN list with 2 of its characters randomly changed.
func BenchmarkFuzzyPostalCodeQuery(b *testing.B) {
	
	for i := 0; i < b.N; i++ {
		
		rand.Seed(time.Now().UnixNano())
		rand_sdn := 0
		rand_alphabet := 0
		rand_addresses := 0
		rand_postal_code := 0
		postal_code := ""
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
		
		if (len(addresses) > 0) && (len(addresses[rand_addresses].PostalCode) > 1) {
			rand_postal_code = rand.Intn(len(addresses[rand_addresses].PostalCode)-1)
		}
		
		if len(addresses) > 0 {
			
			postal_code = addresses[rand_addresses].PostalCode
		}
		
		if len(postal_code) > 1 {
			rand_postal_code = rand.Intn(len(postal_code)-1)
			rand_alphabet = rand.Intn(len(alphabet)-1)
			postal_code = postal_code[:rand_postal_code] + string(alphabet[rand_alphabet]) + postal_code[rand_postal_code+1:]
		}
		
		//b.Logf("Querying: %v", postal_code)
	
		_, err := client.QueryFuzzyServer("query_address", postal_code)
		if err != nil {
			b.FailNow()
		}
	}
}

// Metaphone search a random name from the SDN list with 2 of its characters randomly changed.
func BenchmarkMetaphoneNameQuery(b *testing.B) {
	
	for i := 0; i < b.N; i++ {
		
		rand.Seed(time.Now().UnixNano())
		rand_num := rand.Intn(len(metaphone_server.SdnList.SdnEntries)-1)
		rand_alphabet := 1
	
		first_name := metaphone_server.SdnList.SdnEntries[rand_num].FirstName
		
		if len(first_name) > 1 {
			rand_first_name := rand.Intn(len(first_name)-1)
			rand_alphabet = rand.Intn(len(alphabet)-1)
			first_name = first_name[:rand_first_name] + string(alphabet[rand_alphabet]) + first_name[rand_first_name+1:]
		}
		
		last_name := metaphone_server.SdnList.SdnEntries[rand_num].LastName
		
		if len(last_name) > 1 {
			rand_last_name := rand.Intn(len(last_name)-1)
			rand_alphabet = rand.Intn(len(alphabet)-1)
			last_name = last_name[:rand_last_name] + string(alphabet[rand_alphabet]) + last_name[rand_last_name+1:]
		}
		
		//b.Logf("Querying: %v %v", first_name, last_name)
	
		res, err := client.QueryMetaphoneServer("query_name", first_name+" "+last_name)
		if err != nil {
			b.FailNow()
		}
		_ = res
		
		//b.Logf("Result: %v", res)
	}
}

// Metaphone search a random reverse name from the SDN list with 2 of its characters randomly changed.
func BenchmarkMetaphoneRevNameQuery(b *testing.B) {
	
	for i := 0; i < b.N; i++ {
		
		rand.Seed(time.Now().UnixNano())
		rand_num := rand.Intn(len(metaphone_server.SdnList.SdnEntries)-1)
		rand_alphabet := 0
	
		first_name := metaphone_server.SdnList.SdnEntries[rand_num].FirstName
		
		if len(first_name) > 1 {
			rand_first_name := rand.Intn(len(first_name)-1)
			rand_alphabet = rand.Intn(len(alphabet)-1)
			first_name = first_name[:rand_first_name] + string(alphabet[rand_alphabet]) + first_name[rand_first_name+1:]
		}
		
		last_name := metaphone_server.SdnList.SdnEntries[rand_num].LastName
		
		if len(last_name) > 1 {
			rand_last_name := rand.Intn(len(last_name)-1)
			rand_alphabet = rand.Intn(len(alphabet)-1)
			last_name = last_name[:rand_last_name] + string(alphabet[rand_alphabet]) + last_name[rand_last_name+1:]
		}
		
		//b.Logf("Querying: %v %v", last_name, first_name)
	
		_, err := client.QueryMetaphoneServer("query_name", last_name+" "+first_name)
		if err != nil {
			b.FailNow()
		}
	}
}

// Metaphone search a random aka from the SDN list with 2 of its characters randomly changed.
func BenchmarkMetaphoneAkaQuery(b *testing.B) {
	
	for i := 0; i < b.N; i++ {
		
		rand.Seed(time.Now().UnixNano())
		rand_num := rand.Intn(len(metaphone_server.SdnList.SdnEntries)-1)
		rand_alphabet := 0
		rand_aka := 0
		first_name := ""
		last_name := ""
	
		sdn_entry := metaphone_server.SdnList.SdnEntries[rand_num]
		
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
	
		//b.Logf("Querying: %v %v", first_name, last_name)
		
		_, err := client.QueryMetaphoneServer("query_name", first_name+" "+last_name)
		if err != nil {
			b.FailNow()
		}
	}
}

// Metaphone search a random reverse aka from the SDN list with 2 of its characters randomly changed.
func BenchmarkMetaphoneRevAkaQuery(b *testing.B) {
	
	for i := 0; i < b.N; i++ {
		
		rand.Seed(time.Now().UnixNano())
		rand_num := rand.Intn(len(metaphone_server.SdnList.SdnEntries)-1)
		rand_alphabet := 0
		rand_aka := 0
		first_name := ""
		last_name := ""
	
		sdn_entry := metaphone_server.SdnList.SdnEntries[rand_num]
		
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
	
		//b.Logf("Querying: %v %v", last_name, first_name)
	
		_, err := client.QueryMetaphoneServer("query_name", last_name+" "+first_name)
		if err != nil {
			b.FailNow()
		}
	}
}

// Metaphone search a random address from the SDN list with 2 of its characters randomly changed.
func BenchmarkMetaphoneAddressQuery(b *testing.B) {
	
	for i := 0; i < b.N; i++ {
		
		rand.Seed(time.Now().UnixNano())
		rand_sdn := 0
		rand_alphabet := 0
		rand_addresses := 0
		rand_address := 0
		address1 := ""
		addresses := []kyc_aml_metaphone.AddressS{}
		
		if len(metaphone_server.SdnList.SdnEntries) > 1 {
			rand_sdn = rand.Intn(len(metaphone_server.SdnList.SdnEntries)-1)
		}
		
		if len(metaphone_server.SdnList.SdnEntries) > 0 {
			addresses = metaphone_server.SdnList.SdnEntries[rand_sdn].AddressList.Addresses
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
		
		//b.Logf("Querying: %v", address1)
	
		_, err := client.QueryMetaphoneServer("query_address", address1)
		if err != nil {
			b.FailNow()
		}
	}
}

// Metaphone search a random postal code from the SDN list with 2 of its characters randomly changed.
func BenchmarkMetaphonePostalCodeQuery(b *testing.B) {
	
	for i := 0; i < b.N; i++ {
		
		rand.Seed(time.Now().UnixNano())
		rand_sdn := 0
		rand_alphabet := 0
		rand_addresses := 0
		rand_postal_code := 0
		postal_code := ""
		addresses := []kyc_aml_metaphone.AddressS{}
		
		if len(metaphone_server.SdnList.SdnEntries) > 1 {
			rand_sdn = rand.Intn(len(metaphone_server.SdnList.SdnEntries)-1)
		}
		
		if len(metaphone_server.SdnList.SdnEntries) > 0 {
			addresses = metaphone_server.SdnList.SdnEntries[rand_sdn].AddressList.Addresses
		}
		
		if len(addresses) > 1 {
			rand_addresses = rand.Intn(len(addresses)-1)
		}
		
		if (len(addresses) > 0) && (len(addresses[rand_addresses].PostalCode) > 1) {
			rand_postal_code = rand.Intn(len(addresses[rand_addresses].PostalCode)-1)
		}
		
		if len(addresses) > 0 {
			
			postal_code = addresses[rand_addresses].PostalCode
		}
		
		if len(postal_code) > 1 {
			rand_postal_code = rand.Intn(len(postal_code)-1)
			rand_alphabet = rand.Intn(len(alphabet)-1)
			postal_code = postal_code[:rand_postal_code] + string(alphabet[rand_alphabet]) + postal_code[rand_postal_code+1:]
		}
		
		//b.Logf("Querying: %v", postal_code)
	
		_, err := client.QueryMetaphoneServer("query_address", postal_code)
		if err != nil {
			b.FailNow()
		}
	}
}

// --- End Benchmarks ---
