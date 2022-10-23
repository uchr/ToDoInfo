package main

import (
	"ToDoInfo/internal/todo"
	"time"
)

type ExampleTasks struct {
}

func getDate(dayBefore int) time.Time {
	d := time.Duration(int(time.Hour) * -24 * dayBefore)
	return time.Now().Add(d)
}

func (e ExampleTasks) GetTasks(token string) (*todo.TaskLists, error) {
	t := todo.TaskLists{Lists: []todo.TaskList{
		{
			Name: "🏠Home",
			Tasks: []todo.Task{
				{
					Title:                "Take out the trash",
					CreatedDateTime:      getDate(4),
					LastModifiedDateTime: getDate(4),
				},
				{
					Title:                "Hang a painting",
					CreatedDateTime:      getDate(20),
					LastModifiedDateTime: getDate(20),
				},
			},
		},
		{
			Name: "🎒Hiking",
			Tasks: []todo.Task{
				{
					Title:                "Check with John for an itinerary",
					CreatedDateTime:      getDate(8),
					LastModifiedDateTime: getDate(8),
				},
				{
					Title:                "Pack a bag",
					CreatedDateTime:      getDate(1),
					LastModifiedDateTime: getDate(1),
				},
			},
		},
		{
			Name: "💻Work",
			Tasks: []todo.Task{
				{
					Title:                "Answer the letter from Bob",
					CreatedDateTime:      getDate(10),
					LastModifiedDateTime: getDate(10),
				},
			},
		},
	}}

	return &t, nil
}
