package main

import (
	"fmt"
	"flag"
	"net/http"
	"html/template"
	"strconv"
	"os"
	"strings"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

var db, db_err = sql.Open("sqlite3", "foo.db")

type HTMLTemplate func(any, http.ResponseWriter)

func GenerateTemplate(f string) (HTMLTemplate) {
	temp, err := template.ParseFiles(f)
	if err != nil {fmt.Println(err)}
	return func(s any, r http.ResponseWriter) { 
		err := temp.Execute(r, s)
		if err != nil {fmt.Println(err)}
	}
}
var html_templates = make(map[string]HTMLTemplate)

func LoadTemplates() {
	html_templates["task"] = GenerateTemplate("templates/task.html")
	html_templates["default"] = GenerateTemplate("templates/index.html")
}

type StaticResource struct {
	Content string
	MimeType string
}

var static_content = make(map[string]StaticResource)

func getMimeType(f string) string {
	if strings.HasSuffix(f, ".css") {return "text/css"}
	if strings.HasSuffix(f, ".js") {return "text/javascript"}
	return "text/html"
}

func LoadStaticFile(f string) StaticResource {
	content, err := os.ReadFile(f)
	if err != nil {fmt.Println(err)}
	resource := StaticResource{Content: string(content[:]), MimeType: getMimeType(f)}
	return resource
}

type StaticBind func(http.ResponseWriter, *http.Request)

func getBindResource(res StaticResource) StaticBind {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", res.MimeType)
		fmt.Fprintf(w, res.Content)
	}	
}

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

func loadTasks(w http.ResponseWriter, r *http.Request) {
	rows,err := db.Query("SELECT * FROM tasks WHERE completed='false';")
	if err != nil {
		fmt.Println(err)
		fmt.Fprintf(w, "<h2>Currenttly no tasks</h2>")
	} else {
		var task T_task
		for rows.Next() {
			err = rows.Scan(&task.Id, &task.Task, &task.Created, &task.Completed)
			if err != nil {
				fmt.Println(err)
			} else {
				html_templates["task"](task, w)
			}
		}
	}
}

func postCreateTask(w http.ResponseWriter, r *http.Request) {
	// create task and return all active tasks
	q, err := db.Prepare("INSERT INTO tasks(task, completed) VALUES(?,?);")
	if err != nil {fmt.Println(err)}
	_, err = q.Exec(r.FormValue("task"), "false")
	if err != nil {fmt.Println(err)}

	loadTasks(w,r)
	/*
	rows, err := db.Query("SELECT * FROM tasks WHERE completed = 'false';")
	
	if err != nil {
		fmt.Println(err)
		fmt.Fprintf(w, "<h2>Currently no tasks</h2>")
	} else {
		var task T_task
		for rows.Next() {
			err = rows.Scan(&task.Id, &task.Task, &task.Created, &task.Completed)
			
			if err != nil {
				fmt.Println(err)
			} 
			html_templates["task"](task, w)
			
		}
	}*/
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
	http.HandleFunc("/tasks", loadTasks)
	http.ListenAndServe(":"+strconv.Itoa(port), nil)
}
