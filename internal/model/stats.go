package model

import "github.com/google/uuid"

// AssigneeTaskCount captures how many tasks belong to an assignee in a project.
type AssigneeTaskCount struct {
	AssigneeID *uuid.UUID `json:"assignee_id"`
	Count      int        `json:"count"`
}

// ProjectStats contains aggregate task metrics for a project.
type ProjectStats struct {
	ByStatus   map[string]int      `json:"by_status"`
	ByAssignee []AssigneeTaskCount `json:"by_assignee"`
}
