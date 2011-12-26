package location

import (
	"os"
	"time"
)

type Coordinate struct {
	Lat float64
	Lng float64
}

type BoundingBox struct {
	lowerLeft  Coordinate
	upperRight Coordinate
}

func NewBoundingBox(lowerLeft, upperRight Coordinate) (*BoundingBox, os.Error) {
	if lowerLeft.Lat > upperRight.Lat {
		return nil, os.NewError("Latitude of lowerLeft must be less than longitude of upperRight")
	}
	return &BoundingBox{lowerLeft: lowerLeft, upperRight: upperRight}, nil
}

func (b *BoundingBox) LowerLeft() Coordinate {
	return b.lowerLeft
}

func (b *BoundingBox) UpperRight() Coordinate {
	return b.upperRight
}

func (b *BoundingBox) isReversed() bool {
	return b.lowerLeft.Lng > b.upperRight.Lng
}

func (b *BoundingBox) Contains(c *Coordinate) bool {
	boxShift := 0.0
	pointShift := 0.0
	if b.isReversed() {
		boxShift = 360.0
		if c.Lng < 0 {
			pointShift = 360.0
		}
	}

	return c.Lat > b.lowerLeft.Lat &&
		c.Lat < b.upperRight.Lat &&
		c.Lng+pointShift > b.lowerLeft.Lng &&
		c.Lng+pointShift < b.upperRight.Lng+boxShift
}

func (b *BoundingBox) WidthFraction(c *Coordinate) float64 {
	if b.isReversed() && c.Lng < 0 {
		return (c.Lng + 360.0 - b.lowerLeft.Lng) / b.Width()
	}
	return (c.Lng - b.lowerLeft.Lng) / b.Width()
}

func (b *BoundingBox) HeightFraction(c *Coordinate) float64 {
	return (c.Lat - b.lowerLeft.Lat) / b.Height()
}

func (b *BoundingBox) Width() float64 {
	if b.isReversed() {
		return b.upperRight.Lng - b.lowerLeft.Lng + 360
	}
	return b.upperRight.Lng - b.lowerLeft.Lng
}

func (b *BoundingBox) Height() float64 {
	return b.upperRight.Lat - b.lowerLeft.Lat
}

type History []*Coordinate

func (h *History) Len() int {
	return len(*h)
}

func (h *History) Add(c *Coordinate) {
	history := *h
	curLen := len(history)

	if curLen+1 > cap(history) {
		newHistory := make([]*Coordinate, curLen, 2*curLen+1)
		copy(newHistory, history)
		history = newHistory
	}
	history = history[0 : curLen+1]
	history[curLen] = c
	*h = history
}

func (this *History) AddAll(that *History) {
	for i := 0; i < that.Len(); i++ {
		this.Add(that.At(i))
	}
}

func (h *History) At(i int) *Coordinate {
	return (*h)[i]
}

type HistorySource interface {
	GetHistory(year int64, month int) (*History, os.Error)
	FetchRange(start, end time.Time) (*History, os.Error)
}
