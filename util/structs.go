package util

/****Standard Data Structures****/

// KeyPair represents a cryptographic key pair
type KeyPair struct {
	PrivateKey string `json:"privateKey"`
	PublicKey  string `json:"publicKey"`
}

// Command represents a previously executed command
type Command struct {
	Cmdline  string
	Node     int
	ServerID int
}

// Service represents a service for a blockchain.
// All env variables will be passed to the container.
type Service struct {
	Name    string            `json:"name"`
	Image   string            `json:"image"`
	Env     map[string]string `json:"env"`
	Network string            `json:"network"`
}

// EndPoint represents an endpoint with basic auth
type EndPoint struct {
	URL  string `json:"url"`
	User string `json:"user"`
	Pass string `json:"pass"`
}