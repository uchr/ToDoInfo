package todoclient

import (
	"encoding/json"
	"fmt"

	"ToDoInfo/internal/httpclient"
	"ToDoInfo/internal/todo"
)

const baseRequestUrl = "https://graph.microsoft.com/v1.0/me/todo/lists"

type taskListInfo struct {
	ID                string `json:"id"`
	DisplayName       string `json:"displayName"`
	WellknownListName string `json:"wellknownListName"`
	IsShared          bool   `json:"isShared"`
}

type TodoParser struct {
}

func New() *TodoParser {
	return &TodoParser{}
}

func (parser *TodoParser) requestTaskListInfos(token string) ([]taskListInfo, error) {
	responseBody, err := httpclient.GetRequest(baseRequestUrl, token)
	if err != nil {
		return nil, err
	}

	type taskListResponse struct {
		Value []taskListInfo `json:"value"`
	}

	err = httpclient.GetResponseError(responseBody)
	if err != nil {
		return nil, err
	}

	tl := taskListResponse{}
	err = json.Unmarshal(responseBody, &tl)
	if err != nil {
		return nil, err
	}
	return tl.Value, nil
}

func (parser *TodoParser) requestTaskList(token string, taskListId string) ([]todo.Task, error) {
	const taskListUrl = "tasks?$filter=status%20eq%20'notStarted'"
	responseBody, err := httpclient.GetRequest(baseRequestUrl+fmt.Sprintf("/%s/", taskListId)+taskListUrl, token)
	if err != nil {
		return nil, err
	}

	type taskListResponse struct {
		Value []todo.Task `json:"value"`
	}

	err = httpclient.GetResponseError(responseBody)
	if err != nil {
		return nil, err
	}

	tl := taskListResponse{}
	err = json.Unmarshal(responseBody, &tl)
	if err != nil {
		return nil, err
	}

	return tl.Value, nil
}

func (parser *TodoParser) GetTasks(token string) ([]todo.TaskList, error) {
	taskListInfos, err := parser.requestTaskListInfos(token)
	if err != nil {
		return nil, err
	}

	taskLists := make([]todo.TaskList, 0)
	for _, info := range taskListInfos {
		tasks, err := parser.requestTaskList(token, info.ID)
		if err != nil {
			return nil, err
		}

		taskList := todo.TaskList{
			Name:              info.DisplayName,
			WellknownListName: info.WellknownListName,
			IsShared:          info.IsShared,
			Tasks:             tasks}
		taskLists = append(taskLists, taskList)
	}

	return taskLists, nil
}
