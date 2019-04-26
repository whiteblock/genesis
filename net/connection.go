package netconf

//Connection represents a uni-directional connection
type Connection struct {
	To   int `json:"to"`
	From int `json:"from"`
}

//Connections represents a graph of the network connections
type Connections struct {
	cons [][]bool //[from][to]
}

//NewConnections creates a Connections object with all connections marked
//as open
func NewConnections(nodes int) *Connections {
	out := new(Connections)
	out.cons = make([][]bool, nodes)
	for i := range out.cons {
		out.cons[i] = make([]bool, nodes)
		for j := range out.cons[i] {
			out.cons[i][j] = true
		}
	}
	return out
}

//RemoveAll will mark all of the given connections as not
//connected
func (this *Connections) RemoveAll(conns []Connection) {
	for _, conn := range conns {
		this.cons[conn.From][conn.To] = false
	}
}

func findPossiblePeers(cons [][]bool, node int) []int {
	out := []int{}
	for i, con := range cons[node] {
		if node == i {
			continue
		}
		if con || cons[i][node] {
			out = append(out, i)
		}
	}
	return out
}

func filterPeers(peers []int, alreadyDone []int) []int {
	if len(alreadyDone) == 0 {
		return peers
	}
	out := []int{}
	for _, peer := range peers {
		shouldAdd := true
		for _, done := range alreadyDone {
			if done == peer {
				shouldAdd = false
			}
		}
		if shouldAdd {
			out = append(out, peer)
		}
	}
	return out
}

func mergeUniquePeers(peers1 []int, peers2 []int) []int {
	out := make([]int, len(peers1))
	copy(out, peers1)

	for _, peer := range peers2 {
		if !containsPeer(out, peer) {
			out = append(out, peer)
		}
	}
	return out
}

func containsPeer(peers []int, peer int) bool {
	for _, p := range peers {
		if peer == p {
			return true
		}
	}
	return false
}

//Networks calculates the distinct, completely separate partitions in the network
func (this *Connections) Networks() [][]int {
	nodes := []int{}
	nodesFinalized := []int{}
	nodesToTry := []int{}

	out := [][]int{}

	for len(nodesFinalized) < len(this.cons) {
		//fmt.Printf("\n\nNodes : %#v\n Nodes Finalized: %#v\nNodes To Try%#v\n\n",nodes,nodesFinalized,nodesToTry)
		if len(nodesToTry) == 0 {
			if len(nodes) > 0 {
				nodesFinalized = mergeUniquePeers(nodes, nodesFinalized)
				out = append(out, nodes)
				nodes = []int{}
			}
			for i := 0; i < len(this.cons); i++ {
				if !containsPeer(nodesFinalized, i) {
					nodesToTry = findPossiblePeers(this.cons, i)
					nodes = []int{i}
					nodes = mergeUniquePeers(nodes, nodesToTry)
					break
				}
			}

		} else {
			newPeers := findPossiblePeers(this.cons, nodesToTry[0])
			nodes = mergeUniquePeers(nodes, []int{nodesToTry[0]})
			newPeers = filterPeers(newPeers, nodes)
			if len(nodesToTry) > 1 {
				nodesToTry = nodesToTry[1:]
			} else {
				nodesToTry = []int{}
			}
			nodesToTry = mergeUniquePeers(nodesToTry, newPeers)

		}
	}
	return out
}
