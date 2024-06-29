package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

	_ "github.com/lib/pq"
)

type Film struct {
	Title    string
	Director string
}

type Todo struct {
	Description string
	IsDone      bool
}

func main() {
	// Database
	connStr := "postgres://postgres:password@localhost:5432/todo_app?sslmode=disable"

	db, err := sql.Open("postgres", connStr)

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	dropTodoTable(db)
	createTodoTable(db)

	todo := Todo{
		"Wash my socks",
		false,
	}

	newId := insertTodo(db, todo)

	selectTodo(db, newId)

	selectDoneTodos(db)

	// Server
	port := 8000

	// handler function #1 - returns the index.html template, with film data
	h1 := func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("index.html"))
		films := map[string][]Film{
			"Films": {
				{Title: "The Godfather", Director: "Francis Ford Coppola"},
				{Title: "Blade Runner", Director: "Ridley Scott"},
				{Title: "The Thing", Director: "John Carpenter"},
			},
		}
		tmpl.Execute(w, films)
	}

	// handler function #2 - returns the template block with the newly added film, as an HTMX response
	h2 := func(w http.ResponseWriter, r *http.Request) {
		title := r.PostFormValue("title")
		director := r.PostFormValue("director")
		tmpl := template.Must(template.ParseFiles("index.html"))
		tmpl.ExecuteTemplate(w, "film-list-element", Film{Title: title, Director: director})
	}

	// define handlers
	http.HandleFunc("/", h1)
	http.HandleFunc("/add-film/", h2)

	fmt.Println("Listening on port " + strconv.Itoa(port) + "...")
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), nil))

}

func dropTodoTable(db *sql.DB) {
	query := `DROP TABLE IF EXISTS todo`

	_, err := db.Exec(query)

	if err != nil {
		log.Fatal(err)
	}
}

func dropTodoAppTable(db *sql.DB) {
	query := `DROP TABLE IF EXISTS todo_app`

	_, err := db.Exec(query)

	if err != nil {
		log.Fatal(err)
	}
}

func createTodoTable(db *sql.DB) {
	query := `CREATE TABLE IF NOT EXISTS todo (
		id SERIAL PRIMARY KEY,
		description VARCHAR(200) NOT NULL,
		is_done BOOLEAN DEFAULT FALSE,
		created timestamp DEFAULT NOW()
	)`

	_, err := db.Exec(query)

	if err != nil {
		log.Fatal(err)
	}
}

func insertTodo(db *sql.DB, todo Todo) int {
	query := `INSERT INTO todo (description, is_done)
		VALUES ($1, $2) RETURNING id`

	var newId int
	err := db.QueryRow(query, todo.Description, todo.IsDone).Scan(&newId)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("id: %d\n", newId)

	return newId
}

func selectTodo(db *sql.DB, id int) {
	var description string
	var isDone bool

	query := " SELECT description, is_done FROM todo WHERE id = $1"
	err := db.QueryRow(query, id).Scan(&description, &isDone)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Fatalf("No rows found with id %d", id)
		}

		log.Fatal(err)
	}

	fmt.Printf("description: %s\n", description)
	fmt.Printf("isDone: %v\n", isDone)
}

func selectDoneTodos(db *sql.DB) {
	data := []Todo{}
	rows, err := db.Query("SELECT description, is_done FROM todo WHERE is_done = 't'")

	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()

	var description string
	var isDone bool

	for rows.Next() {
		err := rows.Scan(&description, &isDone)

		if err != nil {
			log.Fatal(err)
		}

		data = append(data, Todo{description, isDone})
	}

	fmt.Println(data)
}
