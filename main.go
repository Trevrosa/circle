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

	Init()
	ebiten.SetWindowTitle("Circle")
	ebiten.SetWindowSize(WIDTH, HEIGHT)
	ebiten.SetVsyncEnabled(true)
	if err := ebiten.RunGame(&Window{People: people, draggingIndex: -1, connStartIndex: -1}); err != nil {
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
	if file, err := os.ReadFile("people.json"); err == nil {
		var savedPeople []savedPerson
		if err := json.Unmarshal(file, &savedPeople); err != nil {
			return nil, err
		}

		people := make([]Person, len(savedPeople))
		for i, s := range savedPeople {
			people[i] = Person{
				Name: s.Name,
				Position: s.Position,
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
			people = append(people, Person{Name: person})
		}
	}
	return people, nil
}