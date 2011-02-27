package location

import (
	"os"
)

type Coordinate struct {
	Lat float
	Lng float
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
