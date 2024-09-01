package main
import (
	"time"
	"log"
)

type RadioStationMetadata struct {
	Title   string
	Updated time.Time
}

type RadioStation struct {
	Url         string
	Name        string
	CurrentMeta *RadioStationMetadata
}

func (rs *RadioStation) Update() {

	if title, err := rs.GetStreamTitle(); err != nil {
		log.Println(err)
	} else {
		if rs.CurrentMeta.Title != title {
			rs.CurrentMeta.Title = title
			rs.CurrentMeta.Updated = time.Now()
		}
	}
}

type RadioStations []RadioStation

func (s *RadioStations) Add(name, url string) {
	log.Printf("Adding station: %s (%s)", name, url)
	*s = append(*s, RadioStation{
		Url:         url,
		Name:        name,
		CurrentMeta: &RadioStationMetadata{Title: "", Updated: time.Now()},
	})
}

func (s *RadioStations) Update() {
	for {
		for _, v := range *s {
			v.Update()
		}

		time.Sleep(time.Second * 5)
	}
}
