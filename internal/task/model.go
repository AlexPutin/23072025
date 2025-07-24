package task

import "github.com/google/uuid"

type Status = string

var (
	StatusDone       Status = "done"
	StatusProcessing Status = "processing"
	StatusCreated    Status = "created"
	StatusError      Status = "error"
)

type Task struct {
	Id         uuid.UUID `json:"id"`
	Status     Status    `json:"status"`
	Files      []File    `json:"files,omitempty"`
	ArchiveUrl string    `json:"archive,omitempty"`
	StoredFile string    `json:"-"`
}

type File struct {
	URL   string `json:"url"`
	Error string `json:"error,omitempty"`
}

func (t *Task) AddFile(file File) {
	t.Files = append(t.Files, file)
}
