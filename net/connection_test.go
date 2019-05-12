package netconf

import (
	"reflect"
	"strconv"
	"testing"
)

func Test_NewConnections(t *testing.T) {
	var test = []struct {
		nodes    int
		expected [][]bool
	}{
		{2, [][]bool{{true, true}, {true, true}}},
		{1, [][]bool{{true}}},
		{3, [][]bool{{true, true, true}, {true, true, true}, {true, true, true}}},
	}

	for i, tt := range test {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if !reflect.DeepEqual((NewConnections(tt.nodes)).cons, tt.expected) {
				t.Errorf("return value of NewConnections does not match expected value")
			}
		})
	}
}

func Test_RemoveAll(t *testing.T) {
	var test = []struct {
		mesh     Connections
		conns    []Connection
		expected Connections
	}{
		{
			Connections{[][]bool{{true, true, true, true}, {true, true, true, true}, {true, true, true, true}, {true, true, true, true}}},
			[]Connection{{2, 3}, {1, 0}},
			Connections{[][]bool{{true, false, true, true}, {true, true, true, true}, {true, true, true, true}, {true, true, false, true}}},
		},
		{
			Connections{[][]bool{{true, true, true, true}, {true, true, true, true}, {true, true, true, true}, {true, true, true, true}}},
			[]Connection{{3, 2}, {0, 1}},
			Connections{[][]bool{{true, true, true, true}, {false, true, true, true}, {true, true, true, false}, {true, true, true, true}}},
		},
		{
			Connections{[][]bool{{true, true, true, true}, {true, true, true, true}, {true, true, true, true}, {true, true, true, true}}},
			[]Connection{{2, 3}, {1, 0}, {3, 1}, {2, 0}},
			Connections{[][]bool{{true, false, false, true}, {true, true, true, false}, {true, true, true, true}, {true, true, false, true}}},
		},
	}

	for i, tt := range test {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			tt.mesh.RemoveAll(tt.conns)
			if !reflect.DeepEqual(tt.mesh.cons, tt.expected) {
				t.Errorf("not all expected values were removed")
			}
		})
	}
}

func Test_findPossiblePeers(t *testing.T) {
	var test = []struct {
		cons     [][]bool
		node     int
		expected []int
	}{
		{[][]bool{{false, true, true}, {true, true, true}, {false, false, false}, {true, false, true}}, 1, []int{0, 2}},
		{[][]bool{{false, true, true}, {false, true, true}, {true, true, true}, {true, false, true}}, 2, []int{0, 1}},
		{[][]bool{{true, true, true}, {true, false, true}, {false, false, false}, {true, false, true}}, 0, []int{1, 2}},
	}

	for i, tt := range test {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if !reflect.DeepEqual(findPossiblePeers(tt.cons, tt.node), tt.expected) {
				t.Errorf("return value of findPossiblePeers did not match expected value")
			}
		})
	}
}

func Test_filterPeers(t *testing.T) {
	var test = []struct {
		peers    []int
		already  []int
		expected []int
	}{
		{[]int{1, 2, 3, 4, 5}, []int{2, 4}, []int{1, 3, 5}},
		{[]int{5, 6, 7, 8, 9}, []int{}, []int{5, 6, 7, 8, 9}},
		{[]int{1, 2, 6, 7, 8, 9}, []int{7, 8, 9}, []int{1, 2, 6}},
	}

	for i, tt := range test {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if !reflect.DeepEqual(filterPeers(tt.peers, tt.already), tt.expected) {
				t.Errorf("return value of filterPeers did not match expected value")
			}
		})
	}
}

func Test_mergeUniquePeers(t *testing.T) {
	var test = []struct {
		peers1   []int
		peers2   []int
		expected []int
	}{
		{[]int{1, 5, 7, 8, 9, 15}, []int{4, 8, 3, 1, 10, 2}, []int{1, 5, 7, 8, 9, 15, 4, 3, 10, 2}},
		{[]int{2, 5, 8, 10}, []int{1, 2, 9, 5, 18, 6}, []int{2, 5, 8, 10, 1, 9, 18, 6}},
		{[]int{3, 6, 7, 2, 1}, []int{1, 2, 7, 5}, []int{3, 6, 7, 2, 1, 5}},
	}

	for i, tt := range test {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if !reflect.DeepEqual(mergeUniquePeers(tt.peers1, tt.peers2), tt.expected) {
				t.Errorf("return value of mergeUniquePeers did not match expected value")
			}
		})
	}
}

func Test_containsPeer(t *testing.T) {
	var test = []struct {
		peers       []int
		validPeer   int
		invalidPeer int
	}{
		{[]int{1, 2, 3, 4, 5, 6}, 1, 9},
		{[]int{15, 16, 4, 5, 6}, 16, 300},
		{[]int{7, 8, 9, 2, 3, 4, 5, 6, 40, 34, 80}, 40, 0},
	}

	for i, tt := range test {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if !containsPeer(tt.peers, tt.validPeer) {
				t.Errorf("return value of containsPeer did not match expected value")
			}

			if containsPeer(tt.peers, tt.invalidPeer) {
				t.Errorf("return value of containsPeer did not match expected value")
			}
		})
	}
}

func Test_Networks(t *testing.T) {
	var test = []struct {
		mesh     Connections
		expected [][]int
	}{
		{Connections{[][]bool{{true, true, true, true}, {true, true, true, true}, {true, true, true, true}, {true, true, true, true}}}, [][]int{{0, 1, 2, 3}}},
		{Connections{[][]bool{{true, true}, {true, true}}}, [][]int{{0, 1}}},
		{Connections{[][]bool{{true, true, true}, {true, true, true}, {true, true, true}}}, [][]int{{0, 1, 2}}},
	}

	for i, tt := range test {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if !reflect.DeepEqual(tt.mesh.Networks(), tt.expected) {
				t.Errorf("return value of Networks does not match expected value")
			}
		})
	}
}
