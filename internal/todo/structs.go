package todo

import (
	"time"
)

type TaskBody struct {
	Content     string `json:"content"`
	ContentType string `json:"contentType"`
}

type Task struct {
	IsReminderOn         bool      `json:"isReminderOn"`
	Title                string    `json:"title"`
	CreatedDateTime      time.Time `json:"createdDateTime"`
	LastModifiedDateTime time.Time `json:"lastModifiedDateTime"`
	Body                 TaskBody  `json:"body"`
}

type TaskList struct {
	Name              string
	WellknownListName string
	IsShared          bool

	Tasks []Task
}
