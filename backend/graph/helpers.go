package graph

import (
	"strings"

	"github.com/romshark/taskhub/graph/model"
	"github.com/romshark/taskhub/slices"
)

func GetUserID(u *model.User) string               { return u.ID }
func GetProjectID(u *model.Project) string         { return u.ID }
func GetTaskProjectID(u *model.Task) string        { return u.Project.ID }
func GetTaskAssignees(t *model.Task) []*model.User { return t.Assignees }

func SortFnUsers(order *model.UsersOrder, asc bool) func(a, b *model.User) bool {
	if order == nil {
		return nil
	}
	switch *order {
	case model.UsersOrderNameAlpha:
		return func(a, b *model.User) bool {
			if asc {
				return a.DisplayName < b.DisplayName
			}
			return a.DisplayName > b.DisplayName
		}
	}
	return nil
}

func SortFnProjects(
	order *model.ProjectsOrder,
	asc bool,
	resolver *Resolver,
) func(a, b *model.Project) bool {
	if order == nil {
		return nil
	}
	switch *order {
	case model.ProjectsOrderNameAlpha:
		return func(a, b *model.Project) bool {
			if asc {
				return a.Name < b.Name
			}
			return a.Name > b.Name
		}
	case model.ProjectsOrderNumMembers:
		return func(a, b *model.Project) bool {
			am, bm := []string{}, []string{}
			for _, t := range resolver.Tasks {
				switch t.Project {
				case a:
					for _, a := range t.Assignees {
						am = slices.AppendUnique(am, a.ID)
					}
					for _, a := range t.Reporters {
						am = slices.AppendUnique(am, a.ID)
					}
				case b:
					for _, a := range t.Assignees {
						bm = slices.AppendUnique(bm, a.ID)
					}
					for _, a := range t.Reporters {
						bm = slices.AppendUnique(bm, a.ID)
					}
				}
			}
			if asc {
				return len(am) < len(bm)
			}
			return len(am) > len(bm)
		}
	case model.ProjectsOrderNumTasks:
		return func(a, b *model.Project) bool {
			countA, countB := 0, 0
			for _, t := range resolver.Tasks {
				switch t.Project {
				case a:
					countA++
				case b:
					countB++
				}
			}
			if asc {
				return countA < countB
			}
			return countA > countB
		}
	}
	return nil
}

func SortFnTasks(order *model.TasksOrder, asc bool) func(a, b *model.Task) bool {
	if order == nil {
		return nil
	}
	switch *order {
	case model.TasksOrderTitleAlpha:
		return func(a, b *model.Task) bool {
			if asc {
				return a.Title < b.Title
			}
			return a.Title > b.Title
		}
	case model.TasksOrderPriority:
		return func(a, b *model.Task) bool {
			if asc {
				return TaskPriorityScalar(a.Priority) < TaskPriorityScalar(b.Priority)
			}
			return TaskPriorityScalar(a.Priority) > TaskPriorityScalar(b.Priority)
		}
	case model.TasksOrderDueTime:
		return func(a, b *model.Task) bool {
			if a.Due == nil || b.Due == nil {
				return asc
			}
			if asc {
				return a.Due.Unix() < b.Due.Unix()
			}
			return a.Due.Unix() > b.Due.Unix()
		}
	case model.TasksOrderCreationTime:
		return func(a, b *model.Task) bool {
			if asc {
				return a.Creation.Unix() < b.Creation.Unix()
			}
			return a.Creation.Unix() > b.Creation.Unix()
		}
	}
	return nil
}

// Limit returns -1 if n == null, otherwise returns the value of n.
func Limit(n *int) int {
	if n == nil {
		return -1
	}
	return *n
}

// TaskPriorityScalar returns the scalar value of a task priority enum value.
func TaskPriorityScalar(p model.TaskPriority) int {
	switch p {
	case model.TaskPriorityBlocker:
		return 4
	case model.TaskPriorityHigh:
		return 3
	case model.TaskPriorityMedium:
		return 2
	case model.TaskPriorityLow:
		return 1
	}
	return 0
}

// MakeID trims spaces, replaces all whitespace sequences with underscores,
// and converts the result to lower case characters.
func MakeID(name string) string {
	s := strings.TrimSpace(name)
	f := strings.Fields(s)
	s = strings.Join(f, "_")
	return strings.ToLower(s)
}

func ProjectByID(resolver *Resolver, id string) *model.Project {
	for i := range resolver.Projects {
		if resolver.Projects[i].ID == id {
			return resolver.Projects[i]
		}
	}
	return nil
}

func TaskByID(resolver *Resolver, id string) *model.Task {
	for i := range resolver.Tasks {
		if resolver.Tasks[i].ID == id {
			return resolver.Tasks[i]
		}
	}
	return nil
}

func UserByID(resolver *Resolver, id string) *model.User {
	for i := range resolver.Users {
		if resolver.Users[i].ID == id {
			return resolver.Users[i]
		}
	}
	return nil
}
