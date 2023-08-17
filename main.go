package main

import (
	"fmt"
	"flag"
	"net/http"
	"strconv"
	"strings"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

var db, db_err = sql.Open("sqlite3", "foo.db")

var html_templates = make(map[string]HTMLTemplate)

func LoadTemplates() {
	html_templates["task"] = GenerateTemplate("templates/task.html")
	html_templates["default"] = GenerateTemplate("templates/index.html")
}

var static_content = make(map[string]StaticResource)

func LoadStaticFiles() {
	static_content["htmx"] = LoadStaticFile("static/htmx.min.js")
	static_content["css"] = LoadStaticFile("static/style.css")
	static_content["hello"] = LoadStaticFile("static/hello.html")
	static_content[""] = LoadStaticFile("static/index.html")

	for path, resource := range static_content {http.HandleFunc("/"+path, getBindResource(resource))}
}

type DefaultPageData struct {
	Path string
}

type T_task struct {
	Id int
	Task string
	Created string
	Completed string
}

func respondTaskQuery(w http.ResponseWriter, rows *sql.Rows) {
	var task T_task
	for rows.Next() {
		err := rows.Scan(&task.Id, &task.Task, &task.Created, &task.Completed)
		if err != nil {
			fmt.Println(err)
		} else {
			task.Created = strings.ReplaceAll(task.Created, "T", " ")
			task.Created = strings.ReplaceAll(task.Created, "Z", " ")
			html_templates["task"](task, w)
		}
	}
}

func getTasks(w http.ResponseWriter, r *http.Request) {
	sparams := strings.TrimSpace(r.FormValue("search"))
	if len(sparams) > 0 {
		q, err := db.Prepare("SELECT * FROM tasks WHERE task LIKE '%' || ? || '%' ORDER BY id DESC;")
		if err != nil {
			fmt.Println(err)
			fmt.Fprintf(w, "<h2>Currently no tasks</h2>")
		} else {
			rows, err := q.Query(sparams)
			if err != nil {
				fmt.Println(err)
				fmt.Fprintf(w, "<h2>Currently no tasks</h2>")
			} else {
				respondTaskQuery(w, rows)
			}
		}
	} else {
		rows,err := db.Query("SELECT * FROM tasks WHERE completed='false' ORDER BY id DESC;")
		defer rows.Close()
		if err != nil {
			fmt.Println(err)
			fmt.Fprintf(w, "<h2>Currenttly no tasks</h2>")
		} else {
			respondTaskQuery(w, rows)
		}

	}
}

func postCreateTask(w http.ResponseWriter, r *http.Request) {
	tsk := strings.TrimSpace(r.FormValue("task"))
	if len(tsk) != 0 {
		q, err := db.Prepare("INSERT INTO tasks(task, completed) VALUES(?,?);")
		if err != nil {fmt.Println(err)}
		_, err = q.Exec(tsk, "false")
		if err != nil {fmt.Println(err)}
	}
	getTasks(w,r)
}

func postCompleteTask(w http.ResponseWriter, r *http.Request) {
	q, err := db.Prepare("UPDATE tasks SET completed = CURRENT_TIMESTAMP WHERE completed ='false' AND id = ?")
	if err != nil {fmt.Println(err)}
	_, err = q.Exec(r.FormValue("task_id"))
	if err != nil {fmt.Println(err)}
	fmt.Fprintf(w, "")
}

func main() {
	var port int
	flag.IntVar(&port,"p",8080,"port to run on")
	flag.Parse()

	LoadTemplates()
	LoadStaticFiles()
	
	// create tables if inexistent
	if db_err != nil {fmt.Println(db_err)}
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS tasks(id INTEGER PRIMARY KEY AUTOINCREMENT, task, created DATETIME DEFAULT CURRENT_TIMESTAMP, completed);")
	if err != nil {fmt.Println(err)}
	defer db.Close()

	fmt.Println("Server starting!")
	http.HandleFunc("/create_task", postCreateTask)
	http.HandleFunc("/complete_task", postCompleteTask)
	http.HandleFunc("/tasks", getTasks)
	http.ListenAndServe(":"+strconv.Itoa(port), nil)
}
