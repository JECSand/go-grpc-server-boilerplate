package services

import (
	"bytes"
	"context"
	"github.com/JECSand/go-grpc-server-boilerplate/models"
	"github.com/JECSand/go-grpc-server-boilerplate/utilities"
)

// UserDataService is an interface to database.UserService
type UserDataService interface {
	AuthenticateUser(u *models.User) (*models.User, error)
	UpdatePassword(u *models.User, CurrentPassword string, newPassword string) (*models.User, error)
	UserCreate(u *models.User) (*models.User, error)
	UserDelete(u *models.User) (*models.User, error)
	UserDeleteMany(u *models.User) (*models.User, error)
	UsersFind(u *models.User) ([]*models.User, error)
	UserFind(u *models.User) (*models.User, error)
	UserUpdate(u *models.User) (*models.User, error)
	UserDocInsert(u *models.User) (*models.User, error)
	UsersQuery(ctx context.Context, query string, pagination *utilities.Pagination) (*models.UsersRes, error)
}

// GroupDataService is an interface to database.GroupService
type GroupDataService interface {
	GroupCreate(g *models.Group) (*models.Group, error)
	GroupFind(g *models.Group) (*models.Group, error)
	GroupsFind(g *models.Group) ([]*models.Group, error)
	GroupDelete(g *models.Group) (*models.Group, error)
	GroupDeleteMany(g *models.Group) (*models.Group, error)
	GroupUpdate(g *models.Group) (*models.Group, error)
	GroupDocInsert(g *models.Group) (*models.Group, error)
	GroupsQuery(ctx context.Context, query string, pagination *utilities.Pagination) (*models.GroupsRes, error)
}

// TaskDataService is an interface to database.TaskService
type TaskDataService interface {
	TaskCreate(g *models.Task) (*models.Task, error)
	TaskFind(g *models.Task) (*models.Task, error)
	TasksFind(g *models.Task) ([]*models.Task, error)
	TaskDelete(g *models.Task) (*models.Task, error)
	TaskDeleteMany(g *models.Task) (*models.Task, error)
	TaskUpdate(g *models.Task) (*models.Task, error)
	TaskDocInsert(g *models.Task) (*models.Task, error)
	TasksQuery(ctx context.Context, query string, pagination *utilities.Pagination) (*models.TasksRes, error)
}

// FileDataService is an interface to database.FileService
type FileDataService interface {
	FileCreate(g *models.File, content []byte) (*models.File, error)
	FileFind(g *models.File) (*models.File, error)
	FilesFind(g *models.File) ([]*models.File, error)
	FileDelete(g *models.File) (*models.File, error)
	FileDeleteMany(g []*models.File) error
	FileUpdate(g *models.File, content []byte) (*models.File, error)
	RetrieveFile(g *models.File) (*bytes.Buffer, error)
}

// BlacklistDataService is an interface to database.BlacklistService
type BlacklistDataService interface {
	BlacklistAuthToken(authToken string) error
	CheckTokenBlacklist(authToken string) bool
}
