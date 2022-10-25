package todometrics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"ToDoInfo/internal/todo"
)

func getDateFromNow(dayBefore int, hourBefore int) time.Time {
	d := time.Duration(int(time.Hour) * -(24*dayBefore + hourBefore))
	return time.Now().Add(d)
}

func TestGetTaskAge(t *testing.T) {
	tests := []struct {
		name        string
		taskTime    time.Time
		expectedAge int
	}{
		{
			name:        "Current time",
			taskTime:    time.Now(),
			expectedAge: 0,
		},
		{
			name:        "Today",
			taskTime:    getDateFromNow(0, 5),
			expectedAge: 0,
		},
		{
			name:        "Yesterday",
			taskTime:    getDateFromNow(1, 1),
			expectedAge: 1,
		},
		{
			name:        "Many days ago",
			taskTime:    getDateFromNow(10, 1),
			expectedAge: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			age := getTaskAge(todo.Task{
				CreatedDateTime:      tt.taskTime,
				LastModifiedDateTime: tt.taskTime,
			})

			assert.Equal(t, tt.expectedAge, age)
		})
	}
}

func TestGetSortedTasks(t *testing.T) {
	tests := []struct {
		name           string
		taskLists      []todo.TaskList
		expectedResult []TaskRottennessInfo
	}{
		{
			name:           "Empty list",
			taskLists:      make([]todo.TaskList, 0),
			expectedResult: make([]TaskRottennessInfo, 0),
		},
		{
			name: "Empty lists",
			taskLists: []todo.TaskList{
				{
					Name:  "List1",
					Tasks: nil,
				},
				{
					Name:  "List2",
					Tasks: nil,
				},
			},
			expectedResult: make([]TaskRottennessInfo, 0),
		},
		{
			name: "Non empty list",
			taskLists: []todo.TaskList{
				{
					Name: "List1",
					Tasks: []todo.Task{
						{
							Title:                "Task1",
							CreatedDateTime:      getDateFromNow(1, 0),
							LastModifiedDateTime: getDateFromNow(1, 0),
						},
						{
							Title:                "Task2",
							CreatedDateTime:      getDateFromNow(10, 0),
							LastModifiedDateTime: getDateFromNow(10, 0),
						},
					},
				},
				{
					Name: "List2",
					Tasks: []todo.Task{
						{
							Title:                "Task3",
							CreatedDateTime:      getDateFromNow(0, 5),
							LastModifiedDateTime: getDateFromNow(0, 5),
						},
						{
							Title:                "Task4",
							CreatedDateTime:      getDateFromNow(20, 0),
							LastModifiedDateTime: getDateFromNow(20, 0),
						},
						{
							Title:                "Task5",
							CreatedDateTime:      getDateFromNow(4, 0),
							LastModifiedDateTime: getDateFromNow(4, 0),
						},
					},
				},
			},
			expectedResult: []TaskRottennessInfo{
				{
					TaskName:   "Task4",
					TaskList:   "List2",
					Age:        20,
					Rottenness: ZombieTaskRottenness,
				},
				{
					TaskName:   "Task2",
					TaskList:   "List1",
					Age:        10,
					Rottenness: TiredTaskRottenness,
				},
				{
					TaskName:   "Task5",
					TaskList:   "List2",
					Age:        4,
					Rottenness: RipeTaskRottenness,
				},
				{
					TaskName:   "Task1",
					TaskList:   "List1",
					Age:        1,
					Rottenness: FreshTaskRottenness,
				},
				{
					TaskName:   "Task3",
					TaskList:   "List2",
					Age:        0,
					Rottenness: FreshTaskRottenness,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tasks := getSortedTasks(tt.taskLists)

			assert.Equal(t, tt.expectedResult, tasks)
		})
	}
}

func TestGetListAges(t *testing.T) {
	tests := []struct {
		name           string
		taskLists      []todo.TaskList
		expectedResult ListAges
	}{
		{
			name:      "Empty list",
			taskLists: make([]todo.TaskList, 0),
			expectedResult: ListAges{
				TotalAge: 0,
				Ages:     nil,
			},
		},
		{
			name: "Empty lists",
			taskLists: []todo.TaskList{
				{
					Name:  "List1",
					Tasks: nil,
				},
				{
					Name:  "List2",
					Tasks: nil,
				},
			},
			expectedResult: ListAges{
				TotalAge: 0,
				Ages: []ListAge{
					{
						Title: "List1",
						Age:   0,
					},
					{
						Title: "List2",
						Age:   0,
					},
				},
			},
		},
		{
			name: "Non empty list",
			taskLists: []todo.TaskList{
				{
					Name: "List1",
					Tasks: []todo.Task{
						{
							Title:                "Task1",
							CreatedDateTime:      getDateFromNow(1, 0),
							LastModifiedDateTime: getDateFromNow(1, 0),
						},
						{
							Title:                "Task2",
							CreatedDateTime:      getDateFromNow(10, 0),
							LastModifiedDateTime: getDateFromNow(10, 0),
						},
					},
				},
				{
					Name: "List2",
					Tasks: []todo.Task{
						{
							Title:                "Task3",
							CreatedDateTime:      getDateFromNow(0, 5),
							LastModifiedDateTime: getDateFromNow(0, 5),
						},
						{
							Title:                "Task4",
							CreatedDateTime:      getDateFromNow(20, 0),
							LastModifiedDateTime: getDateFromNow(20, 0),
						},
						{
							Title:                "Task5",
							CreatedDateTime:      getDateFromNow(4, 0),
							LastModifiedDateTime: getDateFromNow(4, 0),
						},
					},
				},
			},
			expectedResult: ListAges{
				TotalAge: 35,
				Ages: []ListAge{
					{
						Title: "List2",
						Age:   24,
					},
					{
						Title: "List1",
						Age:   11,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New(tt.taskLists)
			tasks := m.GetListAges()

			assert.Equal(t, tt.expectedResult, tasks)
		})
	}
}

func TestGetTopOldestTasks(t *testing.T) {
	tests := []struct {
		name           string
		taskCount      int
		taskLists      []todo.TaskList
		expectedResult []TaskRottennessInfo
	}{
		{
			name:           "Empty list",
			taskCount:      5,
			taskLists:      make([]todo.TaskList, 0),
			expectedResult: make([]TaskRottennessInfo, 0),
		},
		{
			name:           "Empty list. Requests 0 tasks",
			taskCount:      0,
			taskLists:      make([]todo.TaskList, 0),
			expectedResult: make([]TaskRottennessInfo, 0),
		},
		{
			name:      "Empty lists",
			taskCount: 5,
			taskLists: []todo.TaskList{
				{
					Name:  "List1",
					Tasks: nil,
				},
				{
					Name:  "List2",
					Tasks: nil,
				},
			},
			expectedResult: make([]TaskRottennessInfo, 0),
		},
		{
			name:      "Requests less then number of tasks",
			taskCount: 3,
			taskLists: []todo.TaskList{
				{
					Name: "List1",
					Tasks: []todo.Task{
						{
							Title:                "Task1",
							CreatedDateTime:      getDateFromNow(1, 0),
							LastModifiedDateTime: getDateFromNow(1, 0),
						},
						{
							Title:                "Task2",
							CreatedDateTime:      getDateFromNow(10, 0),
							LastModifiedDateTime: getDateFromNow(10, 0),
						},
					},
				},
				{
					Name: "List2",
					Tasks: []todo.Task{
						{
							Title:                "Task3",
							CreatedDateTime:      getDateFromNow(0, 5),
							LastModifiedDateTime: getDateFromNow(0, 5),
						},
						{
							Title:                "Task4",
							CreatedDateTime:      getDateFromNow(20, 0),
							LastModifiedDateTime: getDateFromNow(20, 0),
						},
						{
							Title:                "Task5",
							CreatedDateTime:      getDateFromNow(4, 0),
							LastModifiedDateTime: getDateFromNow(4, 0),
						},
					},
				},
			},
			expectedResult: []TaskRottennessInfo{
				{
					TaskName:   "Task4",
					TaskList:   "List2",
					Age:        20,
					Rottenness: ZombieTaskRottenness,
				},
				{
					TaskName:   "Task2",
					TaskList:   "List1",
					Age:        10,
					Rottenness: TiredTaskRottenness,
				},
				{
					TaskName:   "Task5",
					TaskList:   "List2",
					Age:        4,
					Rottenness: RipeTaskRottenness,
				},
			},
		},
		{
			name:      "Requests more then number of tasks",
			taskCount: 4,
			taskLists: []todo.TaskList{
				{
					Name: "List1",
					Tasks: []todo.Task{
						{
							Title:                "Task2",
							CreatedDateTime:      getDateFromNow(10, 0),
							LastModifiedDateTime: getDateFromNow(10, 0),
						},
					},
				},
				{
					Name: "List2",
					Tasks: []todo.Task{
						{
							Title:                "Task4",
							CreatedDateTime:      getDateFromNow(20, 0),
							LastModifiedDateTime: getDateFromNow(20, 0),
						},
					},
				},
			},
			expectedResult: []TaskRottennessInfo{
				{
					TaskName:   "Task4",
					TaskList:   "List2",
					Age:        20,
					Rottenness: ZombieTaskRottenness,
				},
				{
					TaskName:   "Task2",
					TaskList:   "List1",
					Age:        10,
					Rottenness: TiredTaskRottenness,
				},
			},
		},
		{
			name:      "Non empty list. Requests 0 tasks",
			taskCount: 0,
			taskLists: []todo.TaskList{
				{
					Name: "List1",
					Tasks: []todo.Task{
						{
							Title:                "Task2",
							CreatedDateTime:      getDateFromNow(10, 0),
							LastModifiedDateTime: getDateFromNow(10, 0),
						},
					},
				},
				{
					Name: "List2",
					Tasks: []todo.Task{
						{
							Title:                "Task4",
							CreatedDateTime:      getDateFromNow(20, 0),
							LastModifiedDateTime: getDateFromNow(20, 0),
						},
					},
				},
			},
			expectedResult: make([]TaskRottennessInfo, 0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New(tt.taskLists)
			tasks := m.GetTopOldestTasks(tt.taskCount)

			assert.Equal(t, tt.expectedResult, tasks)
		})
	}
}

func TestGetOldestTaskForList(t *testing.T) {
	tests := []struct {
		name           string
		taskLists      []todo.TaskList
		expectedResult map[string]string
	}{
		{
			name:           "Empty list",
			taskLists:      make([]todo.TaskList, 0),
			expectedResult: make(map[string]string, 0),
		},
		{
			name: "Empty lists",
			taskLists: []todo.TaskList{
				{
					Name:  "List1",
					Tasks: nil,
				},
				{
					Name:  "List2",
					Tasks: nil,
				},
			},
			expectedResult: map[string]string{
				"List1": "",
				"List2": "",
			},
		},
		{
			name: "Non empty list",
			taskLists: []todo.TaskList{
				{
					Name: "List1",
					Tasks: []todo.Task{
						{
							Title:                "Task1",
							CreatedDateTime:      getDateFromNow(1, 0),
							LastModifiedDateTime: getDateFromNow(1, 0),
						},
						{
							Title:                "Task2",
							CreatedDateTime:      getDateFromNow(10, 0),
							LastModifiedDateTime: getDateFromNow(10, 0),
						},
					},
				},
				{
					Name: "List2",
					Tasks: []todo.Task{
						{
							Title:                "Task3",
							CreatedDateTime:      getDateFromNow(0, 5),
							LastModifiedDateTime: getDateFromNow(0, 5),
						},
						{
							Title:                "Task4",
							CreatedDateTime:      getDateFromNow(20, 0),
							LastModifiedDateTime: getDateFromNow(20, 0),
						},
						{
							Title:                "Task5",
							CreatedDateTime:      getDateFromNow(4, 0),
							LastModifiedDateTime: getDateFromNow(4, 0),
						},
					},
				},
			},
			expectedResult: map[string]string{
				"List1": "Task2",
				"List2": "Task4",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New(tt.taskLists)
			tasks := m.GetOldestTaskForList()

			assert.Equal(t, tt.expectedResult, tasks)
		})
	}
}

func TestGetRottenTasks(t *testing.T) {
	tests := []struct {
		name           string
		minRottenness  TaskRottenness
		taskLists      []todo.TaskList
		expectedResult []TaskRottennessInfo
	}{
		{
			name:           "Empty list",
			minRottenness:  FreshTaskRottenness,
			taskLists:      make([]todo.TaskList, 0),
			expectedResult: make([]TaskRottennessInfo, 0),
		},
		{
			name:          "Empty lists",
			minRottenness: FreshTaskRottenness,
			taskLists: []todo.TaskList{
				{
					Name:  "List1",
					Tasks: nil,
				},
				{
					Name:  "List2",
					Tasks: nil,
				},
			},
			expectedResult: make([]TaskRottennessInfo, 0),
		},
		{
			name:          "All tasks",
			minRottenness: FreshTaskRottenness,
			taskLists: []todo.TaskList{
				{
					Name: "List1",
					Tasks: []todo.Task{
						{
							Title:                "Task1",
							CreatedDateTime:      getDateFromNow(1, 0),
							LastModifiedDateTime: getDateFromNow(1, 0),
						},
						{
							Title:                "Task2",
							CreatedDateTime:      getDateFromNow(10, 0),
							LastModifiedDateTime: getDateFromNow(10, 0),
						},
					},
				},
				{
					Name: "List2",
					Tasks: []todo.Task{
						{
							Title:                "Task3",
							CreatedDateTime:      getDateFromNow(0, 5),
							LastModifiedDateTime: getDateFromNow(0, 5),
						},
						{
							Title:                "Task4",
							CreatedDateTime:      getDateFromNow(20, 0),
							LastModifiedDateTime: getDateFromNow(20, 0),
						},
						{
							Title:                "Task5",
							CreatedDateTime:      getDateFromNow(4, 0),
							LastModifiedDateTime: getDateFromNow(4, 0),
						},
					},
				},
			},
			expectedResult: []TaskRottennessInfo{
				{
					TaskName:   "Task4",
					TaskList:   "List2",
					Age:        20,
					Rottenness: ZombieTaskRottenness,
				},
				{
					TaskName:   "Task2",
					TaskList:   "List1",
					Age:        10,
					Rottenness: TiredTaskRottenness,
				},
				{
					TaskName:   "Task5",
					TaskList:   "List2",
					Age:        4,
					Rottenness: RipeTaskRottenness,
				},
				{
					TaskName:   "Task1",
					TaskList:   "List1",
					Age:        1,
					Rottenness: FreshTaskRottenness,
				},
				{
					TaskName:   "Task3",
					TaskList:   "List2",
					Age:        0,
					Rottenness: FreshTaskRottenness,
				},
			},
		},
		{
			name:          "Tired and zombie tasks",
			minRottenness: TiredTaskRottenness,
			taskLists: []todo.TaskList{
				{
					Name: "List1",
					Tasks: []todo.Task{
						{
							Title:                "Task1",
							CreatedDateTime:      getDateFromNow(1, 0),
							LastModifiedDateTime: getDateFromNow(1, 0),
						},
						{
							Title:                "Task2",
							CreatedDateTime:      getDateFromNow(10, 0),
							LastModifiedDateTime: getDateFromNow(10, 0),
						},
					},
				},
				{
					Name: "List2",
					Tasks: []todo.Task{
						{
							Title:                "Task3",
							CreatedDateTime:      getDateFromNow(0, 5),
							LastModifiedDateTime: getDateFromNow(0, 5),
						},
						{
							Title:                "Task4",
							CreatedDateTime:      getDateFromNow(20, 0),
							LastModifiedDateTime: getDateFromNow(20, 0),
						},
						{
							Title:                "Task5",
							CreatedDateTime:      getDateFromNow(4, 0),
							LastModifiedDateTime: getDateFromNow(4, 0),
						},
					},
				},
			},
			expectedResult: []TaskRottennessInfo{
				{
					TaskName:   "Task4",
					TaskList:   "List2",
					Age:        20,
					Rottenness: ZombieTaskRottenness,
				},
				{
					TaskName:   "Task2",
					TaskList:   "List1",
					Age:        10,
					Rottenness: TiredTaskRottenness,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New(tt.taskLists)
			tasks := m.GetRottenTasks(tt.minRottenness)

			assert.Equal(t, tt.expectedResult, tasks)
		})
	}
}
