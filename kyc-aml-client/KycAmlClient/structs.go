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
