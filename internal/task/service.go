package task

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/alexputin/downloader/config"
)

type TaskRepository interface {
	Create() (Task, error)
	Get(id string) (Task, error)
	AddFile(taskId string, url string) (Task, error)
}

type TaskService struct {
	tr  TaskRepository
	cfg *config.Config
}

func NewTaskService(repo TaskRepository, config *config.Config) *TaskService {
	return &TaskService{
		tr:  repo,
		cfg: config,
	}
}

func (s *TaskService) CreateTask() (*Task, error) {
	task, err := s.tr.Create()
	return &task, err
}

func (s *TaskService) AddFile(taskId, fileUrl string) (*Task, error) {
	task, err := s.tr.Get(taskId)
	if err != nil {
		return &Task{}, fmt.Errorf("add file to task failed: %w", err)
	}

	if task.Status == StatusDone {
		return &Task{}, fmt.Errorf("task already done")
	}

	isSupportedFile, err := s.hasSupportedExtension(fileUrl)
	if err != nil {
		return &Task{}, err
	}

	if !isSupportedFile {
		return &Task{}, fmt.Errorf("invalid file extension")
	}

	if len(task.Files) >= s.cfg.Service.MaxFilesPerTask {
		return &Task{}, fmt.Errorf("max allowed files per task")
	}

	task, err = s.tr.AddFile(taskId, fileUrl)
	if err != nil {
		return &Task{}, fmt.Errorf("file not added: %w", err)
	}

	if len(task.Files) == s.cfg.Service.MaxFilesPerTask {
		// TODO: add processing service call
	}

	return &task, err
}

func (s *TaskService) GetTask(taskId string) (*Task, error) {
	task, err := s.tr.Get(taskId)
	if err != nil {
		return nil, err
	}

	return &task, nil
}

func (s *TaskService) hasSupportedExtension(fileUrl string) (bool, error) {
	parsedURL, err := url.Parse(fileUrl)
	if err != nil {
		return false, fmt.Errorf("invalid url")
	}

	for _, extension := range s.cfg.Service.AllowedExtensions {
		if strings.HasSuffix(parsedURL.Path, extension) {
			return true, nil
		}
	}

	return false, nil
}
