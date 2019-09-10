package generic

import (
	"testing"
)

func TestCreatingNetworkTopologyAll(t *testing.T) {
	str, err := createPeers(1, map[int]string{0:"a", 1:"b", 2:"c"}, all)

}

func TestCreatingNetworkTopologySequence(t *testing.T) {
	str, err := createPeers(1, map[int]string{0:"a", 1:"b", 2:"c"}, all)

}

func TestCreatingNetworkTopologyRandom(t *testing.T) {
	str, err := createPeers(1, map[int]string{0:"a", 1:"b", 2:"c"}, all)

}