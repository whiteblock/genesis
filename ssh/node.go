package ssh

type Node interface {
	GetAbsoluteNumber() int
	GetIP() string
	GetRelativeNumber() int
	GetServerID() int
	GetTestNetID() string
	GetNodeName() string
}
