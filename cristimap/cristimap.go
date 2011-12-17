package main

import (
	"github.com/mrjones/latvis/location"
	"github.com/mrjones/latvis/visualization"

	"io/ioutil"
	"os"
	"time"
)

type StaticHistorySource struct { }

func (s *StaticHistorySource) GetHistory(year int64, month int) (*location.History, os.Error) {
	return nil, nil
}

func (s *StaticHistorySource) FetchRange(start, end time.Time) (*location.History, os.Error) {
	h := make(location.History, 1000)
	h = append(h,  &location.Coordinate{Lat:41.3818500, Lng:-74.6860244})
	return &h, nil
}

func main() {
	var data location.HistorySource
	data = &StaticHistorySource{}
	bounds, err := location.NewBoundingBox(
		  location.Coordinate{Lat: 40.6, Lng: -74.7},
     	location.Coordinate{Lat: 41.4, Lng: -73.9})

	t := time.LocalTime();

	if err != nil {
		panic(err)
	}

	vis := visualization.NewVisualizer(
		512,
		&data,
		bounds,
		*t,
		*t)

	bytes, err := vis.Bytes()
	if err != nil {
		panic(err)
	}

	ioutil.WriteFile("/var/www/cristimap.png", *bytes, 0600)
}
