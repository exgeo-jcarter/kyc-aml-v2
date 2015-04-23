package KycAmlClientServer

type ClientServerQueryReqS struct {
	
	QueryName 		string	`json:"query_name,omitempty"`
	QueryAddress 	string	`json:"query_address,omitempty"`
}

type ClientServerQueryResS struct {
	
	FuzzyName				FuzzyQueryResS				`json:"fuzzy_name,omitempty"`
	FuzzyAddress			FuzzyQueryResS				`json:"fuzzy_address,omitempty"`
	MetaphoneName			MetaphoneQueryResS			`json:"metaphone_name,omitempty"`
	MetaphoneAddress		MetaphoneQueryResS			`json:"metaphone_address,omitempty"`
	DoubleMetaphoneName		DoubleMetaphoneQueryResS	`json:"double_metaphone_name,omitempty"`
	DoubleMetaphoneAddress	DoubleMetaphoneQueryResS	`json:"double_metaphone_address,omitempty"`
	SdnEntry				SdnEntryS					`json:"sdn_entry,omitempty"`
	RiskScore				float64						`json:"risk_score,omitempty"`
	Matches					[]string					`json:"matches,omitempty"`
}

type QueryReqS struct {
	
	Action 	string	`json:"action,omitempty"`
	Value 	string	`json:"value,omitempty"`
}

type FuzzyQueryResS struct {
	
	Query				string		`json:"query,omitempty"`
	NameResult			[]string	`json:"name_result,omitempty"`
	RevNameResult		[]string	`json:"revname_result,omitempty"`
	AkaResult			[]string	`json:"aka_result,omitempty"`
	RevAkaResult		[]string	`json:"revaka_result,omitempty"`
	AddressResult		[]string	`json:"address_result,omitempty"`
	PostalCodeResult	[]string	`json:"postal_code_result,omitempty"`
}

type MetaphoneQueryResS struct {
	
	Query				string		`json:"query,omitempty"`
	EncodedQuery		string		`json:"encoded_query,omitempty"`
	NameResult			[]string	`json:"name_result,omitempty"`
	RevNameResult		[]string	`json:"revname_result,omitempty"`
	AkaResult			[]string	`json:"aka_result,omitempty"`
	RevAkaResult		[]string	`json:"revaka_result,omitempty"`
	AddressResult		[]string	`json:"address_result,omitempty"`
	PostalCodeResult	[]string	`json:"postal_code_result,omitempty"`
}

type DoubleMetaphoneQueryResS struct {
	
	Query				string		`json:"query,omitempty"`
	EncodedQuery1		string		`json:"encoded_query1,omitempty"`
	EncodedQuery2		string		`json:"encoded_query2,omitempty"`
	NameResult1			[]string	`json:"name_result1,omitempty"`
	NameResult2			[]string	`json:"name_result2,omitempty"`
	RevNameResult1		[]string	`json:"revname_result1,omitempty"`
	RevNameResult2		[]string	`json:"revname_result2,omitempty"`
	AkaResult1			[]string	`json:"aka_result1,omitempty"`
	AkaResult2			[]string	`json:"aka_result2,omitempty"`
	RevAkaResult1		[]string	`json:"revaka_result1,omitempty"`
	RevAkaResult2		[]string	`json:"revaka_result2,omitempty"`
	AddressResult1		[]string	`json:"address_result1,omitempty"`
	AddressResult2		[]string	`json:"address_result2,omitempty"`
	PostalCodeResult1	[]string	`json:"postal_code_result1,omitempty"`
	PostalCodeResult2	[]string	`json:"postal_code_result2,omitempty"`
}


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
