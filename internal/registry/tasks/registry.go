// Package tasks provides an in-memory task registry with lifecycle validation.
package tasks

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	// ErrTaskIDRequired reports that a task ID was not provided.
	ErrTaskIDRequired = errors.New("tasks: task id is required")
	// ErrTaskTitleRequired reports that a task title was not provided.
	ErrTaskTitleRequired = errors.New("tasks: task title is required")
	// ErrTaskAlreadyExists reports that a task ID is already present.
	ErrTaskAlreadyExists = errors.New("tasks: task already exists")
	// ErrTaskNotFound reports that a task ID could not be found.
	ErrTaskNotFound = errors.New("tasks: task not found")
	// ErrInvalidTransition reports that a requested status transition is not allowed.
	ErrInvalidTransition = errors.New("tasks: invalid task status transition")
	// ErrUnknownTaskStatus reports that a status value is not supported by the registry.
	ErrUnknownTaskStatus = errors.New("tasks: unknown task status")
)

// Status represents the lifecycle status of a task.
type Status string

const (
	// StatusPending is the initial status for a newly created task.
	StatusPending Status = "pending"
	// StatusInProgress indicates that work has started on the task.
	StatusInProgress Status = "in_progress"
	// StatusCompleted indicates that work on the task is finished.
	StatusCompleted Status = "completed"
	// StatusDeleted indicates that the task has been deleted from active use.
	StatusDeleted Status = "deleted"
)

// Task is the stored task record returned by the registry.
type Task struct {
	ID        string            `json:"id"`
	Title     string            `json:"title"`
	Status    Status            `json:"status"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// CreateInput contains the caller-supplied fields for task creation.
type CreateInput struct {
	ID       string
	Title    string
	Metadata map[string]string
}

// UpdateInput contains the mutable task fields for an update operation.
type UpdateInput struct {
	Title    *string
	Metadata map[string]string
	Status   *Status
}

// Registry stores tasks in memory and validates lifecycle transitions.
type Registry struct {
	mu    sync.RWMutex
	tasks map[string]Task
	now   func() time.Time
}

// NewRegistry returns an empty in-memory task registry.
func NewRegistry() *Registry {
	return &Registry{
		tasks: make(map[string]Task),
		now:   time.Now,
	}
}

// Create stores a new task using the caller-supplied ID and title.
func (r *Registry) Create(input CreateInput) (Task, error) {
	if input.ID == "" {
		return Task{}, ErrTaskIDRequired
	}
	if input.Title == "" {
		return Task{}, ErrTaskTitleRequired
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.ensureStorage()
	if _, exists := r.tasks[input.ID]; exists {
		return Task{}, fmt.Errorf("%w: %s", ErrTaskAlreadyExists, input.ID)
	}

	now := r.nowTime()
	task := Task{
		ID:        input.ID,
		Title:     input.Title,
		Status:    StatusPending,
		CreatedAt: now,
		UpdatedAt: now,
		Metadata:  cloneMetadata(input.Metadata),
	}
	r.tasks[input.ID] = task

	return cloneTask(task), nil
}

// Get returns the task stored under the provided ID.
func (r *Registry) Get(id string) (Task, error) {
	if id == "" {
		return Task{}, ErrTaskIDRequired
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	task, ok := r.tasks[id]
	if !ok {
		return Task{}, fmt.Errorf("%w: %s", ErrTaskNotFound, id)
	}

	return cloneTask(task), nil
}

// Update changes mutable task fields for the provided task ID.
func (r *Registry) Update(id string, input UpdateInput) (Task, error) {
	if id == "" {
		return Task{}, ErrTaskIDRequired
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	task, ok := r.tasks[id]
	if !ok {
		return Task{}, fmt.Errorf("%w: %s", ErrTaskNotFound, id)
	}

	if input.Title != nil {
		if *input.Title == "" {
			return Task{}, ErrTaskTitleRequired
		}
		task.Title = *input.Title
	}

	if input.Metadata != nil {
		task.Metadata = cloneMetadata(input.Metadata)
	}

	if input.Status != nil {
		if err := validateStatus(*input.Status); err != nil {
			return Task{}, err
		}
		if err := validateTransition(task.Status, *input.Status); err != nil {
			return Task{}, err
		}
		task.Status = *input.Status
	}

	task.UpdatedAt = r.nowTime()
	r.tasks[id] = task

	return cloneTask(task), nil
}

// Delete marks the task for the provided ID as deleted.
func (r *Registry) Delete(id string) (Task, error) {
	deleted := StatusDeleted
	return r.Update(id, UpdateInput{Status: &deleted})
}

func (r *Registry) ensureStorage() {
	if r.tasks == nil {
		r.tasks = make(map[string]Task)
	}
	if r.now == nil {
		r.now = time.Now
	}
}

func (r *Registry) nowTime() time.Time {
	r.ensureStorage()
	return r.now()
}

func validateStatus(status Status) error {
	switch status {
	case StatusPending, StatusInProgress, StatusCompleted, StatusDeleted:
		return nil
	default:
		return fmt.Errorf("%w: %s", ErrUnknownTaskStatus, status)
	}
}

func validateTransition(from, to Status) error {
	switch from {
	case StatusPending:
		if to == StatusInProgress || to == StatusDeleted {
			return nil
		}
	case StatusInProgress:
		if to == StatusCompleted || to == StatusDeleted {
			return nil
		}
	case StatusCompleted, StatusDeleted:
	}

	return fmt.Errorf("%w: %s -> %s", ErrInvalidTransition, from, to)
}

func cloneTask(task Task) Task {
	task.Metadata = cloneMetadata(task.Metadata)
	return task
}

func cloneMetadata(metadata map[string]string) map[string]string {
	if metadata == nil {
		return nil
	}

	cloned := make(map[string]string, len(metadata))
	for key, value := range metadata {
		cloned[key] = value
	}

	return cloned
}
