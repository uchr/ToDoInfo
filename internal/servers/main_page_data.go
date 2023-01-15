package servers

import (
	"github.com/uchr/ToDoInfo/internal/todometrics"
)

type ListAge struct {
	Title         string
	Age           int
	OldestTask    string
	OldestTaskAge int
}

type Task struct {
	Task  todometrics.TaskRottennessInfo
	Emoji string
}

type PageData struct {
	TotalAge int
	ListAges []ListAge

	RottenTasks   []Task
	UpcomingTasks []Task
}

func NewPageData(metrics *todometrics.Metrics, fullTaskList bool) PageData {
	const numberOfUpcomingTasks = 3
	pageData := PageData{}

	listAges := metrics.GetListAges()
	oldestTaskForLists := metrics.GetOldestTaskForList()

	rottenTasks := metrics.GetRottenTasks(todometrics.TiredTaskRottenness)

	tasks := metrics.GetSortedTasks()
	var upcomingTasks []todometrics.TaskRottennessInfo
	if fullTaskList {
		upcomingTasks = tasks
	} else {
		upcomingTasks = tasks[len(rottenTasks) : len(rottenTasks)+numberOfUpcomingTasks]
	}

	pageData.TotalAge = listAges.TotalAge
	for _, listAge := range listAges.Ages {
		pageData.ListAges = append(pageData.ListAges, ListAge{
			Title:         listAge.Title,
			Age:           listAge.Age,
			OldestTask:    oldestTaskForLists[listAge.Title].TaskName,
			OldestTaskAge: oldestTaskForLists[listAge.Title].Age,
		})
	}

	for _, task := range upcomingTasks {
		pageData.UpcomingTasks = append(pageData.UpcomingTasks, Task{
			Task:  task,
			Emoji: task.Rottenness.String(),
		})
	}

	for _, task := range rottenTasks {
		pageData.RottenTasks = append(pageData.RottenTasks, Task{
			Task:  task,
			Emoji: task.Rottenness.String(),
		})
	}

	return pageData
}
