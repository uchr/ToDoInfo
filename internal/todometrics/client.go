package todometrics

import (
	"sort"
	"time"

	"github.com/uchr/ToDoInfo/internal/todo"
)

func (r TaskRottenness) String() string {
	switch r {
	case ZombieTaskRottenness:
		return "🤢"
	case TiredTaskRottenness:
		return "🥱"
	case RipeTaskRottenness:
		return "😏"
	case FreshTaskRottenness:
		return "😊"
	}
	return "❓"
}

func New(taskLists []todo.TaskList) *Metrics {
	filteredTasks := filterTasks(taskLists)

	return &Metrics{lists: filteredTasks, sortedTasks: getSortedTasks(filteredTasks)}
}

func (l *Metrics) GetListAges() ListAges {
	sum := 0
	ages := make(map[string]int)
	taskCount := make(map[string]int)

	for _, taskList := range l.lists {
		ages[taskList.Name] = 0
		taskCount[taskList.Name] = len(taskList.Tasks)
		for _, task := range taskList.Tasks {
			age, _ := getTaskAge(task)
			sum += age
			ages[taskList.Name] += age
		}
	}

	listAges := ListAges{TotalAge: sum}
	for listName, listAge := range ages {
		listAges.Ages = append(listAges.Ages, ListAge{Title: listName, Age: listAge, TaskCount: taskCount[listName]})
	}

	sort.Slice(listAges.Ages, func(i, j int) bool {
		if listAges.Ages[i].Age == listAges.Ages[j].Age {
			return listAges.Ages[i].Title < listAges.Ages[j].Title
		}
		return listAges.Ages[i].Age > listAges.Ages[j].Age
	})

	return listAges
}

func (l *Metrics) GetTopTasksByAge(n int) []TaskRottennessInfo {
	if n >= len(l.sortedTasks) {
		n = len(l.sortedTasks)
	}

	return l.sortedTasks[:n]
}

func (l *Metrics) GetSortedTasks() []TaskRottennessInfo {
	return l.sortedTasks
}

func (l *Metrics) GetOldestTaskForList() map[string]TaskRottennessInfo {
	result := make(map[string]TaskRottennessInfo)
	for _, taskList := range l.lists {
		if len(taskList.Tasks) == 0 {
			result[taskList.Name] = TaskRottennessInfo{}
			continue
		}

		var maxExactAge time.Duration
		taskIndex := 0
		for i, task := range taskList.Tasks {
			_, exactAge := getTaskAge(task)
			if exactAge >= maxExactAge {
				maxExactAge = exactAge
				taskIndex = i
			}
		}
		taskAge, exactAge := getTaskAge(taskList.Tasks[taskIndex])
		result[taskList.Name] = TaskRottennessInfo{
			TaskName:   taskList.Tasks[taskIndex].Title,
			TaskList:   taskList.Name,
			Age:        taskAge,
			Rottenness: getTaskRottenness(taskAge),
			exactAge:   exactAge,
		}
	}
	return result
}

func (l *Metrics) GetRottenTasks(minLevel TaskRottenness) []TaskRottennessInfo {
	if minLevel == FreshTaskRottenness {
		return l.sortedTasks
	}

	minAge := getRottennessAge(minLevel)
	for i, task := range l.sortedTasks {
		if task.Age <= minAge {
			return l.sortedTasks[:i]
		}
	}

	return nil
}
