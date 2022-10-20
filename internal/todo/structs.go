package todo

import (
	"encoding/json"
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

type TaskLists struct {
	Lists []TaskList
}

func (t *Task) UnmarshalJSON(data []byte) error {
	const layout = "2006-01-02T15:04:05.9999999Z"

	type TaskAlias Task

	aliasValue := &struct {
		*TaskAlias
		CreatedDateTime      string `json:"createdDateTime"`
		LastModifiedDateTime string `json:"lastModifiedDateTime"`
	}{
		TaskAlias: (*TaskAlias)(t),
	}

	if err := json.Unmarshal(data, aliasValue); err != nil {
		return err
	}

	var err error
	t.CreatedDateTime, err = time.Parse(layout, aliasValue.LastModifiedDateTime)
	if err != nil {
		return err
	}
	t.LastModifiedDateTime, err = time.Parse(layout, aliasValue.LastModifiedDateTime)
	if err != nil {
		return err
	}

	return nil
}
