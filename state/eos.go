package state

/**Stores data from the last eos build, which can be queried from the rest api**/

type EosState struct{
	NumberOfAccounts	int `json:"numberOfAccounts"`
}
var eosState *EosState = nil

func GetEosState() *EosState {
	return eosState
}

