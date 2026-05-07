package main

import "slices"

type Person struct {
	Name        string
	Positions   [][2]float32
	Connections []Connection
}

type Connection struct {
	Person   *Person
	Strength int
}

func (p *Person) connect(other *Person, strength int) {
	p.Connections = append(p.Connections, Connection{Person: other, Strength: strength})
}

func (p *Person) unconnect(other *Person) {
	p.Connections = slices.DeleteFunc(p.Connections, func(c Connection) bool {
		return c.Person == other
	})
}

func (p *Person) isConnectedTo(other *Person) bool {
	for _, conn := range p.Connections {
		if conn.Person == other {
			return true
		}
	}
	return false
}

func (p *Person) toSaved() savedPerson {
	connections := make([]savedConnection, len(p.Connections))
	for i, conn := range p.Connections {
		connections[i] = savedConnection{
			Person:   conn.Person.Name,
			Strength: conn.Strength,
		}
	}
	return savedPerson{
		Name:        p.Name,
		Positions:   p.Positions,
		Connections: connections,
	}
}

type savedPerson struct {
	Name        string            `json:"name"`
	Positions   [][2]float32      `json:"positions"`
	Connections []savedConnection `json:"connections"`
}

type savedConnection struct {
	Person   string `json:"person"`
	Strength int    `json:"strength"`
}

func LoadSavedConnections(people []Person, saved []savedPerson) {
	for i, s := range saved {
		for _, sConn := range s.Connections {
			var other *Person
			for pi := range people {
				p := &people[pi]
				if p.Name == sConn.Person {
					other = p
					break
				}
			}
			if other == nil {
				panic("there is a guy connected with a nonexistent guy")
			}
			conn := Connection{
				Person:   other,
				Strength: sConn.Strength,
			}
			people[i].Connections = append(people[i].Connections, conn)
		}
	}
}
