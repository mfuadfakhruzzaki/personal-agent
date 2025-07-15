package repository

import (
	"todo-agent-backend/internal/models"
	"todo-agent-backend/pkg/supabase"
)

// TodoRepository handles todo database operations
type TodoRepository struct {
	client *supabase.Client
}

// NewTodoRepository creates a new todo repository
func NewTodoRepository(client *supabase.Client) *TodoRepository {
	return &TodoRepository{
		client: client,
	}
}

// InsertTodo inserts a single todo
func (tr *TodoRepository) InsertTodo(todo *models.Todo) error {
	return tr.client.InsertTodo(todo)
}

// InsertTodos inserts multiple todos
func (tr *TodoRepository) InsertTodos(todos []models.Todo) error {
	return tr.client.InsertTodos(todos)
}

// GetTodosByUserID retrieves todos for a user
func (tr *TodoRepository) GetTodosByUserID(userID string) ([]models.Todo, error) {
	return tr.client.GetTodosByUserID(userID)
}
