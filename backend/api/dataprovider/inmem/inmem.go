package inmem

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/romshark/taskhub/api/auth"
	"github.com/romshark/taskhub/api/dataprovider"
	"github.com/romshark/taskhub/api/graph/model"
	"github.com/romshark/taskhub/slices"
)

var _ dataprovider.DataProvider = &Inmem{}

type Inmem struct {
	lock     sync.RWMutex
	Users    []*model.User
	Tasks    []*model.Task
	Projects []*model.Project
}

func (p *Inmem) UserByEmail(
	ctx context.Context, email string,
) (user *model.User, err error) {
	p.lock.RLock()
	defer p.lock.RUnlock()

	for _, user = range p.Users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, fmt.Errorf("user %q not found", email)
}

func (p *Inmem) UserByID(
	ctx context.Context, id string,
) (user *model.User, err error) {
	p.lock.RLock()
	defer p.lock.RUnlock()

	if user = p.userByID(id); user == nil {
		return user, fmt.Errorf("user %q not found", id)
	}
	return user, nil
}

func (p *Inmem) ProjectByID(
	ctx context.Context, id string,
) (project *model.Project, err error) {
	p.lock.RLock()
	defer p.lock.RUnlock()

	if project = p.projectByID(id); project == nil {
		return project, fmt.Errorf("project %q not found", id)
	}
	return project, nil
}

func (p *Inmem) TaskByID(
	ctx context.Context, id string,
) (task *model.Task, err error) {
	p.lock.RLock()
	defer p.lock.RUnlock()

	if task = p.taskByID(id); task == nil {
		return nil, fmt.Errorf("task %q not found", id)
	}
	return task, nil
}

func (p *Inmem) GetUsers(
	ctx context.Context,
	filters *model.UsersFilters,
	order *model.UsersOrder,
	orderAsc bool,
	limit *int,
) ([]*model.User, error) {
	p.lock.RLock()
	defer p.lock.RUnlock()

	users := slices.Copy(p.Users)
	if filters != nil {
		if filters.Projects != nil {
			users = slices.FilterInPlace(users, func(u *model.User) (ok bool) {
				projectIDs := []string{}
				for _, t := range p.Tasks {
					for _, a := range t.Assignees {
						if a == u {
							projectIDs = slices.AppendUnique(projectIDs, t.Project.ID)
							break
						}
					}
					for _, a := range t.Reporters {
						if a == u {
							projectIDs = slices.AppendUnique(projectIDs, t.Project.ID)
							break
						}
					}
				}
				return slices.IsSubset(filters.Projects, projectIDs)
			})
			if len(users) < 1 {
				return nil, nil
			}
		}
	}
	return slices.SortAndLimit(users, sortFnUsers(order, orderAsc), limitInt(limit)), nil
}

func (p *Inmem) GetProjects(
	ctx context.Context,
	filters *model.ProjectsFilters,
	order *model.ProjectsOrder,
	orderAsc bool,
	limit *int,
) ([]*model.Project, error) {
	p.lock.RLock()
	defer p.lock.RUnlock()

	projects := slices.Copy(p.Projects)
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
			projects = slices.FilterInPlace(projects, func(proj *model.Project) (ok bool) {
				memberIDs := []string{}
				for _, t := range p.Tasks {
					if t.Project != proj {
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
		p.sortFnProjects(order, orderAsc),
		limitInt(limit),
	), nil
}

func (p *Inmem) GetTasks(
	ctx context.Context,
	filters *model.TasksFilters,
	order *model.TasksOrder,
	orderAsc bool,
	limit *int,
) ([]*model.Task, error) {
	p.lock.RLock()
	defer p.lock.RUnlock()

	tasks := slices.Copy(p.Tasks)
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
				return slices.IsSubsetGet(filters.Assignees, t.Assignees, getUserID)
			})
			if len(tasks) < 1 {
				return nil, nil
			}
		}
		if filters.Reporters != nil {
			tasks = slices.FilterInPlace(tasks, func(t *model.Task) (ok bool) {
				return slices.IsSubsetGet(filters.Reporters, t.Reporters, getUserID)
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
	return slices.SortAndLimit(tasks, sortFnTasks(order, orderAsc), limitInt(limit)), nil
}

func (p *Inmem) CreateUser(ctx context.Context, email string, passwordHash string, displayName string, role string, location string, manager *string, subordinates []string) (*model.User, error) {
	for _, u := range p.Users {
		if u.DisplayName == displayName {
			return nil, errors.New("non-unique displayName")
		}
		if u.Email == email {
			return nil, errors.New("non-unique email")
		}
	}

	var managerUser *model.User
	if manager != nil {
		managerUser = p.userByID(*manager)
		if managerUser == nil {
			return nil, fmt.Errorf("manager user %q not found", *manager)
		}
	}

	var subordinateUsers []*model.User
	for _, s := range subordinates {
		u := p.userByID(s)
		if u == nil {
			return nil, fmt.Errorf("subordinate user %q not found", s)
		}
	}

	newUser := &model.User{
		ID:           "user_" + makeID(displayName),
		Email:        email,
		DisplayName:  displayName,
		Role:         role,
		Location:     location,
		Manager:      managerUser,
		Subordinates: subordinateUsers,
		PasswordHash: passwordHash,
	}
	p.Users = append(p.Users, newUser)
	return newUser, nil
}

func (p *Inmem) UpdateUser(ctx context.Context, id string, email string, displayName string, role string, location string, personalStatus *string, manager *string, subordinates []string) (*model.User, error) {
	p.lock.Lock()
	defer p.lock.Unlock()

	if err := auth.RequireAuthenticated(ctx); err != nil {
		return nil, err
	}

	user := p.userByID(id)
	if user == nil {
		return nil, fmt.Errorf("user %q not found", id)
	}

	if err := auth.RequireOwner(ctx, id); err != nil {
		return nil, err
	}

	for _, u := range p.Users {
		if u.DisplayName == displayName {
			return nil, errors.New("non-unique displayName")
		}
		if u.Email == email {
			return nil, errors.New("non-unique email")
		}
	}

	var personalStatusText string
	if personalStatus != nil {
		personalStatusText = *personalStatus
	}

	var managerUser *model.User
	if manager != nil {
		managerUser = p.userByID(*manager)
		if managerUser == nil {
			return nil, fmt.Errorf("manager user %q not found", *manager)
		}
		if managerUser == user {
			return nil, errors.New("user references itself as manager")
		}
	}

	var subordinateUsers []*model.User
	for _, s := range subordinates {
		u := p.userByID(s)
		if u == nil {
			return nil, fmt.Errorf("subordinate user %q not found", s)
		}
		if u == user {
			return nil, errors.New("user references itself as subordinate")
		}
		subordinateUsers = slices.AppendUnique(subordinateUsers, u)
	}

	user.Email = email
	user.DisplayName = displayName
	user.Role = role
	user.Location = location
	user.PersonalStatus = personalStatusText
	user.Manager = managerUser
	user.Subordinates = subordinateUsers

	return user, nil
}

func (p *Inmem) CreateTask(ctx context.Context, creation time.Time, title string, project string, status model.TaskStatus, priority model.TaskPriority, description *string, due *time.Time, tags []string, assignees []string, reporters []string, blocks []string, relatesTo []string) (*model.Task, error) {
	p.lock.Lock()
	defer p.lock.Unlock()

	if err := auth.RequireAuthenticated(ctx); err != nil {
		return nil, err
	}

	for _, t := range p.Tasks {
		if t.Title == title {
			return nil, errors.New("non-unique title")
		}
	}

	assignedProject := p.projectByID(project)
	if assignedProject == nil {
		return nil, fmt.Errorf("project %q not found", project)
	}

	var usersAssignees []*model.User
	for _, id := range assignees {
		user := p.userByID(id)
		if user == nil {
			return nil, fmt.Errorf("assignee user %q not found", id)
		}
		usersAssignees = slices.AppendUnique(usersAssignees, user)
	}

	var usersReporters []*model.User
	for _, id := range assignees {
		u := p.userByID(id)
		if u == nil {
			return nil, fmt.Errorf("reporter user %q not found", id)
		}
		usersReporters = slices.AppendUnique(usersReporters, u)
	}

	var blocksTasks []*model.Task
	for _, id := range blocks {
		t := p.taskByID(id)
		if t == nil {
			return nil, fmt.Errorf("blocked task %q not found", id)
		}
		blocksTasks = slices.AppendUnique(blocksTasks, t)
	}

	var relatesToTasks []*model.Task
	for _, id := range relatesTo {
		t := p.taskByID(id)
		if t == nil {
			return nil, fmt.Errorf("related task %q not found", id)
		}
		relatesToTasks = slices.AppendUnique(relatesToTasks, t)
	}

	newTask := &model.Task{
		ID:          "task_" + makeID(title),
		Title:       title,
		Description: description,
		Priority:    priority,
		Status:      status,
		Creation:    creation,
		Due:         due,
		Tags:        tags,
		Project:     assignedProject,
		Assignees:   usersAssignees,
		Reporters:   usersReporters,
		RelatesTo:   relatesToTasks,
		Blocks:      blocksTasks,
	}
	p.Tasks = append(p.Tasks, newTask)
	return newTask, nil
}

// UpdateTask is the resolver for the updateTask field.
func (p *Inmem) UpdateTask(
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
) (*model.Task, error) {
	p.lock.Lock()
	defer p.lock.Unlock()

	if err := auth.RequireAuthenticated(ctx); err != nil {
		return nil, err
	}

	task := p.taskByID(id)
	if task == nil {
		return nil, fmt.Errorf("task %q not found", id)
	}

	for _, t := range p.Tasks {
		if t.Title == title {
			return nil, errors.New("non-unique title")
		}
	}

	var assignedProject *model.Project
	for _, p := range p.Projects {
		if p.ID == project {
			assignedProject = p
			break
		}
	}
	if assignedProject == nil {
		return nil, fmt.Errorf("project %q not found", project)
	}

	var usersAssignees []*model.User
	for _, id := range assignees {
		user := p.userByID(id)
		if user == nil {
			return nil, fmt.Errorf("assignee user %q not found", id)
		}
		usersAssignees = slices.AppendUnique(usersAssignees, user)
	}

	var usersReporters []*model.User
	for _, id := range assignees {
		u := p.userByID(id)
		if u == nil {
			return nil, fmt.Errorf("reporter user %q not found", id)
		}
		usersReporters = slices.AppendUnique(usersReporters, u)
	}

	var blocksTasks []*model.Task
	for _, id := range blocks {
		t := p.taskByID(id)
		if t == nil {
			return nil, fmt.Errorf("blocked task %q not found", id)
		}
		if t == task {
			return nil, errors.New("task references itself as blocker")
		}
		blocksTasks = slices.AppendUnique(blocksTasks, t)
	}

	var relatesToTasks []*model.Task
	for _, id := range relatesTo {
		t := p.taskByID(id)
		if t == nil {
			return nil, fmt.Errorf("related task %q not found", id)
		}
		if t == task {
			return nil, errors.New("task references itself as related")
		}
		relatesToTasks = slices.AppendUnique(relatesToTasks, t)
	}

	task.Status = status
	task.Priority = priority
	task.Description = description
	task.Tags = tags
	task.Due = due
	task.Reporters = usersReporters
	task.Title = title
	task.Project = assignedProject
	task.Assignees = usersAssignees
	task.Blocks = blocksTasks
	task.RelatesTo = relatesToTasks

	return task, nil
}

// CreateProject is the resolver for the createProject field.
func (p *Inmem) CreateProject(
	ctx context.Context,
	creation time.Time,
	name string,
	description string,
	slug string,
	owners []string,
) (*model.Project, error) {
	p.lock.Lock()
	defer p.lock.Unlock()

	if err := auth.RequireAuthenticated(ctx); err != nil {
		return nil, err
	}

	for _, p := range p.Projects {
		if p.Name == name {
			return nil, errors.New("non-unique project name")
		}
		if p.Slug == slug {
			return nil, errors.New("non-unique project slug")
		}
	}

	var ownerUsers []*model.User
	for _, id := range owners {
		u := p.userByID(id)
		if u == nil {
			return nil, fmt.Errorf("owner user %q not found", id)
		}
		ownerUsers = slices.AppendUnique(ownerUsers, u)
	}

	newProject := &model.Project{
		ID:          "project_" + makeID(name),
		Name:        name,
		Description: description,
		Slug:        slug,
		Creation:    creation,
		Owners:      ownerUsers,
	}
	return newProject, nil
}

// UpdateProject is the resolver for the updateProject field.
func (p *Inmem) UpdateProject(ctx context.Context, id string, name string, description string, slug string, owners []string) (*model.Project, error) {
	p.lock.Lock()
	defer p.lock.Unlock()

	if err := auth.RequireAuthenticated(ctx); err != nil {
		return nil, err
	}

	project := p.projectByID(id)
	if project == nil {
		return nil, fmt.Errorf("project %q not found", id)
	}

	for _, p := range p.Projects {
		if p.Name == name {
			return nil, errors.New("non-unique project name")
		}
		if p.Slug == slug {
			return nil, errors.New("non-unique project slug")
		}
	}

	var ownerUsers []*model.User
	for _, id := range owners {
		u := p.userByID(id)
		if u == nil {
			return nil, fmt.Errorf("owner user %q not found", id)
		}
		ownerUsers = slices.AppendUnique(ownerUsers, u)
	}

	project.Name = name
	project.Description = description
	project.Slug = slug
	project.Owners = ownerUsers

	return project, nil
}

func (p *Inmem) GetProjectMembers(
	ctx context.Context,
	projectID string,
) ([]*model.User, error) {
	m := []*model.User{}
	for _, t := range p.Tasks {
		if t.Project.ID != projectID {
			continue
		}
		for _, u := range t.Assignees {
			m = slices.AppendUnique(m, u)
		}
		for _, u := range t.Reporters {
			m = slices.AppendUnique(m, u)
		}
	}
	return m, nil
}

func (p *Inmem) GetBlockingTasks(
	ctx context.Context,
	taskID string,
) ([]*model.Task, error) {
	task := p.taskByID(taskID)
	if task == nil {
		return nil, fmt.Errorf("task %q not found", taskID)
	}

	blockedBy := []*model.Task{}
	for _, t := range p.Tasks {
		if t == task {
			continue
		}
		for _, tb := range t.Blocks {
			if tb == task {
				blockedBy = slices.AppendUnique(blockedBy, t)
			}
		}
	}
	return blockedBy, nil
}

func (p *Inmem) GetRelatedTasks(
	ctx context.Context,
	taskID string,
) ([]*model.Task, error) {
	task := p.taskByID(taskID)
	if task == nil {
		return nil, fmt.Errorf("task %q not found", taskID)
	}

	relatesTo := []*model.Task{}
	relatesTo = append(relatesTo, task.RelatesTo...)
	for _, t := range p.Tasks {
		if t == task {
			continue
		}
		for _, tr := range t.RelatesTo {
			if tr == task {
				relatesTo = slices.AppendUnique(relatesTo, t)
			}
		}
	}
	return relatesTo, nil
}

func (p *Inmem) GetTasksByProject(
	ctx context.Context,
	projectID string,
) ([]*model.Task, error) {
	project := p.projectByID(projectID)
	if project == nil {
		return nil, fmt.Errorf("project %q not found", projectID)
	}

	tasks := []*model.Task{}
	for _, t := range p.Tasks {
		if t.Project == project {
			tasks = append(tasks, t)
		}
	}
	return tasks, nil
}

func (p *Inmem) GetUserProjects(
	ctx context.Context,
	userID string,
) ([]*model.Project, error) {
	user := p.userByID(userID)
	if user == nil {
		return nil, fmt.Errorf("user %q not found", userID)
	}

	projects := []*model.Project{}
	for _, t := range p.Tasks {
		for _, u := range t.Assignees {
			if u == user {
				projects = slices.AppendUnique(projects, t.Project)
				break
			}
		}
		for _, u := range t.Reporters {
			if u == user {
				projects = slices.AppendUnique(projects, t.Project)
				break
			}
		}
	}
	return projects, nil
}

func (p *Inmem) GetTasksAssignedToUser(
	ctx context.Context,
	userID string,
) ([]*model.Task, error) {
	user := p.userByID(userID)
	if user == nil {
		return nil, fmt.Errorf("user %q not found", userID)
	}

	tasks := []*model.Task{}
	for _, t := range p.Tasks {
		for _, r := range t.Assignees {
			if r == user {
				tasks = append(tasks, t)
			}
		}
	}
	return tasks, nil
}

func (p *Inmem) GetTasksReportedByUser(
	ctx context.Context,
	userID string,
) ([]*model.Task, error) {
	user := p.userByID(userID)
	if user == nil {
		return nil, fmt.Errorf("user %q not found", userID)
	}

	tasks := []*model.Task{}
	for _, t := range p.Tasks {
		for _, r := range t.Reporters {
			if r == user {
				tasks = append(tasks, t)
			}
		}
	}
	return tasks, nil
}

func (p *Inmem) userByID(id string) *model.User {
	for _, x := range p.Users {
		if x.ID == id {
			return x
		}
	}
	return nil
}

func (p *Inmem) projectByID(id string) *model.Project {
	for _, x := range p.Projects {
		if x.ID == id {
			return x
		}
	}
	return nil
}

func (p *Inmem) taskByID(id string) *model.Task {
	for _, x := range p.Tasks {
		if x.ID == id {
			return x
		}
	}
	return nil
}

// limitInt returns -1 if n == null, otherwise returns the value of n.
func limitInt(n *int) int {
	if n == nil {
		return -1
	}
	return *n
}

// taskPriorityScalar returns the scalar value of a task priority enum value.
func taskPriorityScalar(p model.TaskPriority) int {
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

// makeID trims spaces, replaces all whitespace sequences with underscores,
// and converts the result to lower case characters.
func makeID(name string) string {
	s := strings.TrimSpace(name)
	f := strings.Fields(s)
	s = strings.Join(f, "_")
	return strings.ToLower(s)
}

func getUserID(u *model.User) string { return u.ID }

func sortFnUsers(order *model.UsersOrder, asc bool) func(a, b *model.User) bool {
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

func (p *Inmem) sortFnProjects(
	order *model.ProjectsOrder,
	asc bool,
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
			for _, t := range p.Tasks {
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
			for _, t := range p.Tasks {
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

func sortFnTasks(order *model.TasksOrder, asc bool) func(a, b *model.Task) bool {
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
				return taskPriorityScalar(a.Priority) < taskPriorityScalar(b.Priority)
			}
			return taskPriorityScalar(a.Priority) > taskPriorityScalar(b.Priority)
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
