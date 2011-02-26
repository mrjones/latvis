package latitude_xml

import (
	"./location"
	"os"
	"strconv"
	"strings"
	"xml"
)

type File struct {
	filename string
}

func New(filename string) (xmlFile *File) {
	return &File{filename: filename}
}

func (xmlFile *File) GetHistory() (*location.History, os.Error) {
	history := &location.History{}
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
