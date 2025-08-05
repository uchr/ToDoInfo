package todo

import (
	"encoding/json"
	"time"
)

func (t *Task) UnmarshalJSON(data []byte) error {
	const layout = "2006-01-02T15:04:05.9999999Z"

	type TaskAlias Task

	// First try the original API format
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
	} else {
		// Try fallback format for stored data (simple string format)
		fallbackValue := &struct {
			*TaskAlias
			CreatedDateTime      string     `json:"createdDateTime"`
			LastModifiedDateTime string     `json:"lastModifiedDateTime"`
			DueDateTime          *time.Time `json:"dueDateTime"`
		}{
			TaskAlias: (*TaskAlias)(t),
		}

		if err := json.Unmarshal(data, fallbackValue); err == nil && fallbackValue.DueDateTime != nil {
			t.DueDateTime = fallbackValue.DueDateTime
		}
	}

	return nil
}

func (t *Task) MarshalJSON() ([]byte, error) {
	type TaskAlias Task

	aliasValue := &struct {
		*TaskAlias
		CreatedDateTime      string  `json:"createdDateTime"`
		LastModifiedDateTime string  `json:"lastModifiedDateTime"`
		DueDateTime          *struct {
			DateTime string `json:"dateTime"`
			TimeZone string `json:"timeZone"`
		} `json:"dueDateTime"`
	}{
		TaskAlias:            (*TaskAlias)(t),
		CreatedDateTime:      t.CreatedDateTime.Format("2006-01-02T15:04:05.9999999Z"),
		LastModifiedDateTime: t.LastModifiedDateTime.Format("2006-01-02T15:04:05.9999999Z"),
	}

	if t.DueDateTime != nil {
		aliasValue.DueDateTime = &struct {
			DateTime string `json:"dateTime"`
			TimeZone string `json:"timeZone"`
		}{
			DateTime: t.DueDateTime.Format("2006-01-02T15:04:05.9999999"),
			TimeZone: "UTC",
		}
	}

	return json.Marshal(aliasValue)
}
