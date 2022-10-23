package todometrics

import (
	"sort"
	"time"

	"ToDoInfo/internal/todo"
)

func getRottennessAge(r TaskRottenness) int {
	switch r {
	case ZombieTaskRottenness:
		return zombieTaskDay
	case TiredTaskRottenness:
		return tiredTaskDay
	case RipeTaskRottenness:
		return ripeTaskDay
	}
	return 0
}

func getTaskRottenness(age int) TaskRottenness {
	if age <= ripeTaskDay {
		return FreshTaskRottenness
	} else if age <= tiredTaskDay {
		return RipeTaskRottenness
	} else if age <= zombieTaskDay {
		return TiredTaskRottenness
	}
	return ZombieTaskRottenness
}

func getTaskAge(task todo.Task) int {
	taskTime := task.CreatedDateTime
	if task.Title == "Daily" || task.Title == "Weekly" {
		taskTime = task.LastModifiedDateTime
	}
	currentTime := time.Now()
	delta := currentTime.Sub(taskTime)
	return int(delta.Hours() / 24)
}

func sortTasks(taskLists []todo.TaskList) []TaskRottennessInfo {
	result := make([]TaskRottennessInfo, 0)
	for _, taskList := range taskLists {
		if taskList.WellknownListName == "defaultList" {
			continue
		}

		for _, task := range taskList.Tasks {
			age := getTaskAge(task)
			result = append(result, TaskRottennessInfo{
				TaskName:   task.Title,
				TaskList:   taskList.Name,
				Age:        age,
				Rottenness: getTaskRottenness(age),
			})
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Age > result[j].Age
	})

	return result
}
