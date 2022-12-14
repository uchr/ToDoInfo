package todo

import (
	"encoding/json"
	"time"
)

func (t *Task) UnmarshalJSON(data []byte) error {
	const layout = "2006-01-02T15:04:05.9999999Z"

	type TaskAlias Task

	aliasValue := &struct {
		*TaskAlias
		CreatedDateTime      string `json:"createdDateTime"`
		LastModifiedDateTime string `json:"lastModifiedDateTime"`
		DueDateTime          *struct {
			DateTime string `json:"dateTime"`
			TimeZone string `json:"timeZone"`
		} `json:"dueDateTime"`
	}{
		TaskAlias: (*TaskAlias)(t),
	}

	if err := json.Unmarshal(data, aliasValue); err != nil {
		return err
	}

	var err error
	t.CreatedDateTime, err = time.Parse(layout, aliasValue.CreatedDateTime)
	if err != nil {
		return err
	}
	t.LastModifiedDateTime, err = time.Parse(layout, aliasValue.LastModifiedDateTime)
	if err != nil {
		return err
	}

	if aliasValue.DueDateTime != nil {
		const DueDateTimeLayout = "2006-01-02T15:04:05.9999999"
		dueDateTime, err := time.Parse(DueDateTimeLayout, aliasValue.DueDateTime.DateTime)
		if err != nil {
			return err
		}
		t.DueDateTime = &dueDateTime
	}

	return nil
}
