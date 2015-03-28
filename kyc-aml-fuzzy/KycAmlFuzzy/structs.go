package KycAmlFuzzy

// ----- XML structs -----
type SdnListS struct {
	
	PublshInformation 	PublishInformationS	`xml:"publshInformation" json:"publsh_information,omitempty"`
	SdnEntries 			[]*SdnEntryS		`xml:"sdnEntry" json:"sdn_entry,omitempty"`
}

type PublishInformationS struct {
	
	Publish_Date string	`json:"publish_date,omitempty"`
	Record_Count int64	`json:"record_count,omitempty"`
}

type SdnEntryS struct {
	
	Uid 		int64			`xml:"uid" json:"uid,omitempty"`
	FirstName	string			`xml:"firstName" json:"first_name,omitempty"`
	LastName 	string			`xml:"lastName" json:"last_name,omitempty"`
	SdnType 	string			`xml:"sdnType" json:"sdn_type,omitempty"`
	ProgramList ProgramS		`xml:"programList" json:"program_list,omitempty"`
	AkaList		AkaListS		`xml:"akaList" json:"aka_list,omitempty"`
	AddressList AddressListS	`xml:"addressList" json:"address_list,omitempty"`
}

type ProgramS struct {
	
	Programs []string	`xml:"program" json:"programs,omitempty"`
}

type AkaListS struct {
	
	Akas	[]AkaS	`xml:"aka" json:"aka,omitempty"`
}

type AkaS struct {
	
	Uid 		int64	`xml:"uid" json:"uid,omitempty"`
	Type 		string	`xml:"type" json:"type,omitempty"`
	Category 	string	`xml:"category" json:"category,omitempty"`
	LastName 	string	`xml:"lastName" json:"last_name,omitempty"`
	FirstName 	string	`xml:"firstName" json:"first_name,omitempty"`
}

type AddressListS struct {
	
	Addresses []AddressS	`xml:"address" json:"addresses,omitempty"`
}

type AddressS struct {
	
	Uid 		int64		`xml:"uid" json:"uid,omitempty"`
	Address1	string		`xml:"address1" json:"address1,omitempty"`
	Address2	string		`xml:"address2" json:"address2,omitempty"`
	City 		string		`xml:"city" json:"city,omitempty"`
	StateOrProvince	string	`xml:"stateOrProvince" json:"state_or_province,omitempty"`
	Country 	string		`xml:"country" json:"country,omitempty"`
	PostalCode 	string		`xml:"postalCode" json:"postal_code,omitempty"`
}
// -----End XML structs -----

// Server Request
type SocketMsgS struct {
	
	Action 	string	`json:"action,omitempty"`
	Value 	string	`json:"value,omitempty"`
}

// Query response
type QueryResS struct {
	
	Query				string		`json:"query,omitempty"`
	NameResult			[]string	`json:"name_result,omitempty"`
	RevNameResult		[]string	`json:"revname_result,omitempty"`
	AkaResult			[]string	`json:"aka_result,omitempty"`
	RevAkaResult		[]string	`json:"revaka_result,omitempty"`
	AddressResult		[]string	`json:"address_result,omitempty"`
	PostalCodeResult	[]string	`json:"postal_code_result,omitempty"`
}
