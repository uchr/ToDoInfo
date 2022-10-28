package todoclient

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

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

	taskLists, err := parser.getListInfos(token, taskListInfos)

	return taskLists, err
}

func (parser *TodoParser) getListInfos(token string, taskListInfos []taskListInfo) ([]todo.TaskList, error) {
	const maxRequestForSecond = 4 // Max number of request to MS TODO

	outputCh := make(chan taskListProcessingResult)
	go func() {
		wg := sync.WaitGroup{}
		counter := 0
		for _, info := range taskListInfos {
			if info.WellknownListName != "none" {
				continue
			}

			if counter >= maxRequestForSecond {
				time.Sleep(time.Second)
				counter = 0
			}

			wg.Add(1)
			go func(info taskListInfo) {
				defer wg.Done()
				parser.processTaskListInfo(token, info, outputCh)
			}(info)

			counter++
		}

		wg.Wait()

		close(outputCh)
	}()

	var err error
	var taskLists []todo.TaskList
	for processingResult := range outputCh {
		if processingResult.err != nil {
			err = processingResult.err
			continue
		}
		if err != nil {
			continue
		}
		taskLists = append(taskLists, processingResult.taskList)
	}

	return taskLists, err
}

type taskListProcessingResult struct {
	taskList todo.TaskList
	err      error
}

func (parser *TodoParser) processTaskListInfo(token string, info taskListInfo, out chan taskListProcessingResult) {
	tasks, err := parser.requestTaskList(token, info.ID)
	if err != nil {
		log.Error(err)
		out <- taskListProcessingResult{err: err}
		return
	}

	taskList := todo.TaskList{
		Name:              info.DisplayName,
		WellknownListName: info.WellknownListName,
		IsShared:          info.IsShared,
		Tasks:             tasks}

	out <- taskListProcessingResult{taskList: taskList}
}
