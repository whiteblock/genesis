package state

/**Stores data from the last eos build, which can be queried from the rest api**/

type EosState struct{
	NumberOfAccounts	int64 `json:"numberOfAccounts"`
}
var eosState *EosState = nil


func SetEOSNumberOfAccounts(numberOfAccounts int64) {
	eosState = new(EosState)
	eosState.NumberOfAccounts = numberOfAccounts
}
func GetEosState() *EosState {
	return eosState
}



