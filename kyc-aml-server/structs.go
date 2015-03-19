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

// Server Request
type SocketMsgS struct {
	
	Action 	string	`json:"action,omitempty"`
	Value 	string	`json:"value,omitempty"`
}

// Query response
type QueryResS struct {
	
	Query			string		`json:"query,omitempty"`
	MetaphoneQuery	string		`json:"metaphone_query"`
	Result 			[]string	`json:"result,omitempty"`
	MetaphoneResult	[]string	`json:"metaphone_result,omitempty"`
}
