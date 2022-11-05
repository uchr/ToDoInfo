package servers

import (
	"github.com/uchr/ToDoInfo/internal/todometrics"
)

type ListAge struct {
	Title      string
	Age        int
	OldestTask string
}

type TaskPageData struct {
	TaskName       string
	TaskList       string
	Age            int
	TaskRottenness string
}

type PageData struct {
	TotalAge int
	ListAges []ListAge

	OldestTasks []TaskPageData
	RottenTasks []TaskPageData
}

func GetPageData(metrics *todometrics.Metrics) PageData {
	pageData := PageData{}

	listAges := metrics.GetListAges()
	oldestTasks := metrics.GetTopOldestTasks(5)
	rottenTasks := metrics.GetRottenTasks(todometrics.TiredTaskRottenness)
	oldestTaskForLists := metrics.GetOldestTaskForList()

	pageData.TotalAge = listAges.TotalAge
	for _, listAge := range listAges.Ages {
		pageData.ListAges = append(pageData.ListAges, ListAge{
			Title:      listAge.Title,
			Age:        listAge.Age,
			OldestTask: oldestTaskForLists[listAge.Title],
		})
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
