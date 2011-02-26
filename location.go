package location

import (
	"os"
	"xml"
	"strconv"
	"strings"
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
	GetHistory() (history *History, err os.Error)
}

type LatitudeXmlFile struct {
	filename string
}

func NewLatitudeXmlFile(filename string) (xmlFile *LatitudeXmlFile) {
	return &LatitudeXmlFile{filename: filename}
}

func (xmlFile *LatitudeXmlFile) GetHistory() (*History, os.Error) {
	history := &History{}
	file, err := os.Open(xmlFile.filename, os.O_RDONLY, 0666)
	if err != nil {
		return nil, err
	}

	p := xml.NewParser(file)
	inCoordinates := false

	for token, err := p.Token(); err == nil; token, err = p.Token() {
		switch t := token.(type) {
		case xml.StartElement:
			if t.Name.Local == "coordinates" {
				inCoordinates = true
			}
		case xml.CharData:
			if inCoordinates {
				parts := strings.Split(string([]byte(t)), ",", -1)
				lat, err := strconv.Atof(parts[0])
				if err != nil {
					return nil, err
				}
				lng, err := strconv.Atof(parts[1])
				if err != nil {
					return nil, err
				}
				point := &Coordinate{Lat: lat, Lng: lng}
				if lat > -74.02 && lat < -73.96 && lng > 40.703 && lng < 40.8 {
					history.Add(point)
				}
			}
		case xml.EndElement:
			if t.Name.Local == "coordinates" {
				inCoordinates = false
			}
		}
	}

	return history, nil

}
