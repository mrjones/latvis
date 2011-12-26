package visualization

import (
	"github.com/mrjones/latvis/location"

	"fmt"
	"testing"
)

func TestSimpleGrid(t *testing.T) {
	g := NewGrid(1, 1)
	if 0 != g.Get(0, 0) {
		t.Fatalf("[0,0] should have been '0', but was: %d\n", g.Get(0, 0))
	}

	g.Inc(0, 0)
	if 1 != g.Get(0, 0) {
		t.Fatalf("[0,0] should have been '1', but was: %d\n", g.Get(0, 0))
	}

	g.Set(0, 0, 1000)
	if 1000 != g.Get(0, 0) {
		t.Fatalf("[0,0] should have been '1000', but was: %d\n", g.Get(0, 0))
	}
}

func TestSimpleAggregateHistory(t *testing.T) {
	history := location.History{}
	history.Add(&location.Coordinate{Lat: 1.0, Lng: 1.0})
	history.Add(&location.Coordinate{Lat: 2.0, Lng: 2.0})
	history.Add(&location.Coordinate{Lat: 3.0, Lng: 3.0})

	bounds, err := location.NewBoundingBox(
		location.Coordinate{Lat: 0.0, Lng: 0.0},
		location.Coordinate{Lat: 4.0, Lng: 4.0})

	if err != nil {
		t.Fatal(err)
	}

	actual := aggregateHistory(&history, bounds, 5, 5)

	expected := gridLiteral([][]int{
		{0, 0, 0, 0, 0},
		{0, 0, 0, 1, 0},
		{0, 0, 1, 0, 0},
		{0, 1, 0, 0, 0},
		{0, 0, 0, 0, 0},
	})

	assertGridsEqual(t, expected, actual)
}

// X 0                
// 0 0 --\  4 wide x =  1 0 0 0 
// X 0 --/   3 tall  =  1 0 0 0
// 0 0                  0 0 0 0
// 0 0
// 0 0
func TestSqueezesTallBoxIntoWideImageNoDistortion(t *testing.T) {
	history := location.History{}
	history.Add(&location.Coordinate{Lat: 5.0, Lng: 0.0})
	history.Add(&location.Coordinate{Lat: 3.0, Lng: 0.0})

	bounds, err := location.NewBoundingBox(
		location.Coordinate{Lat: 0.00001, Lng: -0.00001},
		location.Coordinate{Lat: 5.00001, Lng: 1.00001})

	if err != nil {
		t.Fatal(err)
	}

	actual := aggregateHistory(&history, bounds, 4, 3)

	expected := gridLiteral([][]int{
		{1, 0, 0, 0},
		{1, 0, 0, 0},
		{0, 0, 0, 0},
	})
	assertGridsEqual(t, expected, actual)
}

// X X X X                
// 0 0 0 0 --\  4 wide x =  2 2 0 0 
// 0 0 0 0 --/   3 tall  =  0 0 0 0
// 0 0 0 0                  0 0 0 0
// 0 0 0 0
// 0 0 0 0
func TestSqueezesLineWithoutDistortion(t *testing.T) {
	history := location.History{}
	history.Add(&location.Coordinate{Lat: 5.0, Lng: 0.0})
	history.Add(&location.Coordinate{Lat: 5.0, Lng: 1.0})
	history.Add(&location.Coordinate{Lat: 5.0, Lng: 2.0})
	history.Add(&location.Coordinate{Lat: 5.0, Lng: 3.0})

	bounds, err := location.NewBoundingBox(
		location.Coordinate{Lat: -0.00001, Lng: -0.00001},
		location.Coordinate{Lat: 5.00001, Lng: 3.00001})

	if err != nil {
		t.Fatal(err)
	}

	actual := aggregateHistory(&history, bounds, 4, 3)

	expected := gridLiteral([][]int{
		{2, 2, 0, 0},
		{0, 0, 0, 0},
		{0, 0, 0, 0},
	})
	assertGridsEqual(t, expected, actual)
}

// 0 0 0 0
// X X X X --\  4 wide x =  2 2 0 0 
// 0 0 0 0 --/   3 tall  =  0 0 0 0
// 0 0 0 0                  0 0 0 0
// 0 0 0 0
// 0 0 0 0
func TestSqueezesLineWithoutDistortion2(t *testing.T) {
	history := location.History{}
	history.Add(&location.Coordinate{Lat: 4.0, Lng: 0.0})
	history.Add(&location.Coordinate{Lat: 4.0, Lng: 1.0})
	history.Add(&location.Coordinate{Lat: 4.0, Lng: 2.0})
	history.Add(&location.Coordinate{Lat: 4.0, Lng: 3.0})

	bounds, err := location.NewBoundingBox(
		location.Coordinate{Lat: -0.00001, Lng: -0.00001},
		location.Coordinate{Lat: 5.00001, Lng: 3.00001})

	if err != nil {
		t.Fatal(err)
	}

	actual := aggregateHistory(&history, bounds, 4, 3)

	expected := gridLiteral([][]int{
		{2, 2, 0, 0},
		{0, 0, 0, 0},
		{0, 0, 0, 0},
	})
	assertGridsEqual(t, expected, actual)
}

// X X X X                
// X X X X --\  4 wide x =  4 4 0 0 
// X X X X --/   3 tall  =  4 4 0 0
// X X X X                  0 0 0 0
// 0 0 0 0
// 0 0 0 0
func TestSqueezesSquareWithoutDistortion(t *testing.T) {
	history := location.History{}
	history.Add(&location.Coordinate{Lat: 5.0, Lng: 0.0})
	history.Add(&location.Coordinate{Lat: 5.0, Lng: 1.0})
	history.Add(&location.Coordinate{Lat: 5.0, Lng: 2.0})
	history.Add(&location.Coordinate{Lat: 5.0, Lng: 3.0})
	history.Add(&location.Coordinate{Lat: 4.0, Lng: 0.0})
	history.Add(&location.Coordinate{Lat: 4.0, Lng: 1.0})
	history.Add(&location.Coordinate{Lat: 4.0, Lng: 2.0})
	history.Add(&location.Coordinate{Lat: 4.0, Lng: 3.0})
	history.Add(&location.Coordinate{Lat: 3.0, Lng: 0.0})
	history.Add(&location.Coordinate{Lat: 3.0, Lng: 1.0})
	history.Add(&location.Coordinate{Lat: 3.0, Lng: 2.0})
	history.Add(&location.Coordinate{Lat: 3.0, Lng: 3.0})
	history.Add(&location.Coordinate{Lat: 2.0, Lng: 0.0})
	history.Add(&location.Coordinate{Lat: 2.0, Lng: 1.0})
	history.Add(&location.Coordinate{Lat: 2.0, Lng: 2.0})
	history.Add(&location.Coordinate{Lat: 2.0, Lng: 3.0})

	bounds, err := location.NewBoundingBox(
		location.Coordinate{Lat: -0.00001, Lng: -0.00001},
		location.Coordinate{Lat: 5.00001, Lng: 3.00001})

	if err != nil {
		t.Fatal(err)
	}

	actual := aggregateHistory(&history, bounds, 4, 3)

	expected := gridLiteral([][]int{
		{4, 4, 0, 0},
		{4, 4, 0, 0},
		{0, 0, 0, 0},
	})
	assertGridsEqual(t, expected, actual)
}

func debug(expected, actual *Grid) {
	fmt.Println("Expected")
	printGrid(expected)
	fmt.Println("Actual")
	printGrid(actual)
}

func printGrid(g *Grid) {
	for y := 0; y < g.Height(); y++ {
		for x := 0; x < g.Width(); x++ {
			fmt.Printf("%d ", g.Get(x, y))
		}
		fmt.Println()
	}
}

func gridLiteral(literal [][]int) *Grid {
	// remember this is reversed
	grid := NewGrid(len(literal[0]), len(literal))
	for i := 0; i < len(literal); i++ {
		for j := 0; j < len(literal[i]); j++ {
			grid.Set(j, i, literal[i][j])
		}
	}
	return grid
}

func assertGridsEqual(t *testing.T, expected *Grid, actual *Grid) {
	if expected.Width() != actual.Width() {
		debug(expected, actual)
		t.Fatalf("Grids have different number of columns. Expected: %d, Actual: %d",
			expected.Width(), actual.Width())
	}

	if expected.Height() != actual.Height() {
		debug(expected, actual)
		t.Fatalf("Grids have different number rows. Expected: %d, Actual: %d",
			expected.Height(), actual.Height())
	}

	failed := false
	for i := 0; i < expected.Width(); i++ {
		for j := 0; j < expected.Height(); j++ {
			if expected.Get(i, j) != actual.Get(i, j) {
				failed = true
				t.Errorf("Grid mismatch -- grid[%d][%d]. Expected %d, Actual: %d",
					i, j, expected.Get(i, j), actual.Get(i, j))
			}
		}
	}
	if failed {
		debug(expected, actual)
	}
}
