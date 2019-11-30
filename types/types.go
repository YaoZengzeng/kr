package types

type Service struct {
	Address		string 	`json:"address"`
	Port		int    	`json:"port"`
	Endpoint	string	`json:"endpoint"`
}
