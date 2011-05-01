package visualization

import (
	"github.com/mrjones/latvis/location"

	"testing"
)

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

	actual := AggregateHistory(&history, bounds, 5, 5)

	expected := Grid{points: [][]int {
			{0, 0, 0, 0, 0},
			{0, 0, 0, 1, 0},
			{0, 0, 1, 0, 0},
			{0, 1, 0, 0, 0},
			{0, 0, 0, 0, 0},
		},
	height: 5, width: 5,
	}

	assertGridsEqual(t, expected, *actual)
}

func assertGridsEqual(t *testing.T, expected Grid, actual Grid) {
	if expected.Width() != actual.Width() {
		t.Fatalf("Grids have different number of columns. Expected: %d, Actual: %d",
			expected.Width(), actual.Width());
	}

	if expected.Height() != actual.Height() {
		t.Fatalf("Grids have different number rows. Expected: %d, Actual: %d",
			expected.Height(), actual.Height());
	}

	for i := 0 ; i < expected.Width() ; i++ {
		for j := 0 ; j < expected.Height() ; j++ {
			if expected.Get(i, j) != actual.Get(i, j) {
				t.Errorf("Grid mismatch -- grid[%d][%d]. Expected %d, Actual: %d",
					i, j, expected.Get(i, j), actual.Get(i, j))
			}
		}
	}
}

