package dataprovider

import (
	"context"
	"time"

	"github.com/romshark/taskhub/api/graph/model"
)

type DataProvider interface {
	Reader
	Writer
}

// Reader reads the data source
type Reader interface {
	UserByEmail(ctx context.Context, email string) (*model.User, error)
	UserByID(ctx context.Context, id string) (*model.User, error)
	ProjectByID(ctx context.Context, id string) (*model.Project, error)
	TaskByID(ctx context.Context, id string) (*model.Task, error)

	GetUsers(
		ctx context.Context,
		filters *model.UsersFilters,
		order *model.UsersOrder,
		orderAsc bool,
		limit *int,
	) ([]*model.User, error)

	GetProjects(
		ctx context.Context,
		filters *model.ProjectsFilters,
		order *model.ProjectsOrder,
		orderAsc bool,
		limit *int,
	) ([]*model.Project, error)

	GetTasks(
		ctx context.Context,
		filters *model.TasksFilters,
		order *model.TasksOrder,
		orderAsc bool,
		limit *int,
	) ([]*model.Task, error)

	// GetProjectMembers returns all users that are assigned to tasks
	// or have reported tasks from the given project.
	GetProjectMembers(
		ctx context.Context,
		projectID string,
	) ([]*model.User, error)

	// GetBlockingTasks returns all tasks that are blocking the given task.
	GetBlockingTasks(
		ctx context.Context,
		taskID string,
	) ([]*model.Task, error)

	// GetRelatedTasks returns all related tasks for the given task.
	GetRelatedTasks(
		ctx context.Context,
		taskID string,
	) ([]*model.Task, error)

	// GetTasksByProject returns all tasks assigned to the given project.
	GetTasksByProject(
		ctx context.Context,
		projectID string,
	) ([]*model.Task, error)

	// GetUserProjects returns all projects that the tasks are assigned to
	// where the given user is either assignee or reporter.
	GetUserProjects(
		ctx context.Context,
		userID string,
	) ([]*model.Project, error)

	// GetTasksAssignedToUser returns all tasks assigned to the given user.
	GetTasksAssignedToUser(
		ctx context.Context,
		userID string,
	) ([]*model.Task, error)

	// GetTasksReportedByUser returns all tasks reported by the given user.
	GetTasksReportedByUser(
		ctx context.Context,
		userID string,
	) ([]*model.Task, error)
}

// Writer reads from and writes to the data source
type Writer interface {
	CreateUser(
		ctx context.Context,
		email string,
		password string,
		displayName string,
		role string,
		location string,
		manager *string,
		subordinates []string,
	) (*model.User, error)

	UpdateUser(
		ctx context.Context,
		id string,
		email string,
		displayName string,
		role string,
		location string,
		personalStatus *string,
		manager *string,
		subordinates []string,
	) (*model.User, error)

	CreateProject(
		ctx context.Context,
		creation time.Time,
		name string,
		description string,
		slug string,
		owners []string,
	) (*model.Project, error)

	UpdateProject(
		ctx context.Context,
		id string,
		name string,
		description string,
		slug string,
		owners []string,
	) (*model.Project, error)

	CreateTask(
		ctx context.Context,
		creation time.Time,
		title string,
		project string,
		status model.TaskStatus,
		priority model.TaskPriority,
		description *string,
		due *time.Time,
		tags []string,
		assignees []string,
		reporters []string,
		blocks []string,
		relatesTo []string,
	) (*model.Task, error)

	UpdateTask(
		ctx context.Context,
		id string,
		title string,
		description *string,
		status model.TaskStatus,
		priority model.TaskPriority,
		due *time.Time,
		tags []string,
		project string,
		assignees []string,
		reporters []string,
		blocks []string,
		relatesTo []string,
	) (*model.Task, error)
}
