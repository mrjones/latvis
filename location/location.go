package location

import (
	"os"
)

type Coordinate struct {
	Lat float64
	Lng float64
}

type BoundingBox struct {
	lowerLeft Coordinate
	upperRight Coordinate
}

func NewBoundingBox(lowerLeft, upperRight Coordinate) (*BoundingBox, os.Error) {
	if lowerLeft.Lng > upperRight.Lng {
		return nil, os.NewError("Longitude of lowerLeft must be less than longitude of upperRight")
	}
	return &BoundingBox{lowerLeft: lowerLeft, upperRight: upperRight}, nil
}

func (b *BoundingBox) Contains(c *Coordinate) bool {
	isReversed := b.lowerLeft.Lat > b.upperRight.Lat
	boxShift := 0.0
	pointShift := 0.0
	if isReversed {
		boxShift = 360.0
		if c.Lat < 0 {
			pointShift = 360.0
		}
	}

	return c.Lat + pointShift > b.lowerLeft.Lat &&
		c.Lat + pointShift < b.upperRight.Lat + boxShift &&
		c.Lng > b.lowerLeft.Lng &&
		c.Lng < b.upperRight.Lng
}

type History []*Coordinate

func (h *History) Len() int {
	return len(*h)
}

func (h *History) Add(c *Coordinate) {
	history := *h
	curLen := len(history)

	if curLen + 1 > cap(history) {
		newHistory := make([]*Coordinate, curLen, 2 * curLen + 1)
		copy(newHistory, history)
		history = newHistory
	}
	history = history[0 : curLen + 1]
	history[curLen] = c
	*h = history
}

func (this *History) AddAll(that *History) {
	for i := 0 ; i < that.Len(); i++ {
		this.Add(that.At(i))
	}
}

func (h *History) At(i int) *Coordinate {
	return (*h)[i]
}

type HistorySource interface {
	GetHistory(year int64, month int) (*History, os.Error)
}
