package main

import (
	"fmt"

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

	taskLists, err := todoparser.ParseTasks(token)
	if err != nil {
		log.Error(err)
		return
	}

	listAges := taskLists.GetListAges()
	fmt.Println("# List Ages")
	fmt.Println("Total age:", listAges.TotalAge, "days")
	for _, listAge := range listAges.Ages {
		fmt.Println(fmt.Sprintf(" - %s: %d days", listAge.Title, listAge.Age))
	}

	fmt.Println("# Oldest Tasks")
	oldestTasks := taskLists.GetTopOldestTasks(5)
	for _, task := range oldestTasks {
		fmt.Println(fmt.Sprintf(" - %s %s:%s (%d days)", task.Rottenness.String(), task.TaskList, task.TaskName, task.Age))
	}

	fmt.Println("# Tired Tasks")
	rottenTasks := taskLists.GetRottenTasks(todo.TiredTaskRottenness)
	for _, task := range rottenTasks {
		fmt.Println(fmt.Sprintf(" - %s %s:%s (%d days)", task.Rottenness.String(), task.TaskList, task.TaskName, task.Age))
	}
}
