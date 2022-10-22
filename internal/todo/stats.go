package todo

import (
	"sort"
	"time"
)

const (
	zombieTaskDay = 14
	tiredTaskDay  = 7
	ripeTaskDay   = 3
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

func getTaskAge(task Task) int {
	taskTime := task.CreatedDateTime
	if task.Title == "Daily" || task.Title == "Weekly" {
		taskTime = task.LastModifiedDateTime
	}
	currentTime := time.Now()
	delta := currentTime.Sub(taskTime)
	return int(delta.Hours() / 24)
}

func (l *TaskLists) sortTasks() []TaskRottennessInfo {
	result := make([]TaskRottennessInfo, 0)
	for _, taskList := range l.Lists {
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

func (l *TaskLists) GetListAges() ListAges {
	sum := 0
	ages := make(map[string]int)

	for _, taskList := range l.Lists {
		if taskList.WellknownListName == "defaultList" {
			continue
		}

		ages[taskList.Name] = 0
		for _, task := range taskList.Tasks {
			age := getTaskAge(task)
			sum += age
			ages[taskList.Name] += age
		}
	}

	listAges := ListAges{TotalAge: sum}
	for listName, listAge := range ages {
		listAges.Ages = append(listAges.Ages, ListAge{Title: listName, Age: listAge})
	}

	sort.Slice(listAges.Ages, func(i, j int) bool {
		return listAges.Ages[i].Age > listAges.Ages[j].Age
	})

	return listAges
}

func (l *TaskLists) GetTopOldestTasks(n int) []TaskRottennessInfo {
	tasks := l.sortTasks()
	return tasks[:n]
}

func (l *TaskLists) GetOldestTaskForList() map[string]string {
	result := make(map[string]string)
	for _, taskList := range l.Lists {
		if len(taskList.Tasks) == 0 {
			result[taskList.Name] = ""
		}

		maxAge := 0
		oldestTask := ""
		for _, task := range taskList.Tasks {
			age := getTaskAge(task)
			if age >= maxAge {
				maxAge = age
				oldestTask = task.Title
			}
		}
		result[taskList.Name] = oldestTask
	}
	return result
}

func (l *TaskLists) GetRottenTasks(minLevel TaskRottenness) []TaskRottennessInfo {
	tasks := l.sortTasks()
	if minLevel == FreshTaskRottenness {
		return tasks
	}

	minAge := getRottennessAge(minLevel)
	for i, task := range tasks {
		if task.Age <= minAge {
			return tasks[:i]
		}
	}

	return nil
}
