package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	_ "github.com/lib/pq"
)

type Person struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

var db *sql.DB
var err error

// Initialize the database connection
func initDB() {
	connStr := "host=localhost port=5433 user=postgres password=postgres dbname=mydb sslmode=disable"
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to the database!")
}

// Create a new person
func createPerson(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var person Person
	_ = json.NewDecoder(r.Body).Decode(&person)

	fmt.Println(person)

	sqlStatement := `INSERT INTO people (name, email) VALUES ($1, $2) RETURNING id`
	err := db.QueryRow(sqlStatement, person.Name, person.Email).Scan(&person.ID)
	if err != nil {
		http.Error(w, "Failed to execute the query", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(person)
}

// Get all people
func getPeople(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rows, err := db.Query("SELECT * FROM people")
	if err != nil {
		http.Error(w, "Failed to execute the query", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var people []Person
	for rows.Next() {
		var person Person
		err = rows.Scan(&person.ID, &person.Name, &person.Email)
		if err != nil {
			http.Error(w, "Failed to scan the row", http.StatusInternalServerError)
			return
		}
		people = append(people, person)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(people)
}

// Get a person by ID
func getPerson(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Path[len("/people/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var person Person
	sqlStatement := `SELECT * FROM people WHERE id=$1`
	row := db.QueryRow(sqlStatement, id)
	err = row.Scan(&person.ID, &person.Name, &person.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Person not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to scan the row", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(person)
}

// Update a person
func updatePerson(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Path[len("/people/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var person Person
	_ = json.NewDecoder(r.Body).Decode(&person)

	sqlStatement := `UPDATE people SET name=$1, email=$2 WHERE id=$3`
	_, err = db.Exec(sqlStatement, person.Name, person.Email, id)
	if err != nil {
		http.Error(w, "Failed to execute the query", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(person)
}

// Delete a person
func deletePerson(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Path[len("/people/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	sqlStatement := `DELETE FROM people WHERE id=$1`
	_, err = db.Exec(sqlStatement, id)
	if err != nil {
		http.Error(w, "Failed to execute the query", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func main() {
	initDB()
	defer db.Close()

	// Create a new ServeMux (default Go router)
	http.HandleFunc("/people", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/people" {
			switch r.Method {
			case http.MethodPost:
				createPerson(w, r)
			case http.MethodGet:
				getPeople(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else {
			http.NotFound(w, r)
		}
	})

	// Route to get, update, or delete a person by ID
	http.HandleFunc("/people/", func(w http.ResponseWriter, r *http.Request) {
		if len(r.URL.Path) > len("/people/") {
			switch r.Method {
			case http.MethodGet:
				getPerson(w, r)
			case http.MethodPut:
				updatePerson(w, r)
			case http.MethodDelete:
				deletePerson(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else {
			http.NotFound(w, r)
		}
	})

	fmt.Println("Server is listening on port 5000...")
	log.Fatal(http.ListenAndServe(":5000", nil))
}
