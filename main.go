package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"slices"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
)

type Person struct {
	Name        string `json:"name"`
	Position    [2]float32 `json:"position"`
	Connections []Connection `json:"connections"`
}

func (p *Person) connect(other *Person, strength int) {
	p.Connections = append(p.Connections, Connection{Person: other.Name, Strength: strength})
}

func (p *Person) unconnect(other *Person) {
	p.Connections = slices.DeleteFunc(p.Connections, func(c Connection) bool {
		return c.Person == other.Name
	})
}

func (p *Person) isConnectedTo(other *Person) bool {
	for _, conn := range p.Connections {
		if conn.Person == other.Name {
			return true
		}
	}
	return false
}

type Connection struct {
	Person   string `json:"person"`
	Strength int `json:"strength"`
} // create new savedConnection with the string and this one uses pointers

func main() {
	people, err := loadPeople()
	if err != nil {
		panic(err)
	}

	Init()
	ebiten.SetWindowSize(WIDTH, HEIGHT)
	ebiten.SetWindowTitle("Circle")
	if err := ebiten.RunGame(&Window{People: people, draggingIndex: -1, connStartIndex: -1}); err != nil {
		log.Fatal(err)
	}
	
	if err := savePeople(people); err != nil {
		log.Fatal(err)
	}
}

func savePeople(people []Person) error {
	data, err := json.MarshalIndent(people, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile("people.json", data, 0644)
}

func loadPeople() ([]Person, error) {
	if file, err := os.ReadFile("people.json"); err == nil {
		var people []Person
		if err := json.Unmarshal(file, &people); err != nil {
			return nil, err
		}
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