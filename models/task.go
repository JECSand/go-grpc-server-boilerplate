package models

import (
	"errors"
	tasksService "github.com/JECSand/go-grpc-server-boilerplate/protos/task"
	"github.com/JECSand/go-grpc-server-boilerplate/utilities"
	"google.golang.org/protobuf/types/known/timestamppb"
	"strings"
	"time"
)

type TaskStatus int

const (
	UNSPECIFIED TaskStatus = iota
	NOT_STARTED
	IN_PROGRESS
	COMPLETED
)

// Task is a root struct that is used to store the json encoded data for/from a mongodb group doc.
type Task struct {
	Id           string     `json:"id,omitempty"`
	Name         string     `json:"name,omitempty"`
	Status       TaskStatus `json:"status,omitempty"`
	Due          time.Time  `json:"due,omitempty"`
	Description  string     `json:"description,omitempty"`
	UserId       string     `json:"user_id,omitempty"`
	GroupId      string     `json:"group_id,omitempty"`
	LastModified time.Time  `json:"last_modified,omitempty"`
	CreatedAt    time.Time  `json:"created_at,omitempty"`
	DeletedAt    time.Time  `json:"deleted_at,omitempty"`
}

// ToProto Convert Task to proto
func (g *Task) ToProto() *tasksService.Task {
	return &tasksService.Task{
		Id:           g.Id,
		Name:         g.Name,
		Status:       tasksService.TaskStatus(g.Status),
		Due:          timestamppb.New(g.Due),
		Description:  g.Description,
		UserId:       g.UserId,
		GroupId:      g.GroupId,
		LastModified: timestamppb.New(g.LastModified),
		CreatedAt:    timestamppb.New(g.CreatedAt),
		DeletedAt:    timestamppb.New(g.DeletedAt),
	}
}

// LoadTaskProto inputs a usersService and returns a Task
func LoadTaskProto(u *tasksService.Task) *Task {
	return &Task{
		Id:           u.GetId(),
		Name:         u.GetName(),
		Status:       TaskStatus(u.GetStatus().Number()),
		Due:          u.GetDue().AsTime(),
		Description:  u.GetDescription(),
		UserId:       u.GetUserId(),
		GroupId:      u.GetGroupId(),
		LastModified: u.GetLastModified().AsTime(),
		CreatedAt:    u.GetCreatedAt().AsTime(),
		DeletedAt:    u.GetDeletedAt().AsTime(),
	}
}

// LoadTaskCreateProto inputs a tasksService.CreateReq and returns a Task
func LoadTaskCreateProto(u *tasksService.CreateReq) *Task {
	return &Task{
		Name:        u.GetName(),
		Due:         u.GetDue().AsTime(),
		Description: u.GetDescription(),
		UserId:      u.GetUserId(),
		GroupId:     u.GetGroupId(),
	}
}

// LoadTaskUpdateProto inputs a tasksService.UpdateReq and returns a Task
func LoadTaskUpdateProto(u *tasksService.UpdateReq) *Task {
	return &Task{
		Id:          u.GetId(),
		Name:        u.GetName(),
		Status:      TaskStatus(u.GetStatus().Number()),
		Due:         u.GetDue().AsTime(),
		Description: u.GetDescription(),
		UserId:      u.GetUserId(),
		GroupId:     u.GetGroupId(),
	}
}

// LoadScope scopes the Task struct
func (g *Task) LoadScope(scopeUser *User) {
	if !scopeUser.RootAdmin {
		g.GroupId = scopeUser.GroupId
		if scopeUser.Role != "admin" {
			g.UserId = scopeUser.Id
		}
	}
	if !g.CheckID("user_id") {
		g.UserId = scopeUser.Id
	}
	if !g.CheckID("group_id") {
		g.GroupId = scopeUser.GroupId
	}
	return
}

// CheckID determines whether a specified ID is set or not
func (g *Task) CheckID(chkId string) bool {
	switch chkId {
	case "id":
		if !utilities.CheckObjectID(g.Id) {
			return false
		}
	case "group_id":
		if !utilities.CheckObjectID(g.GroupId) {
			return false
		}
	case "user_id":
		if !utilities.CheckObjectID(g.UserId) {
			return false
		}
	}
	return true
}

// Validate a Group for different scenarios such as loading TokenData, creating new Group, or updating a Group
func (g *Task) Validate(valCase string) (err error) {
	var missingFields []string
	switch valCase {
	case "create":
		if g.Name == "" {
			missingFields = append(missingFields, "name")
		}
		if !g.CheckID("user_id") {
			missingFields = append(missingFields, "user_id")
		}
		if !g.CheckID("group_id") {
			missingFields = append(missingFields, "group_id")
		}
		if g.Due.IsZero() {
			missingFields = append(missingFields, "due")
		}
	case "update":
		if !g.CheckID("id") {
			missingFields = append(missingFields, "id")
		}
	default:
		return errors.New("unrecognized validation case")
	}
	if len(missingFields) > 0 {
		return errors.New("missing the following group fields: " + strings.Join(missingFields, ", "))
	}
	return
}

// BuildUpdate is a function that setups the base task struct during a user modification request
func (g *Task) BuildUpdate(cur *Task) {
	if len(g.Name) == 0 {
		g.Name = cur.Name
	}
	if g.Status == UNSPECIFIED {
		g.Status = cur.Status
	}
	if g.Due.IsZero() {
		g.Due = cur.Due
	}
	if len(g.Description) == 0 {
		g.Description = cur.Description
	}
	if len(g.UserId) == 0 {
		g.UserId = cur.UserId
	}
	if len(g.GroupId) == 0 {
		g.GroupId = cur.GroupId
	}
}

// TasksRes Multiple Tasks in a paginated response
type TasksRes struct {
	TotalCount int64   `json:"total_count"`
	TotalPages int64   `json:"total_pages"`
	Page       int64   `json:"page"`
	Size       int64   `json:"size"`
	HasMore    bool    `json:"has_more"`
	Tasks      []*Task `json:"tasks"`
}

// ToProto convert TasksRes to proto
func (p *TasksRes) ToProto() []*tasksService.Task {
	uList := make([]*tasksService.Task, 0, len(p.Tasks))
	for _, u := range p.Tasks {
		uList = append(uList, u.ToProto())
	}
	return uList
}
