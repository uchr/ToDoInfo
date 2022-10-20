package main

import (
	"ToDoInfo/internal/todo"
)

type ListDaysum struct {
	Title  string
	Daysum int
}

type TaskPageData struct {
	TaskName       string
	TaskList       string
	Age            int
	TaskRottenness string
}

type PageData struct {
	OverallDaysum int
	ListDaysums   []ListDaysum

	OldestTasks []TaskPageData
	RottenTasks []TaskPageData
}

func GetPageData(daysums todo.Daysums, oldestTasks []todo.TaskRottennessInfo, rottenTasks []todo.TaskRottennessInfo) PageData {
	pageData := PageData{}

	pageData.OverallDaysum = daysums.Overall
	for listName, listDaysum := range daysums.ListDaysums {
		pageData.ListDaysums = append(pageData.ListDaysums, ListDaysum{listName, listDaysum})
	}

	for _, task := range oldestTasks {
		pageData.OldestTasks = append(pageData.OldestTasks, TaskPageData{
			TaskName:       task.TaskName,
			TaskList:       task.TaskList,
			Age:            task.Age,
			TaskRottenness: task.Rottenness.String(),
		})
	}

	for _, task := range rottenTasks {
		pageData.RottenTasks = append(pageData.RottenTasks, TaskPageData{
			TaskName:       task.TaskName,
			TaskList:       task.TaskList,
			Age:            task.Age,
			TaskRottenness: task.Rottenness.String(),
		})
	}

	return pageData
}
