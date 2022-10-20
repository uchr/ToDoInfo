package main

import (
	"html/template"
	"net/http"

	"ToDoInfo/internal/config"
	"ToDoInfo/internal/log"
	"ToDoInfo/internal/login"
	"ToDoInfo/internal/todo"
	"ToDoInfo/internal/todoparser"
)

func main() {
	log.Init(log.InfoLevel)
	cfg, err := config.New()
	if err != nil {
		log.Error(err)
		return
	}
	log.Debug(cfg.ClientId)

	token, err := login.Login(cfg.ClientId)
	if err != nil {
		log.Error(err)
		return
	}
	log.Debug(token)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("web/template/index.html")

		taskLists, err := todoparser.ParseTasks(token)
		if err != nil {
			log.Error(err)
		}

		daysums := taskLists.GetDaysum()
		oldestTasks := taskLists.GetTopOldestTasks(5)
		rottenTasks := taskLists.GetRottenTasks(todo.TiredTaskRottenness)

		pageData := GetPageData(daysums, oldestTasks, rottenTasks)

		err = tmpl.Execute(w, pageData)
		if err != nil {
			log.Error(err)
			return
		}
	})

	fs := http.FileServer(http.Dir("web/static/"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	err = http.ListenAndServe(":80", nil)
	if err != nil {
		log.Error(err)
		return
	}

}
