package db

//SideCar represents a supporting node within the network
type SideCar struct {
	ID string `json:"id"`

	NodeID string `json:"nodeID"`

	AbsoluteNodeNum int `json:"absNum"`

	// TestNetID is the id of the testnet to which the node belongs to
	TestnetID string `json:"testnetID"`

	// Server is the id of the server on which the node resides
	Server int `json:"server"`

	//LocalID is the number of the node in the testnet
	LocalID int `json:"localID"`

	NetworkIndex int `json:"networkIndex"`

	// IP is the ip address of the node
	IP string `json:"ip"`

	// Image is the docker image on which the sidecar was built
	Image string `json:"image"`

	// Type is the type of sidecar
	Type string `json:"type"`
}



func (n SideCar) GetAbsoluteNumber() int{
	return n.AbsoluteNum
}

func (n SideCar) GetIP() string{
	return n.IP
}

func (n SideCar) GetRelativeNumber() int{
	return n.LocalID
}

func (n SideCar) GetServerID() int{
	return n.Server
}

func (n SideCar) GetTestNetID() string{
	return n.TestnetID
}