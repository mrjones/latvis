package latitude_xml

import (
	"github.com/mrjones/latvis/location"

	"fmt"
	"os"
	"strconv"
	"strings"
	"xml"
)

type FileSet struct {
	directory string
}

func New(directory string) (xmlFileSet *FileSet) {
	return &FileSet{directory: directory}
}

func (files *FileSet) GetHistory(year int64, month int) (*location.History, os.Error) {
	history := &location.History{}
	filename := fmt.Sprintf("%s/%0.4d-%0.2d.kml", files.directory, year, month)
	file, err := os.Open(filename, os.O_RDONLY, 0666)
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
				lat, err := strconv.Atof64(parts[0])
				if err != nil {
					return nil, err
				}
				lng, err := strconv.Atof64(parts[1])
				if err != nil {
					return nil, err
				}
				point := &location.Coordinate{Lat: lat, Lng: lng}
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
