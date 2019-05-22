package util

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"
)

func TestDistances(t *testing.T) {
	var test = []struct {
		pnts []Point
		expected [][]float64
	}{
		{
			[]Point{{2, 3}, {4, 6}, {0, 1}, {3, 3}, {10, 7}},
			[][]float64{{0, 3.605551275463989, 2.8284271247461903, 1,  8.94427190999916},
				{3.605551275463989, 0, 6.4031242374328485, 3.1622776601683795, 6.082762530298219},
				{2.8284271247461903, 6.4031242374328485, 0, 3.605551275463989, 11.661903789690601},
				{1, 3.1622776601683795, 3.605551275463989, 0, 8.06225774829855},
				{8.94427190999916, 6.082762530298219, 11.661903789690601, 8.06225774829855, 0}},
		},
		{
			[]Point{{2, 3}, {4, 6}, {0, 1}, {3, 3}, {10, 7}},
			[][]float64{{0, 3.605551275463989, 2.8284271247461903, 1, 8.94427190999916},
				{3.605551275463989, 0, 6.4031242374328485, 3.1622776601683795, 6.082762530298219},
				{2.8284271247461903, 6.4031242374328485, 0, 3.605551275463989, 11.661903789690601},
				{1, 3.1622776601683795, 3.605551275463989, 0, 8.06225774829855},
				{8.94427190999916, 6.082762530298219, 11.661903789690601, 8.06225774829855, 0}},
		},
		{
			[]Point{{0, 0}, {19, 3}, {5, 8}, {2, 0}, {0, 15}},
			[][]float64{{0, 19.235384061671343, 9.433981132056603, 2, 15},
				{19.235384061671343, 0, 14.866068747318506, 17.26267650163207, 22.47220505424423},
				{9.433981132056603, 14.866068747318506, 0, 8.54400374531753, 8.602325267042627},
				{2, 17.26267650163207, 8.54400374531753, 0, 15.132745950421556},
				{15, 22.47220505424423, 8.602325267042627, 15.132745950421556, 0}},
		},
	}

	for i, tt := range test {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if !reflect.DeepEqual(Distances(tt.pnts), tt.expected) {
				t.Errorf("return value from Distances does not equal expected return value")
			}
		})
	}
}

//TODO finish this test by figuring out what an appropriate value is for dist
func TestDistribute(t *testing.T) {
	nodes := []string{"1,", "2", "3", "4", "5"}
	dist := []int{}

	fmt.Println(Distribute(nodes, dist))
}

func Test_generateWorstCaseNetwork(t *testing.T) {
	var test = []struct {
		nodes int
		seed int64
		expected [][]int
	}{
		{8, 25, [][]int{{5}, {6}, {1}, {4}, {7}, {2}, {0}, {0}}},
		{10, 123, [][]int{{4}, {0}, {6}, {9}, {3}, {0}, {1}, {2}, {7}, {8}}},
		{5, 123, [][]int{{2}, {0}, {3}, {4}, {1}}},
	}

	for i, tt := range test {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if !reflect.DeepEqual((generateWorstCaseNetwork(tt.nodes, tt.seed)), tt.expected) {
				t.Errorf("return value from GenerateWorstCaseNetwork does not match expected value")
			}
		})
	}
}

// TODO why are tests passing even though test.expected is empty???
func Test_generateUniformRandMeshNetwork(t *testing.T) {
	var test = []struct {
		nodes    int
		conns    int
		seed     int64
		expected [][]int
	}{
		{
			8,
			5,
			123,
			[][]int{
				{3, 1, 7, 6, 5},
				{6, 7, 5, 4, 2},
				{5, 6, 0, 1, 3},
				{2, 0, 4, 6, 1},
				{7, 5, 2, 6, 1},
				{0, 3, 1, 2, 4},
				{0, 1, 7, 5, 4},
				{1, 0, 2, 3, 5},
			},
		},
		{
			10,
			9,
			8,
			[][]int{
				{6, 8, 9, 1, 7, 4, 3, 2, 5},
				{5, 6, 8, 4, 0, 7, 9, 2, 3},
				{4, 7, 0, 9, 8, 6, 5, 1, 3},
				{7, 1, 2, 8, 6, 4, 9, 0, 5},
				{9, 3, 1, 0, 6, 7, 5, 8, 2},
				{8, 2, 9, 0, 4, 6, 7, 1, 3},
				{2, 8, 3, 7, 9, 4, 1, 0, 5},
				{0, 5, 4, 3, 8, 9, 1, 2, 6},
				{0, 2, 5, 3, 1, 6, 9, 7, 4},
				{3, 7, 8, 5, 1, 4, 0, 6, 2},
			},
		},
		{
			16,
			8,
			15,
			[][]int{
				{7, 11, 4, 8, 10, 13, 9, 1},
				{2, 3, 0, 7, 11, 4, 14, 8},
				{13, 5, 12, 1, 15, 10, 7, 0},
				{11, 9, 1, 12, 15, 5, 10, 2},
				{1, 14, 10, 9, 12, 3, 11, 6},
				{3, 4, 9, 12, 11, 0, 2, 8},
				{4, 14, 1, 13, 11, 3, 10, 15},
				{10, 15, 6, 9, 1, 2, 8, 5},
				{0, 2, 5, 6, 14, 7, 3, 15},
				{8, 5, 12, 6, 2, 4, 7, 13},
				{14, 4, 9, 12, 5, 7, 15, 0},
				{0, 13, 1, 5, 12, 15, 4, 2},
				{6, 15, 10, 4, 14, 5, 3, 9},
				{9, 1, 8, 6, 7, 14, 2, 4},
				{12, 4, 5, 11, 3, 10, 15, 0},
				{5, 1, 13, 14, 3, 10, 8, 0},
			},
		},
	}

	for i, tt := range test {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			out, _ := generateUniformRandMeshNetwork(tt.nodes, tt.conns, tt.seed)
			if !reflect.DeepEqual(out, tt.expected) {
				t.Errorf("return value from GenerateUniformRandMeshNetwork does not match expected value")
			}
		})
	}
}

func Test_generateNoDuplicateMeshNetwork(t *testing.T) {
	var test = []struct {
		nodes    int
		conns    int
		seed     int64
		expected [][]int
	}{
		{8, 7, 123, [][]int{}},
		{5, 4, 3, [][]int{}},
		{3, 2, 15, [][]int{}},
	}

	for i, tt := range test {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			out, _ := generateNoDuplicateMeshNetwork(tt.nodes, tt.conns, tt.seed)
			//fmt.Println(reflect.DeepEqual(out, tt.expected))
			if !reflect.DeepEqual(out, tt.expected) {
				t.Errorf("return value from GenerateNoDuplicateMeshNetwork does not match expected return value")
			}
		})
	}
}
