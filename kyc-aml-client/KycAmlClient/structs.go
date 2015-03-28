package KycAmlClient

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