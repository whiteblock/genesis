package netconf

import (
	"reflect"
	"testing"
)

func Test_NewConnections(t *testing.T) {
	nodes := 2

	out := NewConnections(nodes)

	expected := [][]bool{{true, true}, {true, true}}

	if !reflect.DeepEqual(out.cons, expected) {
		t.Errorf("return value of NewConnections does not match expected value")
	}
}

func Test_RemoveAll(t *testing.T) {
	mesh := Connections{[][]bool{{true, true, true, true}, {true, true, true, true}, {true, true, true, true}, {true, true, true, true}}}
	con1 := Connection{To: 2, From: 3}
	con2:= Connection{To: 1, From: 0}

	conns := []Connection{con1, con2}

	mesh.RemoveAll(conns)

	expected := Connections{[][]bool{{true, false, true, true}, {true, true, true, true}, {true, true, true, true}, {true, true, false, true}}}

	if !reflect.DeepEqual(mesh.cons, expected.cons) {
		t.Errorf("not all expected values were removed")
	}
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

func Test_filterPeers(t *testing.T) {
	peers := []int{1, 2, 3, 4, 5}
	alreadyDone := []int{3}

	out := filterPeers(peers, alreadyDone)

	expected := []int{1, 2, 4, 5}

	if !reflect.DeepEqual(out, expected) {
		t.Errorf("return value of filterPeers did not match expected value")
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

func Test_containsPeer(t *testing.T) {
	peers := []int{1, 2, 3, 4, 5, 6, 7, 8}

	if !containsPeer(peers, 4) {
		t.Errorf("return value of containsPeer did not match expected value")
	}

	if containsPeer(peers, 18) {
		t.Errorf("return value of containsPeer did not match expected value")
	}
}

func Test_Networks(t *testing.T) {
	mesh := Connections{[][]bool{{true, true, true, true}, {true, true, true, true}, {true, true, true, true}, {true, true, true, true}}}

	out := mesh.Networks()

	expected := [][]int{{0, 1, 2, 3}}

	if !reflect.DeepEqual(out, expected) {
		t.Errorf("return value of Networks does not match expected value")
	}
}