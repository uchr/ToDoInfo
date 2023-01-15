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

func getTaskAge(task todo.Task) (int, time.Duration) {
	taskTime := task.CreatedDateTime
	if task.DueDateTime != nil {
		taskTime = *task.DueDateTime
	}

	currentTime := time.Now()
	delta := currentTime.Sub(taskTime)
	if delta <= 0 {
		return 0, 0
	}
	return int(delta.Hours() / 24), delta
}

func getSortedTasks(taskLists []todo.TaskList) []TaskRottennessInfo {
	result := make([]TaskRottennessInfo, 0)
	for _, taskList := range taskLists {
		for _, task := range taskList.Tasks {
			age, exactAge := getTaskAge(task)
			result = append(result, TaskRottennessInfo{
				TaskName:   task.Title,
				TaskList:   taskList.Name,
				Age:        age,
				Rottenness: getTaskRottenness(age),

				exactAge: exactAge,
			})
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].exactAge > result[j].exactAge
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
