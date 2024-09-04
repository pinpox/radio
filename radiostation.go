package main

import (
	"log"
)

type RadioStation struct {
	Url   string
	StationName  string
	StationTitle string
}

func (rs *RadioStation) Update() {
	if title, err := rs.GetStreamTitle(); err != nil {
		log.Println(err)
		return
	} else {
		rs.StationTitle = title
	}
}

type RadioStations []RadioStation

func (s *RadioStations) Add(name, url string) {
	log.Printf("Adding station: %s (%s)", name, url)
	*s = append(*s, RadioStation{
		Url:   url,
		StationName:  name,
		StationTitle: "",
	})
}
