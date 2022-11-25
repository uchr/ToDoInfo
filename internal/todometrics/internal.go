package todometrics

import (
	"sort"
	"strings"
	"time"

	"github.com/uchr/ToDoInfo/internal/todo"
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
	if task.Recurrence != nil {
		taskTime = task.LastModifiedDateTime
	}
	currentTime := time.Now()
	delta := currentTime.Sub(taskTime)
	return int(delta.Hours() / 24)
}

func getSortedTasks(taskLists []todo.TaskList) []TaskRottennessInfo {
	result := make([]TaskRottennessInfo, 0)
	for _, taskList := range taskLists {
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

func filterTasks(taskLists []todo.TaskList) []todo.TaskList {
	filteredTaskList := make([]todo.TaskList, 0, len(taskLists))

	for _, taskList := range taskLists {
		tasks := make([]todo.Task, 0, len(taskList.Tasks))
		for _, task := range taskList.Tasks {
			if strings.Contains(task.Body.Content, "#todo-info-skip") {
				continue
			}
			tasks = append(tasks, task)
		}

		filteredTaskList = append(filteredTaskList, todo.TaskList{
			Name:              taskList.Name,
			WellknownListName: taskList.WellknownListName,
			IsShared:          taskList.IsShared,
			Tasks:             tasks,
		})
	}

	return filteredTaskList
}
