package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.34

import (
	"context"
	"errors"
	"time"

	"github.com/romshark/taskhub/api/graph/auth"
	"github.com/romshark/taskhub/api/graph/model"
	"github.com/romshark/taskhub/slices"
)

// AccessToken is the resolver for the accessToken field.
func (r *queryResolver) AccessToken(ctx context.Context, email string, password string) (string, error) {
	var user *model.User
	for _, u := range r.Resolver.Users {
		if u.Email == email {
			user = u
			break
		}
	}
	if user == nil {
		return "", errors.New("user not found")
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

	for _, t := range r.Resolver.Tasks {
		if t.ID == id {
			return t, nil
		}
	}
	return nil, nil
}

// User is the resolver for the user field.
func (r *queryResolver) User(ctx context.Context, id string) (*model.User, error) {
	if err := auth.RequireAuthenticated(ctx); err != nil {
		return nil, err
	}

	for _, u := range r.Resolver.Users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, nil
}

// Project is the resolver for the project field.
func (r *queryResolver) Project(ctx context.Context, id string) (*model.Project, error) {
	if err := auth.RequireAuthenticated(ctx); err != nil {
		return nil, err
	}

	for _, p := range r.Resolver.Projects {
		if p.ID == id {
			return p, nil
		}
	}
	return nil, nil
}

// Tasks is the resolver for the tasks field.
func (r *queryResolver) Tasks(ctx context.Context, filters *model.TasksFilters, order *model.TasksOrder, orderAsc bool, limit *int) ([]*model.Task, error) {
	if err := auth.RequireAuthenticated(ctx); err != nil {
		return nil, err
	}

	tasks := slices.Copy(r.Resolver.Tasks)
	if filters != nil {
		if filters.CreatedAfter != nil {
			tasks = slices.FilterInPlace(tasks, func(t *model.Task) (ok bool) {
				return t.Creation.Unix() > filters.CreatedAfter.Unix()
			})
			if len(tasks) < 1 {
				return nil, nil
			}
		}
		if filters.CreatedBefore != nil {
			tasks = slices.FilterInPlace(tasks, func(t *model.Task) (ok bool) {
				return t.Creation.Unix() < filters.CreatedBefore.Unix()
			})
			if len(tasks) < 1 {
				return nil, nil
			}
		}
		if filters.Assignees != nil {
			tasks = slices.FilterInPlace(tasks, func(t *model.Task) (ok bool) {
				return slices.IsSubsetGet(filters.Assignees, t.Assignees, GetUserID)
			})
			if len(tasks) < 1 {
				return nil, nil
			}
		}
		if filters.Reporters != nil {
			tasks = slices.FilterInPlace(tasks, func(t *model.Task) (ok bool) {
				return slices.IsSubsetGet(filters.Reporters, t.Reporters, GetUserID)
			})
			if len(tasks) < 1 {
				return nil, nil
			}
		}
		if filters.Tags != nil {
			tasks = slices.FilterInPlace(tasks, func(t *model.Task) (ok bool) {
				return slices.IsSubset(filters.Tags, t.Tags)
			})
			if len(tasks) < 1 {
				return nil, nil
			}
		}
		if filters.Status != nil {
			tasks = slices.FilterInPlace(tasks, func(t *model.Task) (ok bool) {
				return slices.Contains(filters.Status, t.Status)
			})
			if len(tasks) < 1 {
				return nil, nil
			}
		}
		if filters.Projects != nil {
			tasks = slices.FilterInPlace(tasks, func(t *model.Task) (ok bool) {
				return t.Project != nil &&
					slices.Contains(filters.Projects, t.Project.ID)
			})
			if len(tasks) < 1 {
				return nil, nil
			}
		}
	}
	return slices.SortAndLimit(tasks, SortFnTasks(order, orderAsc), Limit(limit)), nil
}

// Users is the resolver for the users field.
func (r *queryResolver) Users(ctx context.Context, filters *model.UsersFilters, order *model.UsersOrder, orderAsc bool, limit *int) ([]*model.User, error) {
	if err := auth.RequireAuthenticated(ctx); err != nil {
		return nil, err
	}

	users := slices.Copy(r.Resolver.Users)
	if filters != nil {
		if filters.Projects != nil {
			users = slices.FilterInPlace(users, func(u *model.User) (ok bool) {
				p := []string{}
				for _, t := range r.Resolver.Tasks {
					for _, a := range t.Assignees {
						if a == u {
							p = slices.AppendUnique(p, t.Project.ID)
							break
						}
					}
					for _, a := range t.Reporters {
						if a == u {
							p = slices.AppendUnique(p, t.Project.ID)
							break
						}
					}
				}
				return slices.IsSubset(filters.Projects, p)
			})
			if len(users) < 1 {
				return nil, nil
			}
		}
	}
	return slices.SortAndLimit(users, SortFnUsers(order, orderAsc), Limit(limit)), nil
}

// Projects is the resolver for the projects field.
func (r *queryResolver) Projects(ctx context.Context, filters *model.ProjectsFilters, order *model.ProjectsOrder, orderAsc bool, limit *int) ([]*model.Project, error) {
	if err := auth.RequireAuthenticated(ctx); err != nil {
		return nil, err
	}

	projects := slices.Copy(r.Resolver.Projects)
	if filters != nil {
		if filters.CreatedAfter != nil {
			projects = slices.FilterInPlace(projects, func(p *model.Project) (ok bool) {
				return p.Creation.Unix() > filters.CreatedAfter.Unix()
			})
			if len(projects) < 1 {
				return nil, nil
			}
		}
		if filters.CreatedBefore != nil {
			projects = slices.FilterInPlace(projects, func(p *model.Project) (ok bool) {
				return p.Creation.Unix() < filters.CreatedBefore.Unix()
			})
			if len(projects) < 1 {
				return nil, nil
			}
		}
		if filters.Members != nil {
			projects = slices.FilterInPlace(projects, func(p *model.Project) (ok bool) {
				memberIDs := []string{}
				for _, t := range r.Resolver.Tasks {
					if t.Project != p {
						continue
					}
					for _, u := range t.Assignees {
						memberIDs = slices.AppendUnique(memberIDs, u.ID)
					}
					for _, u := range t.Reporters {
						memberIDs = slices.AppendUnique(memberIDs, u.ID)
					}
				}
				return slices.IsSubset(filters.Members, memberIDs)
			})
			if len(projects) < 1 {
				return nil, nil
			}
		}
	}
	return slices.SortAndLimit(
		projects,
		SortFnProjects(order, orderAsc, r.Resolver),
		Limit(limit),
	), nil
}

// Query returns QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

type queryResolver struct{ *Resolver }
