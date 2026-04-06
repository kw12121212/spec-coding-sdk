package tasks

import (
	"errors"
	"testing"
	"time"
)

func newTestRegistry(times ...time.Time) *Registry {
	r := NewRegistry()
	if len(times) == 0 {
		return r
	}

	index := 0
	r.now = func() time.Time {
		if index >= len(times) {
			return times[len(times)-1]
		}
		current := times[index]
		index++
		return current
	}

	return r
}

func TestNewRegistryStartsEmpty(t *testing.T) {
	registry := NewRegistry()

	_, err := registry.Get("missing")
	if !errors.Is(err, ErrTaskNotFound) {
		t.Fatalf("expected ErrTaskNotFound, got %v", err)
	}
}

func TestCreateSetsPendingFieldsAndClonesMetadata(t *testing.T) {
	createdAt := time.Date(2026, 4, 6, 10, 0, 0, 0, time.UTC)
	registry := newTestRegistry(createdAt)
	metadata := map[string]string{"priority": "high"}

	task, err := registry.Create(CreateInput{
		ID:       "task-1",
		Title:    "Draft proposal",
		Metadata: metadata,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if task.ID != "task-1" {
		t.Fatalf("expected task id task-1, got %q", task.ID)
	}
	if task.Title != "Draft proposal" {
		t.Fatalf("expected title to be preserved, got %q", task.Title)
	}
	if task.Status != StatusPending {
		t.Fatalf("expected pending status, got %q", task.Status)
	}
	if !task.CreatedAt.Equal(createdAt) {
		t.Fatalf("expected created_at %v, got %v", createdAt, task.CreatedAt)
	}
	if !task.UpdatedAt.Equal(createdAt) {
		t.Fatalf("expected updated_at %v, got %v", createdAt, task.UpdatedAt)
	}
	if task.Metadata["priority"] != "high" {
		t.Fatalf("expected metadata to be stored, got %#v", task.Metadata)
	}

	metadata["priority"] = "low"
	stored, err := registry.Get("task-1")
	if err != nil {
		t.Fatalf("get stored task: %v", err)
	}
	if stored.Metadata["priority"] != "high" {
		t.Fatalf("expected stored metadata to be isolated from input mutation, got %#v", stored.Metadata)
	}

	stored.Metadata["priority"] = "changed"
	reloaded, err := registry.Get("task-1")
	if err != nil {
		t.Fatalf("reload task: %v", err)
	}
	if reloaded.Metadata["priority"] != "high" {
		t.Fatalf("expected retrieved metadata to be cloned, got %#v", reloaded.Metadata)
	}
}

func TestCreateValidation(t *testing.T) {
	t.Run("empty id", func(t *testing.T) {
		registry := NewRegistry()

		_, err := registry.Create(CreateInput{Title: "Title"})
		if !errors.Is(err, ErrTaskIDRequired) {
			t.Fatalf("expected ErrTaskIDRequired, got %v", err)
		}
	})

	t.Run("empty title", func(t *testing.T) {
		registry := NewRegistry()

		_, err := registry.Create(CreateInput{ID: "task-1"})
		if !errors.Is(err, ErrTaskTitleRequired) {
			t.Fatalf("expected ErrTaskTitleRequired, got %v", err)
		}
	})

	t.Run("duplicate id", func(t *testing.T) {
		registry := NewRegistry()
		if _, err := registry.Create(CreateInput{ID: "task-1", Title: "First"}); err != nil {
			t.Fatalf("initial create: %v", err)
		}

		_, err := registry.Create(CreateInput{ID: "task-1", Title: "Second"})
		if !errors.Is(err, ErrTaskAlreadyExists) {
			t.Fatalf("expected ErrTaskAlreadyExists, got %v", err)
		}

		stored, err := registry.Get("task-1")
		if err != nil {
			t.Fatalf("get stored task: %v", err)
		}
		if stored.Title != "First" {
			t.Fatalf("expected original task to remain unchanged, got %q", stored.Title)
		}
	})
}

func TestUpdateChangesFieldsAndPreservesCreatedAt(t *testing.T) {
	createdAt := time.Date(2026, 4, 6, 10, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(5 * time.Minute)
	registry := newTestRegistry(createdAt, updatedAt)

	created, err := registry.Create(CreateInput{
		ID:       "task-1",
		Title:    "Initial title",
		Metadata: map[string]string{"priority": "medium"},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	newTitle := "Updated title"
	newStatus := StatusInProgress
	updated, err := registry.Update("task-1", UpdateInput{
		Title:    &newTitle,
		Metadata: map[string]string{"priority": "high"},
		Status:   &newStatus,
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}

	if updated.Title != "Updated title" {
		t.Fatalf("expected updated title, got %q", updated.Title)
	}
	if updated.Status != StatusInProgress {
		t.Fatalf("expected in_progress, got %q", updated.Status)
	}
	if updated.Metadata["priority"] != "high" {
		t.Fatalf("expected updated metadata, got %#v", updated.Metadata)
	}
	if !updated.CreatedAt.Equal(created.CreatedAt) {
		t.Fatalf("expected created_at to be preserved, got %v", updated.CreatedAt)
	}
	if !updated.UpdatedAt.Equal(updatedAt) {
		t.Fatalf("expected updated_at %v, got %v", updatedAt, updated.UpdatedAt)
	}
}

func TestDeleteMarksTaskDeletedAndKeepsItReadable(t *testing.T) {
	createdAt := time.Date(2026, 4, 6, 10, 0, 0, 0, time.UTC)
	deletedAt := createdAt.Add(10 * time.Minute)
	registry := newTestRegistry(createdAt, deletedAt)

	if _, err := registry.Create(CreateInput{ID: "task-1", Title: "Disposable"}); err != nil {
		t.Fatalf("create: %v", err)
	}

	deleted, err := registry.Delete("task-1")
	if err != nil {
		t.Fatalf("delete: %v", err)
	}
	if deleted.Status != StatusDeleted {
		t.Fatalf("expected deleted status, got %q", deleted.Status)
	}
	if !deleted.UpdatedAt.Equal(deletedAt) {
		t.Fatalf("expected updated_at %v, got %v", deletedAt, deleted.UpdatedAt)
	}

	stored, err := registry.Get("task-1")
	if err != nil {
		t.Fatalf("get deleted task: %v", err)
	}
	if stored.Status != StatusDeleted {
		t.Fatalf("expected stored task to remain readable with deleted status, got %q", stored.Status)
	}
}

func TestUnknownTaskOperationsFail(t *testing.T) {
	registry := NewRegistry()
	newTitle := "Updated title"
	inProgress := StatusInProgress

	_, getErr := registry.Get("missing")
	if !errors.Is(getErr, ErrTaskNotFound) {
		t.Fatalf("expected ErrTaskNotFound from get, got %v", getErr)
	}

	_, updateErr := registry.Update("missing", UpdateInput{Title: &newTitle, Status: &inProgress})
	if !errors.Is(updateErr, ErrTaskNotFound) {
		t.Fatalf("expected ErrTaskNotFound from update, got %v", updateErr)
	}

	_, deleteErr := registry.Delete("missing")
	if !errors.Is(deleteErr, ErrTaskNotFound) {
		t.Fatalf("expected ErrTaskNotFound from delete, got %v", deleteErr)
	}
}

func TestInvalidStatusTransitionsDoNotMutateTask(t *testing.T) {
	createdAt := time.Date(2026, 4, 6, 10, 0, 0, 0, time.UTC)
	startedAt := createdAt.Add(5 * time.Minute)
	completedAt := createdAt.Add(10 * time.Minute)
	registry := newTestRegistry(createdAt, startedAt, completedAt)

	if _, err := registry.Create(CreateInput{ID: "task-1", Title: "Transition task"}); err != nil {
		t.Fatalf("create: %v", err)
	}

	completed := StatusCompleted
	if _, err := registry.Update("task-1", UpdateInput{Status: &completed}); !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("expected ErrInvalidTransition for pending->completed, got %v", err)
	}

	stored, err := registry.Get("task-1")
	if err != nil {
		t.Fatalf("get after failed transition: %v", err)
	}
	if stored.Status != StatusPending {
		t.Fatalf("expected task to remain pending after failed transition, got %q", stored.Status)
	}
	if !stored.UpdatedAt.Equal(createdAt) {
		t.Fatalf("expected updated_at to remain unchanged, got %v", stored.UpdatedAt)
	}

	inProgress := StatusInProgress
	if _, err := registry.Update("task-1", UpdateInput{Status: &inProgress}); err != nil {
		t.Fatalf("pending->in_progress: %v", err)
	}
	if _, err := registry.Update("task-1", UpdateInput{Status: &completed}); err != nil {
		t.Fatalf("in_progress->completed: %v", err)
	}

	if _, err := registry.Delete("task-1"); !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("expected ErrInvalidTransition for completed->deleted, got %v", err)
	}

	finalTask, err := registry.Get("task-1")
	if err != nil {
		t.Fatalf("get completed task: %v", err)
	}
	if finalTask.Status != StatusCompleted {
		t.Fatalf("expected task to remain completed, got %q", finalTask.Status)
	}
	if !finalTask.UpdatedAt.Equal(completedAt) {
		t.Fatalf("expected updated_at to remain at completion time, got %v", finalTask.UpdatedAt)
	}
}

func TestUnknownStatusFailsWithoutChangingTask(t *testing.T) {
	createdAt := time.Date(2026, 4, 6, 10, 0, 0, 0, time.UTC)
	registry := newTestRegistry(createdAt)

	if _, err := registry.Create(CreateInput{ID: "task-1", Title: "Unknown status task"}); err != nil {
		t.Fatalf("create: %v", err)
	}

	unknown := Status("archived")
	if _, err := registry.Update("task-1", UpdateInput{Status: &unknown}); !errors.Is(err, ErrUnknownTaskStatus) {
		t.Fatalf("expected ErrUnknownTaskStatus, got %v", err)
	}

	stored, err := registry.Get("task-1")
	if err != nil {
		t.Fatalf("get after failed status update: %v", err)
	}
	if stored.Status != StatusPending {
		t.Fatalf("expected task to remain pending, got %q", stored.Status)
	}
}
