package task

import (
	"fmt"
	"slices"

	"github.com/alexputin/downloader/config"
	"github.com/alexputin/downloader/internal/utils"
)

type TaskRepository interface {
	Create() (Task, error)
	Get(id string) (Task, error)
	AddFile(taskId string, file File) (Task, error)
	Update(task *Task) (Task, error)
	GetActiveTaskCount() int
}

type TaskDownloadService interface {
	StartProcess(task *Task) (*Task, error)
}

type TaskService struct {
	tr  TaskRepository
	dl  TaskDownloadService
	cfg *config.Config
}

func NewTaskService(repo TaskRepository, downloader TaskDownloadService, config *config.Config) *TaskService {
	return &TaskService{
		tr:  repo,
		dl:  downloader,
		cfg: config,
	}
}

func (s *TaskService) CreateTask() (*Task, error) {
	if s.tr.GetActiveTaskCount() >= s.cfg.Service.MaxActiveTasks {
		return &Task{}, fmt.Errorf("to many active tasks")
	}
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

	extension, err := utils.GetUrlExtension(fileUrl)
	if err != nil {
		return &Task{}, fmt.Errorf("invalid url")
	}

	isSupportedFile := s.hasSupportedExtension(extension)
	if !isSupportedFile {
		return &Task{}, fmt.Errorf("invalid file extension")
	}

	if len(task.Files) >= s.cfg.Service.MaxFilesPerTask {
		return &Task{}, fmt.Errorf("max allowed files per task")
	}

	task, err = s.tr.AddFile(taskId, File{
		URL: fileUrl,
	})

	if err != nil {
		return &Task{}, fmt.Errorf("file not added: %w", err)
	}

	if len(task.Files) == s.cfg.Service.MaxFilesPerTask {
		return s.dl.StartProcess(&task)
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

func (s *TaskService) hasSupportedExtension(extension string) bool {
	return slices.Contains(s.cfg.Service.AllowedExtensions, extension)
}
