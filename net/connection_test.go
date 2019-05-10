package netconf

import (
	"reflect"
	"testing"
)

func Test_findPossiblePeers(t *testing.T) {

	cons := [][]bool{{false, true, true}, {true, true, true}, {false, false, false}, {true, false, true}}
	node := 1

	out := findPossiblePeers(cons, node)

	expected := []int{0, 2}

	if !reflect.DeepEqual(out, expected) {
		t.Errorf("return value of findPossiblePeers did not match expected value")
	}
}

func Test_containsPeer(t *testing.T) {
	peers := []int{1, 2, 3, 4, 5, 6, 7, 8}

	if !containsPeer(peers, 4) {
		t.Errorf("return value of containsPeer did not match expected value")
	}

	if containsPeer(peers, 18) {
		t.Errorf("return value of containsPeer did not match expected value")
	}

}

func Test_mergeUniquePeers(t *testing.T) {
	peers1 := []int{1, 5, 7, 8, 9, 15}
	peers2 := []int{4, 8, 3, 1, 10, 2}

	out := mergeUniquePeers(peers1, peers2)

	expected := []int{1, 5, 7, 8, 9, 15, 4, 3, 10, 2}

	if !reflect.DeepEqual(out, expected) {
		t.Errorf("return value of mergeUniquePeers did not match expected value")
	}
}