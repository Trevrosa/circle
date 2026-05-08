package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	people, err := loadPeople()
	if err != nil {
		panic(err)
	}

	ebiten.SetWindowTitle("Circle")
	ebiten.SetWindowSize(WIDTH, HEIGHT)
	ebiten.SetVsyncEnabled(true)

	pagePeopleCount = len(people)
	if err := ebiten.RunGame(NewWindow(people)); err != nil {
		log.Fatal(err)
	}

	if err := savePeople(people); err != nil {
		log.Fatal(err)
	}
}

func savePeople(people []Person) error {
	toSave := make([]savedPerson, len(people))
	for i, p := range people {
		toSave[i] = p.toSaved()
	}
	data, err := json.MarshalIndent(toSave, "", "  ")
	if err != nil {
		return err
	}
	fmt.Printf("saved %v bytes", len(data))
	return os.WriteFile("people.json", data, 0644)
}

func loadPeople() ([]Person, error) {
	type migratablePerson struct {
		Name        string
		Positions   [][2]float32
		Position    [2]float32
		Connections []savedConnection
	}

	if file, err := os.ReadFile("people.json"); err == nil {
		var migratable []migratablePerson
		if err := json.Unmarshal(file, &migratable); err != nil {
			return nil, err
		}

		hasMigrated := false
		savedPeople := make([]savedPerson, len(migratable))
		for i, m := range migratable {
			savedPeople[i] = savedPerson{
				Name:        m.Name,
				Connections: m.Connections,
			}
			if m.Position != [2]float32{} {
				hasMigrated = true
				savedPeople[i].Positions = append(savedPeople[i].Positions, m.Position)
			} else {
				savedPeople[i].Positions = m.Positions
			}
		}
		if hasMigrated {
			fmt.Println("need to migrate people, making backup")
			if err := os.WriteFile("people.json.bak", file, 0644); err != nil {
				panic(err)
			}
			data, err := json.MarshalIndent(savedPeople, "", "  ")
			if err != nil {
				panic(err)
			}
			if err := os.WriteFile("people.json", data, 0644); err != nil {
				panic(err)
			}
			fmt.Println("updated people.json")
		}

		people := make([]Person, len(savedPeople))
		for i, s := range savedPeople {
			people[i] = Person{
				Name:      s.Name,
				Positions: s.Positions,
			}
		}

		LoadSavedConnections(people, savedPeople)

		fmt.Printf("loaded %v people from people.json\n", len(people))
		return people, nil
	}

	file, err := os.ReadFile("people.txt")
	if err != nil {
		return nil, err
	}

	var people []Person
	for person := range strings.SplitSeq(string(file), ",") {
		person = strings.Split(strings.TrimSpace(person), " ")[0]
		if person != "" {
			people = append(people, Person{Name: person, Positions: [][2]float32{{0, 0}}})
		}
	}
	return people, nil
}
