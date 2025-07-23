package task

import "github.com/google/uuid"

type Task struct {
	Id     uuid.UUID `json:"id"`
	Status string    `json:"status"`
	Files  []File    `json:"files"`
}

type File struct {
	URL   string `json:"url"`
	Error string `json:"error,omitempty"`
}
