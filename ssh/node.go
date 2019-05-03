package ssh

// Node represents the interface which all nodes must follow.
type Node interface {
	GetAbsoluteNumber() int
	GetIP() string
	GetRelativeNumber() int
	GetServerID() int
	GetTestNetID() string
	GetNodeName() string
}
