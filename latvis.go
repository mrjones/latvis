package main

import (
	"container/vector"
	"fmt"
	"image"
	"image/png"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"xml"
)

type heatmap struct {
	points [512][512]float
}

type coordinate struct {
	x float32
	y float32
}

func readLocationHistory(filename string) (*vector.Vector, os.Error) {
	file, err := os.Open(filename, os.O_RDONLY, 0666)
	if err != nil {
		return nil, err
	}

	p := xml.NewParser(file)
	inCoordinates := false
	var result *vector.Vector = new(vector.Vector)

	for token, err := p.Token(); err == nil; token, err = p.Token() {
		switch t := token.(type) {
		case xml.StartElement:
			if t.Name.Local == "coordinates" {
				inCoordinates = true
			}
		case xml.CharData:
			if inCoordinates {
				parts := strings.Split(string([]byte(t)), ",", -1)
				x, err := strconv.Atof32(parts[0])
				if err != nil {
					return nil, err
				}
				y, err := strconv.Atof32(parts[1])
				if err != nil {
					return nil, err
				}
				point := coordinate{x: x, y: y}
				if x > -74.02 && x < -73.96 && y > 40.703 && y < 40.8 {
					result.Push(point)
				}
			}
		case xml.EndElement:
			if t.Name.Local == "coordinates" {
				inCoordinates = false
			}
		}
	}

	return result, nil
}

func scaleHeat(input int) float {
	return float(math.Sqrt(math.Sqrt(float64(input))))
}

func generateHeatmap(points *vector.Vector, size int) heatmap {
	var result heatmap
	maxX := points.At(0).(coordinate).x
	minX := points.At(0).(coordinate).x
	maxY := points.At(0).(coordinate).y
	minY := points.At(0).(coordinate).y

	for i := 0; i < points.Len(); i++ {
		if points.At(i).(coordinate).x < minX {
			minX = points.At(i).(coordinate).x
		}
		if points.At(i).(coordinate).x > maxX {
			maxX = points.At(i).(coordinate).x
		}
		if points.At(i).(coordinate).y < minY {
			minY = points.At(i).(coordinate).y
		}
		if points.At(i).(coordinate).y > maxY {
			maxY = points.At(i).(coordinate).y
		}
	}

	fmt.Printf("xrange %f %f\n", minX, maxX)
	fmt.Printf("yrange %f %f\n", minY, maxY)

	var counts [512][512]int

	xScale := float32(size-1) / (maxX - minX)
	yScale := float32(size-1) / (maxY - minY)
	scale := xScale
	if yScale < xScale {
		scale = yScale
	}

	for i := 0; i < points.Len(); i++ {
		xBucket := int((points.At(i).(coordinate).x - minX) * scale)
		yBucket := size - int((points.At(i).(coordinate).y-minY)*scale) - 1
		//        fmt.Printf("counter[%f][%f]++\n", points.At(i).(coordinate).x, points.At(i).(coordinate).y)
		//        fmt.Printf("counter[%d][%d]++\n", xBucket, yBucket)
		counts[xBucket][yBucket]++
	}

	maxCount := float(0.0)
	for x := 0; x < len(counts); x++ {
		for y := 0; y < len(counts[x]); y++ {
			if scaleHeat(counts[x][y]) > maxCount {
				maxCount = scaleHeat(counts[x][y])
				fmt.Printf("count[%d][%d] = %d\n", x, y, scaleHeat(counts[x][y]))
			}
		}
	}

	fmt.Printf("max: %d\n", maxCount)

	for x := 0; x < len(counts); x++ {
		for y := 0; y < len(counts[x]); y++ {
			result.points[x][y] = scaleHeat(counts[x][y]) / float(maxCount)
		}
	}

	return result
}

func readAndAppendData(filename string, points *vector.Vector) {
	localPoints, err := readLocationHistory(filename)
	if err != nil {
		log.Exit(err)
	}
	points.AppendVector(localPoints)
}

func imageOfHeatmap(mapdata heatmap, size int) image.Image {
	img := image.NewNRGBA(size, size)
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			val := mapdata.points[x][y]
			if val > 0 {
				img.Pix[y*img.Stride+x] = image.NRGBAColor{uint8(0), uint8(0), uint8(0), 255}
			} else {
				img.Pix[y*img.Stride+x] = image.NRGBAColor{uint8(255), uint8(255), uint8(255), 255}
			}
		}
	}
	return img
}

func renderImage(img image.Image, filename string) {
	f, err := os.Open(filename, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Exit(err)
	}
	if err = png.Encode(f, img); err != nil {
		log.Exit(err)
	}
}

func main() {
	size := 512

	points := new(vector.Vector)
	readAndAppendData("/home/mrjones/src/latvis/data/2010-07.kml", points)
	readAndAppendData("/home/mrjones/src/latvis/data/2010-08.kml", points)
	readAndAppendData("/home/mrjones/src/latvis/data/2010-09.kml", points)
	readAndAppendData("/home/mrjones/src/latvis/data/2010-10.kml", points)
	readAndAppendData("/home/mrjones/src/latvis/data/2010-11.kml", points)
	readAndAppendData("/home/mrjones/src/latvis/data/2010-12.kml", points)
	readAndAppendData("/home/mrjones/src/latvis/data/jan2011.kml", points)
	readAndAppendData("/home/mrjones/src/latvis/data/feb2011.kml", points)

	mapdata := generateHeatmap(points, size)

	img := imageOfHeatmap(mapdata, size)
	renderImage(img, "vis.png")

}
