package task

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
)

type InMemoryTaskRepository struct {
	table map[string]Task
	mutex sync.RWMutex
}

func NewInMemoryTaskRepository() *InMemoryTaskRepository {
	return &InMemoryTaskRepository{
		table: make(map[string]Task),
	}
}

func (r *InMemoryTaskRepository) Create() (Task, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	id := uuid.New()
	task := Task{Id: id, Status: StatusCreated}
	r.table[id.String()] = task
	return task, nil
}

func (r *InMemoryTaskRepository) Get(id string) (Task, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return r.internalGet(id)
}

func (r *InMemoryTaskRepository) AddFile(taskId string, file File) (Task, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	task, err := r.internalGet(taskId)
	if err != nil {
		return Task{}, err
	}
	task.AddFile(file)

	r.table[taskId] = task
	return task, nil
}

func (r *InMemoryTaskRepository) Update(task *Task) (Task, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	id := task.Id.String()
	tsk, err := r.internalGet(id)
	if err != nil {
		return Task{}, err
	}

	r.table[id] = *task
	return tsk, nil
}

func (r *InMemoryTaskRepository) GetActiveTaskCount() int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	result := 0
	for _, value := range r.table {
		if value.Status == StatusCreated || value.Status == StatusProcessing {
			result += 1
		}
	}

	return result
}

func (r *InMemoryTaskRepository) internalGet(id string) (Task, error) {
	task, ok := r.table[id]
	if !ok {
		return Task{}, fmt.Errorf("task with id '%s' not found", id)
	}
	return task, nil
}
