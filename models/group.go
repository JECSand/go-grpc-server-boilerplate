package models

import (
	"errors"
	groupsService "github.com/JECSand/go-grpc-server-boilerplate/protos/group"
	"github.com/JECSand/go-grpc-server-boilerplate/utilities"
	"google.golang.org/protobuf/types/known/timestamppb"
	"strings"
	"time"
)

// Group is a root struct that is used to store the json encoded data for/from a mongodb group doc.
type Group struct {
	Id           string    `json:"id,omitempty"`
	Name         string    `json:"name,omitempty"`
	RootAdmin    bool      `json:"root_admin,omitempty"`
	LastModified time.Time `json:"last_modified,omitempty"`
	CreatedAt    time.Time `json:"created_at,omitempty"`
	DeletedAt    time.Time `json:"deleted_at,omitempty"`
}

// ToProto Convert Group to proto
func (g *Group) ToProto() *groupsService.Group {
	return &groupsService.Group{
		Id:           g.Id,
		Name:         g.Name,
		RootAdmin:    g.RootAdmin,
		LastModified: timestamppb.New(g.LastModified),
		CreatedAt:    timestamppb.New(g.CreatedAt),
		DeletedAt:    timestamppb.New(g.DeletedAt),
	}
}

// LoadGroupProto inputs a groupsService and returns a Group
func LoadGroupProto(u *groupsService.Group) *Group {
	return &Group{
		Id:           u.GetId(),
		Name:         u.GetName(),
		RootAdmin:    u.GetRootAdmin(),
		LastModified: u.GetLastModified().AsTime(),
		CreatedAt:    u.GetCreatedAt().AsTime(),
		DeletedAt:    u.GetDeletedAt().AsTime(),
	}
}

// CheckID determines whether a specified ID is set or not
func (g *Group) CheckID(chkId string) bool {
	switch chkId {
	case "id":
		if !utilities.CheckObjectID(g.Id) {
			return false
		}
	}
	return true
}

// Validate a Group for different scenarios such as loading TokenData, creating new Group, or updating a Group
func (g *Group) Validate(valCase string) (err error) {
	var missingFields []string
	switch valCase {
	case "create":
		if g.Name == "" {
			missingFields = append(missingFields, "name")
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

// GroupsRes Multiple Groups in a paginated response
type GroupsRes struct {
	TotalCount int64    `json:"total_count"`
	TotalPages int64    `json:"total_pages"`
	Page       int64    `json:"page"`
	Size       int64    `json:"size"`
	HasMore    bool     `json:"has_more"`
	Groups     []*Group `json:"groups"`
}

// ToProto convert GroupsRes to proto
func (p *GroupsRes) ToProto() []*groupsService.Group {
	uList := make([]*groupsService.Group, 0, len(p.Groups))
	for _, u := range p.Groups {
		uList = append(uList, u.ToProto())
	}
	return uList
}
