package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"os"

	"github.com/gocarina/gocsv"
	"github.com/uber/h3-go/v4"
	"github.com/urfave/cli/v3"
)

const (
	stateResolution   = 2
	zipCodeResolution = 6
	statesPath        = "./assets/states.xml"
	zipCodesPath      = "./assets/uszips.csv"
)

type States struct {
	XMLName xml.Name `xml:"states"`
	States  []State  `xml:"state"`
}

type USZip struct {
	Zip string  `csv:"zip"`
	Lat float64 `csv:"lat"`
	Lng float64 `csv:"lng"`
}

type State struct {
	XMLName xml.Name `xml:"state"`
	Name    string   `xml:"name,attr"`
	Colour  string   `xml:"colour,attr"`
	Points  []Point  `xml:"point"`
}

type Point struct {
	XMLName xml.Name `xml:"point"`
	Lat     float64  `xml:"lat,attr"`
	Lng     float64  `xml:"lng,attr"`
}

func (s *States) OneByName(name string) *State {
	for _, s := range s.States {
		if name == s.Name {
			return &s
		}
	}

	return nil
}

func (s *State) Polygon() h3.GeoPolygon {
	loop := make([]h3.LatLng, len(s.Points))
	for i, p := range s.Points {
		loop[i] = h3.LatLng{Lat: p.Lat, Lng: p.Lng}
	}

	return h3.GeoPolygon{GeoLoop: loop}
}

func (z *USZip) Cell(resolution int) h3.Cell {
	cell, err := h3.LatLngToCell(h3.LatLng{
		Lat: z.Lat,
		Lng: z.Lng,
	}, resolution)
	if err != nil {
		panic(err)
	}

	return cell
}

func zips() []*USZip {
	file, err := os.Open(zipCodesPath)
	if err != nil {
		panic("invalid file")
	}

	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}

	zips := []*USZip{}
	if err := gocsv.UnmarshalBytes(bytes, &zips); err != nil {
		panic(err)
	}

	return zips
}

func states() States {
	file, err := os.Open(statesPath)
	if err != nil {
		panic("invalid file")
	}

	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}

	var states States
	err = xml.Unmarshal(bytes, &states)
	if err != nil {
		panic(err)
	}

	return states
}

func main() {
	zips := zips()
	states := states()

	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "location",
				Value: "California",
				Usage: "zip or state",
			},
			&cli.IntFlag{
				Name:  "resolution",
				Value: 0,
				Usage: "h3 cell resolution",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			var id string
			if cmd.NArg() > 0 {
				id = cmd.Args().Get(0)
			}

			if cmd.String("location") == "state" {
				resolution := stateResolution
				if res := cmd.Int("resolution"); res > 0 {
					resolution = res
				}

				state := states.OneByName(id)
				if state == nil {
					panic("state not found")
				}

				poly := state.Polygon()
				cells, _ := poly.Cells(resolution)
				fmt.Printf("%v\n", cells)
			}

			if cmd.String("location") == "zip" {
				resolution := zipCodeResolution
				if res := cmd.Int("resolution"); res > 0 {
					resolution = res
				}
				var code *USZip
				for _, usz := range zips {
					if usz.Zip == id {
						code = usz
						break
					}
				}

				if code == nil {
					panic("zip code not found")
				} else {
					fmt.Printf("%v\n", code.Cell(resolution))
				}
			}

			return nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		panic(err)
	}
}
