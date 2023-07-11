package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.34

import (
	"context"
	"time"

	"github.com/romshark/taskhub/api/auth"
	"github.com/romshark/taskhub/api/graph/model"
)

// AccessToken is the resolver for the accessToken field.
func (r *queryResolver) AccessToken(ctx context.Context, email string, password string) (string, error) {
	user, err := r.DataProvider.UserByEmail(ctx, email)
	if err != nil {
		return "", err
	}

	ok, err := r.Resolver.PasswordHasher.ComparePassword(
		[]byte(password), []byte(user.PasswordHash),
	)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", auth.ErrUnauthorized
	}

	return r.Resolver.JWTGenerator.GenerateJWT(user.ID, 24*time.Hour)
}

// Task is the resolver for the task field.
func (r *queryResolver) Task(ctx context.Context, id string) (*model.Task, error) {
	if err := auth.RequireAuthenticated(ctx); err != nil {
		return nil, err
	}
	return r.DataProvider.TaskByID(ctx, id)
}

// User is the resolver for the user field.
func (r *queryResolver) User(ctx context.Context, id string) (*model.User, error) {
	if err := auth.RequireAuthenticated(ctx); err != nil {
		return nil, err
	}
	return r.DataProvider.UserByID(ctx, id)
}

// Project is the resolver for the project field.
func (r *queryResolver) Project(ctx context.Context, id string) (*model.Project, error) {
	if err := auth.RequireAuthenticated(ctx); err != nil {
		return nil, err
	}
	return r.DataProvider.ProjectByID(ctx, id)
}

// Tasks is the resolver for the tasks field.
func (r *queryResolver) Tasks(ctx context.Context, filters *model.TasksFilters, order *model.TasksOrder, orderAsc bool, limit *int) ([]*model.Task, error) {
	if err := auth.RequireAuthenticated(ctx); err != nil {
		return nil, err
	}
	return r.DataProvider.GetTasks(ctx, filters, order, orderAsc, limit)
}

// Users is the resolver for the users field.
func (r *queryResolver) Users(ctx context.Context, filters *model.UsersFilters, order *model.UsersOrder, orderAsc bool, limit *int) ([]*model.User, error) {
	if err := auth.RequireAuthenticated(ctx); err != nil {
		return nil, err
	}
	return r.DataProvider.GetUsers(ctx, filters, order, orderAsc, limit)
}

// Projects is the resolver for the projects field.
func (r *queryResolver) Projects(ctx context.Context, filters *model.ProjectsFilters, order *model.ProjectsOrder, orderAsc bool, limit *int) ([]*model.Project, error) {
	if err := auth.RequireAuthenticated(ctx); err != nil {
		return nil, err
	}
	return r.DataProvider.GetProjects(ctx, filters, order, orderAsc, limit)
}

// Query returns QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

type queryResolver struct{ *Resolver }
