package model

import "time"

type User struct {
	ID             string  `json:"id"`
	DisplayName    string  `json:"displayName"`
	Role           string  `json:"role"`
	Location       string  `json:"location"`
	PersonalStatus string  `json:"personalStatus"`
	Manager        *User   `json:"manager,omitempty"`
	Subordinates   []*User `json:"subordinates,omitempty"`
}

type Project struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Slug        string    `json:"slug"`
	Creation    time.Time `json:"creation"`
	Owners      []*User   `json:"owners,omitempty"`
}

type Task struct {
	ID          string       `json:"id"`
	Title       string       `json:"title"`
	Description *string      `json:"description,omitempty"`
	Priority    TaskPriority `json:"priority"`
	Status      TaskStatus   `json:"status"`
	Creation    time.Time    `json:"creation"`
	Due         *time.Time   `json:"due,omitempty"`
	Tags        []string     `json:"tags"`
	Project     *Project     `json:"project"`
	Assignees   []*User      `json:"assignees"`
	Reporters   []*User      `json:"reporters"`
	Blocks      []*Task      `json:"blocks"`
	RelatesTo   []*Task      `json:"relatesTo"`
}
