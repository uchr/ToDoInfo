package todo

import (
	"time"
)

type TaskBody struct {
	Content     string `json:"content"`
	ContentType string `json:"contentType"`
}

type RecurrenceTask struct {
	Pattern struct {
		Type           string `json:"type"`
		Interval       int    `json:"interval"`
		Month          int    `json:"month"`
		DayOfMonth     int    `json:"dayOfMonth"`
		FirstDayOfWeek string `json:"firstDayOfWeek"`
		Index          string `json:"index"`
	} `json:"pattern"`
	Range struct {
		Type string `json:"type"`
	} `json:"range"`
}

type Task struct {
	IsReminderOn         bool            `json:"isReminderOn"`
	Title                string          `json:"title"`
	CreatedDateTime      time.Time       `json:"createdDateTime"`
	LastModifiedDateTime time.Time       `json:"lastModifiedDateTime"`
	DueDateTime          time.Time       `json:"dueDateTime"`
	Body                 TaskBody        `json:"body"`
	Recurrence           *RecurrenceTask `json:"recurrence"`
}

type TaskList struct {
	Name              string
	WellknownListName string
	IsShared          bool

	Tasks []Task
}
