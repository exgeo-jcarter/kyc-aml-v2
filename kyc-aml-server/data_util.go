package main

import (
	/*
    "encoding/xml"
    "log"
    "net/http"
    "strings"
    "encoding/json"
    */
)

/*
var sdn SdnListS

func LoadDataset() (res []byte, err error) {
	
	client := http.Client{}
	
	log.Printf("Retrieving dataset. Please wait before performing any queries...")
	
	sdn_res, err := client.Get("http://www.treasury.gov/ofac/downloads/sdn.xml")
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	
	log.Printf("Parsing dataset")
	
	d := xml.NewDecoder(sdn_res.Body)
	err = d.Decode(&sdn)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	
	log.Printf("Dataset loaded. You can perform queries now.")
	
	res = []byte(`{"msg": "Dataset loaded"}`)
	
	return
}


/* Query the SDN list. q = name query, aq = address query, pq = postal code query
		
		A name query can be First, Last, or Full name.
		
		An address query should be only street number and street name 
		(you must omit City/Town, Provice or State, and Country).	
*//*
func Query(q, aq, pq string) (res []byte, err error) {
	
	// ----- Name query -----
	queryRes := QueryResS{}
	q_lower := strings.ToLower(q)
	
	log.Printf("Querying SDN list for: %v %v %v", q, aq, pq)
	
	for _, val := range sdn.SdnEntries {
		
		if q != "" {
		
			lastname_lower := strings.ToLower(val.LastName)
			firstname_lower := strings.ToLower(val.FirstName)
			
			// check for last name
			if lastname_lower == q_lower {
				log.Printf("Last name match found")
				queryRes.Num_lastname_matches++
				queryRes.Num_matches++
			}
			
			// check for first names
			if firstname_lower == q_lower {
				log.Printf("First name match found")
				queryRes.Num_firstname_matches++
				queryRes.Num_matches++
			}
			
			// check full name
			if (firstname_lower + " " + lastname_lower) == q_lower {
				log.Printf("Full name match found")
				queryRes.Num_fullname_matches++
				queryRes.Num_matches++
			}
			
			// check reverse full name
			if (lastname_lower + " " + firstname_lower) == q_lower {
				log.Printf("Reverse full name match found")
				queryRes.Num_reversefullname_matches++
				queryRes.Num_matches++
			}
			
			// check aka lists
			for _, val2 := range val.AkaList.Akas {
				
				aka_lastname_lower := strings.ToLower(val2.LastName)
				aka_firstname_lower := strings.ToLower(val2.FirstName)
				
				if aka_lastname_lower == q_lower {
					
					if val2.Category == "strong" {
						log.Printf("AKA strong last name match found")
						queryRes.Num_aka_strong_lastname_matches++
						queryRes.Num_matches++
						
					} else if val2.Category == "weak" {
						log.Printf("AKA weak last name match found")
						queryRes.Num_aka_weak_lastname_matches++
						queryRes.Num_matches++
					}
				}
				
				if aka_firstname_lower == q_lower {
					
					if val2.Category == "strong" {
						log.Printf("AKA strong first name match found")
						queryRes.Num_aka_strong_firstname_matches++
						queryRes.Num_matches++
						
					} else if val2.Category == "weak" {
						log.Printf("AKA weak first name match found")
						queryRes.Num_aka_weak_firstname_matches++
						queryRes.Num_matches++
					}
				}
				
				if (aka_firstname_lower + " " + aka_lastname_lower) == q_lower {
					
					if val2.Category == "strong" {
						log.Printf("AKA strong full name match found")
						queryRes.Num_aka_strong_fullname_matches++
						queryRes.Num_matches++
						
					} else if val2.Category == "weak" {
						log.Printf("AKA weak full name match found")
						queryRes.Num_aka_weak_fullname_matches++
						queryRes.Num_matches++
					}
				}
				
				if (aka_lastname_lower + " " + aka_firstname_lower) == q_lower {
					
					if val2.Category == "strong" {
						log.Printf("AKA strong full name match found")
						queryRes.Num_aka_strong_reversefullname_matches++
						queryRes.Num_matches++
						
					} else if val2.Category == "weak" {
						log.Printf("AKA weak full name match found")
						queryRes.Num_aka_weak_reversefullname_matches++
						queryRes.Num_matches++
					}
				}
			}
		}
		
		// ----- Address query -----
		if aq != "" || pq != "" {
			
			for _, val2 := range val.AddressList.Addresses {
				
				if aq != "" {
				
					aq_lower := strings.ToLower(aq)
					address1_lower := strings.ToLower(val2.Address1)
					
					if (address1_lower == aq_lower) {
						
						log.Printf("Address1 match found")
						queryRes.Num_address1_matches++
						queryRes.Num_matches++
					}
				}
				
				if pq != "" {
					
					pq_lower := strings.ToLower(pq)
					postalcode_lower := strings.ToLower(val2.PostalCode)
					
					if (postalcode_lower == pq_lower) {
						
						log.Printf("Postal code match found")
						queryRes.Num_postalcode_matches++
						queryRes.Num_matches++
					}
				}
			}
		}
		// ----- End Address query -----
		
	}
	// ----- End Name query -----
	
	if queryRes.Num_matches > 0 {
		
		RiskScore(&queryRes)
		
		log.Printf("Query finished. Num matches: %v", queryRes.Num_matches)
	} else {
		log.Printf("Query finished. No matches.")
	}
	
	res, err = json.Marshal(queryRes)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	
	return
}

func RiskScore(query_result *QueryResS) {
	
	if query_result.Num_lastname_matches > 0 {
		query_result.Risk_score += 3
	}
	
	if query_result.Num_firstname_matches > 0 {
		query_result.Risk_score += 3
	}
	
	if query_result.Num_fullname_matches > 0 {
		query_result.Risk_score += 3
	}
	
	if query_result.Num_reversefullname_matches > 0 {
		query_result.Risk_score += 3
	}
	
	if query_result.Num_aka_strong_lastname_matches > 0 {
		query_result.Risk_score += 2
	}
	
	if query_result.Num_aka_weak_lastname_matches > 0 {
		query_result.Risk_score++
	}
	
	if query_result.Num_aka_strong_firstname_matches > 0 {
		query_result.Risk_score += 2
	}
	
	if query_result.Num_aka_weak_firstname_matches > 0 {
		query_result.Risk_score++
	}
	
	if query_result.Num_aka_strong_fullname_matches > 0 {
		query_result.Risk_score += 2
	}
	
	if query_result.Num_aka_weak_fullname_matches > 0 {
		query_result.Risk_score++
	}
	
	if query_result.Num_aka_strong_reversefullname_matches > 0 {
		query_result.Risk_score += 2
	}
	
	if query_result.Num_aka_weak_reversefullname_matches > 0 {
		query_result.Risk_score++
	}
	
	if query_result.Num_address1_matches > 0 {
		query_result.Risk_score += 3
	}
	
	if query_result.Num_postalcode_matches > 0 {
		query_result.Risk_score += 3
	}
	
	query_result.Risk_score = query_result.Risk_score / 14 * 100
}
*/
