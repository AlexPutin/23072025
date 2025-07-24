package downloader

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/alexputin/downloader/config"
	"github.com/alexputin/downloader/internal/task"
	"github.com/labstack/gommon/log"
)

type TaskDownloader struct {
	repo task.TaskRepository
	cfg  *config.Config
}

func NewTaskDownloader(repo task.TaskRepository, config *config.Config) *TaskDownloader {
	return &TaskDownloader{
		repo: repo,
		cfg:  config,
	}
}

func (d *TaskDownloader) StartProcess(task *task.Task) (*task.Task, error) {
	go d.runProcessTask(task)
	return task, nil
}

func (d *TaskDownloader) runProcessTask(downloadTask *task.Task) {
	downloadTask.Status = task.StatusProcessing

	resultDirectory := d.cfg.Service.ArchiveDirectory

	err := os.MkdirAll(resultDirectory, os.ModePerm)
	if err != nil {
		log.Error("create directory error:", resultDirectory, err.Error())
		downloadTask.Status = "error"
		return
	}

	tmpFile, err := os.CreateTemp(resultDirectory, fmt.Sprintf("%s_*.tmp", downloadTask.Id))
	if err != nil {
		log.Error("create temp file error:", resultDirectory, err.Error())
		downloadTask.Status = "error"
		return
	}

	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	zipWriter := zip.NewWriter(tmpFile)
	defer zipWriter.Close()

	client := http.Client{
		Timeout: 30 * time.Second,
	}
	usedNames := make(map[string]int)

	for i := range downloadTask.Files {
		file := &downloadTask.Files[i]

		baseName := generateFileName(file.URL, i)
		if count, exists := usedNames[baseName]; exists {
			usedNames[baseName] = count + 1
			baseName = uniqueName(baseName, count)
		} else {
			usedNames[baseName] = 1
		}

		resp, err := client.Get(file.URL)
		if err != nil {
			log.Error("download error:", err.Error())
			file.Error = fmt.Sprintf("download error: %v", err)
			continue
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			resp.Body.Close()
			log.Error("download error code:", resp.StatusCode)
			file.Error = fmt.Sprintf("HTTP status %d", resp.StatusCode)
			continue
		}

		zipEntry, err := zipWriter.Create(baseName)
		if err != nil {
			resp.Body.Close()
			log.Error("zip create error:", err.Error())
			file.Error = fmt.Sprintf("zip create error: %v", err)
			continue
		}

		_, err = io.Copy(zipEntry, resp.Body)
		resp.Body.Close()
		if err != nil {
			log.Error("zip write error:", err.Error())
			file.Error = fmt.Sprintf("zip write error: %v", err)
			continue
		}
	}

	if err := zipWriter.Close(); err != nil {
		log.Error("zip writer close error:", err.Error())
		downloadTask.Status = task.StatusError
		return
	}

	if err := tmpFile.Close(); err != nil {
		log.Error("tmp file close error:", err.Error())
		downloadTask.Status = task.StatusError
		return
	}

	finalName := fmt.Sprintf("%s/%s.zip", resultDirectory, downloadTask.Id.String())
	if err := os.Rename(tmpFile.Name(), finalName); err != nil {
		log.Error("rename temp archive error:", err.Error())
		downloadTask.Status = task.StatusError
		return
	}

	downloadTask.Status = task.StatusDone
	downloadTask.StoredFile = finalName
	downloadTask.ArchiveUrl = fmt.Sprintf("%s:%d/archive/%s", d.cfg.Server.Host, d.cfg.Server.Port, downloadTask.Id)
	_, err = d.repo.Update(downloadTask)
	if err != nil {
		log.Error("update task error:", err.Error())
		return
	}
}

func generateFileName(urlPath string, index int) string {
	url, err := url.Parse(urlPath)
	if err != nil {
		return fmt.Sprintf("file_%d", index)
	}

	base := path.Base(url.Path)
	if base == "." || base == "/" || base == "" {
		return fmt.Sprintf("file_%d", index)
	}

	return base
}

func uniqueName(name string, count int) string {
	ext := filepath.Ext(name)
	base := name[:len(name)-len(ext)]
	return fmt.Sprintf("%s_%d%s", base, count, ext)
}
