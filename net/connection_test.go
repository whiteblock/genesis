package netconf

import (
	"fmt"
	"reflect"
	"testing"
)

func Test_NewConnections(t *testing.T) {
	nodes := 2

	out := NewConnections(nodes)

	fmt.Println(out.cons)

	expected := [][]bool{{true, true}, {true, true}}

	if !reflect.DeepEqual(out.cons, expected) {
		t.Errorf("return value of NewConnections does not match expected value")
	}
}


func Test_RemoveAll(t *testing.T) {

}

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

func Test_filterPeers(t *testing.T) {
	peers := []int{1, 2, 3, 4, 5}
	alreadyDone := []int{3}

	out := filterPeers(peers, alreadyDone)

	expected := []int{1, 2, 4, 5}

	if !reflect.DeepEqual(out, expected) {
		t.Errorf("return value of filterPeers did not match expected value")
	}
}

func Test_Networks(t *testing.T) {

}