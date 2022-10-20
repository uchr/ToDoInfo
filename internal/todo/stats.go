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
	currentTime := time.Now()
	delta := currentTime.Sub(task.LastModifiedDateTime)
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

func (l *TaskLists) GetDaysum() Daysums {
	daysums := Daysums{}
	daysums.ListDaysums = make(map[string]int)

	for _, taskList := range l.Lists {
		if taskList.WellknownListName == "defaultList" {
			continue
		}

		daysums.ListDaysums[taskList.Name] = 0
		for _, task := range taskList.Tasks {
			age := getTaskAge(task)
			daysums.Overall += age
			daysums.ListDaysums[taskList.Name] += age
		}
	}

	return daysums
}

func (l *TaskLists) GetTopOldestTasks(n int) []TaskRottennessInfo {
	tasks := l.sortTasks()
	return tasks[:n]
}

func (l *TaskLists) GetRottennessTasks(minLevel TaskRottenness) []TaskRottennessInfo {
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
