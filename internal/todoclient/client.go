package todoclient

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/pkg/errors"

	"ToDoInfo/internal/httpclient"
	"ToDoInfo/internal/log"
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
		return nil, errors.Errorf("request task lists error. %v", err)
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

	log.Debug(fmt.Sprintf("Request tasks infos '%s'", taskListId))

	responseBody, err := httpclient.GetRequest(baseRequestUrl+fmt.Sprintf("/%s/", taskListId)+taskListUrl, token)
	if err != nil {
		return nil, err
	}

	type taskListResponse struct {
		Value []todo.Task `json:"value"`
	}

	err = httpclient.GetResponseError(responseBody)
	if err != nil {
		return nil, errors.Errorf("request tasks '%s' error. %v", taskListId, err)
	}

	tl := taskListResponse{}
	err = json.Unmarshal(responseBody, &tl)
	if err != nil {
		return nil, err
	}

	return tl.Value, nil
}

func (parser *TodoParser) GetTasks(token string) ([]todo.TaskList, error) {
	log.Info("Request task lists")

	taskListInfos, err := parser.requestTaskListInfos(token)
	if err != nil {
		return nil, err
	}

	taskListCh := make(chan todo.TaskList)
	wg := sync.WaitGroup{}
	for i := range taskListInfos {
		wg.Add(1)
		go func(info taskListInfo) {
			defer wg.Done()
			tasks, err := parser.requestTaskList(token, info.ID)
			if err != nil {
				log.Error(err)
				return
			}

			taskList := todo.TaskList{
				Name:              info.DisplayName,
				WellknownListName: info.WellknownListName,
				IsShared:          info.IsShared,
				Tasks:             tasks}
			taskListCh <- taskList
		}(taskListInfos[i])
	}

	outTaskListCh := make(chan []todo.TaskList)
	go func() {
		taskLists := make([]todo.TaskList, 0)
		for taskList := range taskListCh {
			taskLists = append(taskLists, taskList)
		}
		outTaskListCh <- taskLists
	}()

	wg.Wait()
	close(taskListCh)
	taskLists := <-outTaskListCh
	close(outTaskListCh)

	return taskLists, err
}
