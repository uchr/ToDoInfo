package todometrics

import (
	"time"

	"github.com/uchr/ToDoInfo/internal/todo"
)

type TaskRottenness int

const (
	ZombieTaskRottenness = iota
	TiredTaskRottenness
	RipeTaskRottenness
	FreshTaskRottenness
)

const (
	zombieTaskDay = 14
	tiredTaskDay  = 7
	ripeTaskDay   = 3
)

type ListAge struct {
	Age       int
	TaskCount int
	Title     string
}

type ListAges struct {
	TotalAge int
	Ages     []ListAge
}

type TaskRottennessInfo struct {
	TaskName   string
	TaskList   string
	Age        int
	Rottenness TaskRottenness

	exactAge time.Duration
}

type Metrics struct {
	lists       []todo.TaskList
	sortedTasks []TaskRottennessInfo
}
