package util


/****Standard Data Structures****/
type KeyPair struct {
    PrivateKey  string  `json:"privateKey"`
    PublicKey   string  `json:"publicKey"`
}

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