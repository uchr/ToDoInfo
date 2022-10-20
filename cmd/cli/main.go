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

	daysums := taskLists.GetDaysum()
	fmt.Println("# Daysums")
	fmt.Println("Overall:", daysums.Overall, "days")
	for listName, listDaysum := range daysums.ListDaysums {
		fmt.Println(fmt.Sprintf(" - %s: %d days", listName, listDaysum))
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
