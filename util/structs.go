package util


/****Standard Data Structures****/
/*
    KeyPair represents a cryptographic key pair
 */
type KeyPair struct {
    PrivateKey  string  `json:"privateKey"`
    PublicKey   string  `json:"publicKey"`
}

/*
    Service represents a service for a blockchain. 
    All env variables will be passed to the container.
 */
type Service struct {
    Name    string              `json:"name"`
    Image   string              `json:"image"`
    Env     map[string]string   `json:"env"`
}

type EndPoint struct {
    Url     string  `json:"url"`
    User    string  `json:"user"`
    Pass    string  `json:"pass"`
}