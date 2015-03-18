package main

// ----- XML structs -----
type SdnListS struct {
	
	PublshInformation 	PublishInformationS	`xml:"publshInformation"`
	SdnEntries 			[]*SdnEntryS		`xml:"sdnEntry"`
}

type PublishInformationS struct {
	
	Publish_Date string
	Record_Count int64
}

type SdnEntryS struct {
	
	Uid 		int64			`xml:"uid"`
	FirstName	string			`xml:"firstName"`
	LastName 	string			`xml:"lastName"`
	SdnType 	string			`xml:"sdnType"`
	ProgramList ProgramS		`xml:"programList"`
	AkaList		AkaListS		`xml:"akaList"`
	AddressList AddressListS	`xml:"addressList"`
}

type ProgramS struct {
	
	Programs []string	`xml:"program"`
}

type AkaListS struct {
	
	Akas	[]AkaS	`xml:"aka"`
}

type AkaS struct {
	
	Uid 		int64	`xml:"uid"`
	Type 		string	`xml:"type"`
	Category 	string	`xml:"category"`
	LastName 	string	`xml:"lastName"`
	FirstName 	string	`xml:"firstName"`
}

type AddressListS struct {
	
	Addresses []AddressS	`xml:"address"`
}

type AddressS struct {
	
	Uid 		int64		`xml:"uid"`
	Address1	string		`xml:"address1"`
	Address2	string		`xml:"address2"`
	City 		string		`xml:"city"`
	StateOrProvince	string	`xml:"stateOrProvince"`
	Country 	string		`xml:"country"`
	PostalCode 	string		`xml:"postalCode"`
}
// -----End XML structs -----

type SocketMsgS struct {
	
	Action 	string	`json:"action,omitempty"`
	Value 	string	`json:"value,omitempty"`
}

/*
type QueryResS struct {
	
	Num_matches 							int64	`json:"num_matches,omitempty"`
	
	// Name matches
	Num_lastname_matches 					int64	`json:"num_lastname_matches,omitempty"`
	Num_firstname_matches 					int64	`json:"num_firstname_matches,omitempty"`
	Num_fullname_matches 					int64	`json:"num_fullname_matches,omitempty"`
	Num_reversefullname_matches 			int64	`json:"num_reversefullname_matches,omitempty"`
	Num_aka_strong_lastname_matches 		int64	`json:"num_aka_strong_lastname_matches,omitempty"`
	Num_aka_weak_lastname_matches 			int64	`json:"num_aka_weak_lastname_matches,omitempty"`
	Num_aka_strong_firstname_matches 		int64	`json:"num_aka_strong_firstname_matches,omitempty"`
	Num_aka_weak_firstname_matches 			int64	`json:"num_aka_weak_firstname_matches,omitempty"`
	Num_aka_strong_fullname_matches 		int64	`json:"num_aka_strong_fullname_matches,omitempty"`
	Num_aka_weak_fullname_matches 			int64	`json:"num_aka_weak_fullname_matches,omitempty"`
	Num_aka_strong_reversefullname_matches 	int64	`json:"num_aka_strong_reversefullname_matches,omitempty"`
	Num_aka_weak_reversefullname_matches 	int64	`json:"num_aka_weak_reversefullname_matches,omitempty"`
	
	// Address matches
	Num_address1_matches 					int64	`json:"num_address1_matches,omitempty"`
	Num_postalcode_matches 					int64	`json:"num_postalcode_matches,omitempty"`
	
	Risk_score								float64	`json:"risk_score,omitempty"`
	
	Error 									string	`json:"error,omitempty"`
}
*/
