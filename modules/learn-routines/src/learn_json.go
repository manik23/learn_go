package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

type Event struct {
	Name      string          `json:"name,omitempty"`      // Omitted if empty string or nil
	Year      int             `json:"year,omitzero"`       // Omitted if value is 0
	CreatedAt time.Time       `json:"created_at,omitzero"` // Omitted if the zero time value
	Test      struct{ A int } `json:"test,omitzero"`
}

type Name struct {
	First string `json:"first"`
	Last  string `json:"last"`
}

type UserInlined struct {
	ID   int              `json:"id"`
	Name `json:",inline"` // Fields of Name are inlined into User
	Role string           `json:"role"`
}

func learnJSON() {
	// A struct where Year and CreatedAt have their zero values
	e := Event{
		Name:      "Conference",
		Year:      0,
		CreatedAt: time.Time{}, // zero value of time.Time
	}

	jsonData, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		fmt.Println(err)
	}

	// Output will be: {"name":"Conference"}
	// "year" and "created_at" are omitted due to "omitzero"
	fmt.Printf("%s\n", jsonData)

	u := UserInlined{
		ID: 1,
		Name: Name{
			First: "manik",
			Last:  "mahajan",
		},
		Role: "monitor",
	}

	userData, _ := json.MarshalIndent(u, "", "  ")
	log.Println(string(userData))
}
