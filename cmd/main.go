package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

var config Config

type Name struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type Config struct {
	Names []string `json:"names"`
}

func main() {
	log.Println("Starting Name Server")
	load()
	http.Handle("/", http.FileServer(http.Dir("./assets")))

	http.HandleFunc("/save", handleSave)
	http.HandleFunc("/names", handleNames)

	log.Println("Serving on http://localhost:8766")
	http.ListenAndServe(":8766", nil)
}

func load() {
	var names []Name
	b1, err := os.ReadFile("./data/names.json")
	if err != nil {
		log.Println(fmt.Errorf("warn: %v", err))
	} else {
		json.Unmarshal(b1, &names)
	}
	if len(names) == 0 {
		log.Println("No entries found, loading defaults")
		b2, err := os.ReadFile("./config/defaults.json")
		if err != nil {
			log.Fatalf("Failed to open configs: %v", err)
		}
		err = json.Unmarshal(b2, &config)
		if err != nil {
			log.Fatalf("Failed to parse configs: %v", err)
		}
		for _, n := range config.Names {
			fmt.Println(n)
			err := saveName(n)
			if err != nil {
				panic(err)
			}
		}
	} else {
		log.Println("Entries found, do nothing")
	}
}

func handleSave(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()

	if err != nil {
		fmt.Println(fmt.Errorf("error: %v", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = saveName(r.Form.Get("name"))
	if err != nil {
		fmt.Println(fmt.Errorf("error: %v", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func saveName(name string) error {
	var names []Name
	var id int
	b, err := os.ReadFile("./data/names.json")
	if err != nil {
		fmt.Println(fmt.Errorf("warn: %v", err))
	} else {
		json.Unmarshal(b, &names)
	}
	if len(names) == 0 {
		id = 1
	} else {
		id = names[len(names)-1].Id + 1
	}

	names = append(names, Name{
		Name: name,
		Id:   id,
	})

	bytes, err := json.Marshal(names)
	if err != nil {
		return err
	}
	return os.WriteFile("./data/names.json", bytes, os.ModeAppend)
}

func handleNames(w http.ResponseWriter, r *http.Request) {
	b, _ := os.ReadFile("./data/names.json")
	w.Write(b)
}
