package KycAmlClient

type QueryReqS struct {
	
	Action 	string	`json:"action,omitempty"`
	Value 	string	`json:"value,omitempty"`
}

type QueryResS struct {
	
	Query			string		`json:"query,omitempty"`
	MetaphoneQuery	string		`json:"metaphone_query"`
	Result 			[]string	`json:"result,omitempty"`
	MetaphoneResult	[]string	`json:"metaphone_result,omitempty"`
	RiskScore		float64		`json:"risk_score"`
} 
