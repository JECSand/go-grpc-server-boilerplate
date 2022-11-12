package models

import (
	"errors"
	authService "github.com/JECSand/go-grpc-server-boilerplate/protos/auth"
	usersService "github.com/JECSand/go-grpc-server-boilerplate/protos/user"
	"github.com/JECSand/go-grpc-server-boilerplate/utilities"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/protobuf/types/known/timestamppb"
	"strings"
	"time"
)

// User is a root struct that is used to store the json encoded data for/from a mongodb user doc.
type User struct {
	Id           string    `json:"id,omitempty"`
	Username     string    `json:"username,omitempty"`
	Password     string    `json:"password,omitempty"`
	FirstName    string    `json:"firstname,omitempty"`
	LastName     string    `json:"lastname,omitempty"`
	Email        string    `json:"email,omitempty"`
	Role         string    `json:"role,omitempty"`
	RootAdmin    bool      `json:"root_admin,omitempty"`
	GroupId      string    `json:"group_id,omitempty"`
	ImageId      string    `json:"image_id,omitempty"`
	LastModified time.Time `json:"last_modified,omitempty"`
	CreatedAt    time.Time `json:"created_at,omitempty"`
	DeletedAt    time.Time `json:"deleted_at,omitempty"`
}

// ToProto Convert User to proto
func (g *User) ToProto() *usersService.User {
	return &usersService.User{
		Id:           g.Id,
		Username:     g.Username,
		Password:     g.Password,
		FirstName:    g.FirstName,
		LastName:     g.LastName,
		Email:        g.Email,
		Role:         g.Role,
		RootAdmin:    g.RootAdmin,
		GroupId:      g.GroupId,
		ImageId:      g.ImageId,
		LastModified: timestamppb.New(g.LastModified),
		CreatedAt:    timestamppb.New(g.CreatedAt),
		DeletedAt:    timestamppb.New(g.DeletedAt),
	}
}

// ToAuthProto Convert User to auth proto
func (g *User) ToAuthProto() *authService.User {
	return &authService.User{
		Id:           g.Id,
		Username:     g.Username,
		FirstName:    g.FirstName,
		LastName:     g.LastName,
		Email:        g.Email,
		Role:         g.Role,
		RootAdmin:    g.RootAdmin,
		GroupId:      g.GroupId,
		ImageId:      g.ImageId,
		LastModified: timestamppb.New(g.LastModified),
		CreatedAt:    timestamppb.New(g.CreatedAt),
		DeletedAt:    timestamppb.New(g.DeletedAt),
	}
}

// LoadUserProto inputs a usersService.User and returns a User
func LoadUserProto(u *usersService.User) *User {
	return &User{
		Id:           u.GetId(),
		Username:     u.GetUsername(),
		Password:     u.GetPassword(),
		FirstName:    u.GetFirstName(),
		LastName:     u.GetLastName(),
		Email:        u.GetEmail(),
		Role:         u.GetRole(),
		RootAdmin:    u.GetRootAdmin(),
		GroupId:      u.GetGroupId(),
		ImageId:      u.GetImageId(),
		LastModified: u.GetLastModified().AsTime(),
		CreatedAt:    u.GetCreatedAt().AsTime(),
		DeletedAt:    u.GetDeletedAt().AsTime(),
	}
}

// LoadUserCreateProto inputs a usersService.CreateReq and returns a User
func LoadUserCreateProto(u *usersService.CreateReq) *User {
	return &User{
		Username:  u.GetUsername(),
		Password:  u.GetPassword(),
		FirstName: u.GetFirstName(),
		LastName:  u.GetLastName(),
		Email:     u.GetEmail(),
		Role:      u.GetRole(),
		RootAdmin: u.GetRootAdmin(),
		GroupId:   u.GetGroupId(),
	}
}

// LoadUserUpdateProto inputs a usersService.UpdateReq and returns a User
func LoadUserUpdateProto(u *usersService.UpdateReq) *User {
	return &User{
		Id:        u.GetId(),
		Username:  u.GetUsername(),
		Password:  u.GetPassword(),
		FirstName: u.GetFirstName(),
		LastName:  u.GetLastName(),
		Email:     u.GetEmail(),
		Role:      u.GetRole(),
		RootAdmin: u.GetRootAdmin(),
		GroupId:   u.GetGroupId(),
	}
}

// LoadScope scopes the User struct
func (g *User) LoadScope(scopeUser *User, valCase string) {
	switch valCase {
	case "create":
		if !scopeUser.RootAdmin {
			g.RootAdmin = false
			g.GroupId = scopeUser.GroupId
		}
	case "update":
		g.Id = scopeUser.Id
		if !scopeUser.RootAdmin {
			g.RootAdmin = false
			g.GroupId = scopeUser.GroupId
			if scopeUser.Role != "admin" {
				g.Role = "member"
			}
		}
	case "find":
		if !scopeUser.RootAdmin {
			g.GroupId = scopeUser.GroupId
		}
	}
	return
}

// CheckID determines whether a specified ID is set or not
func (g *User) CheckID(chkId string) bool {
	switch chkId {
	case "id":
		if !utilities.CheckObjectID(g.Id) {
			return false
		}
	case "group_id":
		if !utilities.CheckObjectID(g.GroupId) {
			return false
		}
	case "image_id":
		if !utilities.CheckObjectID(g.ImageId) {
			return false
		}
	}
	return true
}

// Authenticate compares an input password with the hashed password stored in the User model
func (g *User) Authenticate(checkPassword string) error {
	if len(g.Password) != 0 {
		password := []byte(g.Password)
		cPassword := []byte(checkPassword)
		return bcrypt.CompareHashAndPassword(password, cPassword)
	}
	return errors.New("no password set to hash in user model")
}

// HashPassword hashes a user password and associates it with the user struct
func (g *User) HashPassword() error {
	if len(g.Password) != 0 {
		password := []byte(g.Password)
		hashedPassword, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		g.Password = string(hashedPassword)
		return nil
	}
	return errors.New("no password set to hash in user model")
}

// Validate a User for different scenarios such as loading TokenData, creating new User, or updating a User
func (g *User) Validate(valCase string) (err error) {
	var missingFields []string
	switch valCase {
	case "login":
		if g.Email == "" {
			missingFields = append(missingFields, "email")
		}
		if g.Password == "" {
			missingFields = append(missingFields, "password")
		}
	case "auth":
		if !g.CheckID("id") {
			missingFields = append(missingFields, "id")
		}
		if !g.CheckID("group_id") {
			missingFields = append(missingFields, "group_id")
		}
		if g.Role == "" {
			missingFields = append(missingFields, "role")
		}
	case "create":
		if g.Username == "" {
			missingFields = append(missingFields, "id")
		}
		if g.Email == "" {
			missingFields = append(missingFields, "email")
		}
		if g.Password == "" {
			missingFields = append(missingFields, "password")
		}
		if !g.CheckID("group_id") {
			missingFields = append(missingFields, "group_id")
		}
	case "update":
		if !g.CheckID("id") && g.Email == "" {
			missingFields = append(missingFields, "id")
		}
	default:
		return errors.New("unrecognized validation case")
	}
	if len(missingFields) > 0 {
		return errors.New("missing the following user fields: " + strings.Join(missingFields, ", "))
	}
	return
}

// BuildFilter is a function that setups the base user struct during a user modification request
func (g *User) BuildFilter() (*User, error) {
	var filter User
	if g.CheckID("id") {
		filter.Id = g.Id
	} else if g.Email != "" {
		filter.Email = g.Email
	} else {
		return nil, errors.New("user is missing a valid query filter")
	}
	return &filter, nil
}

// BuildUpdate is a function that setups the base user struct during a user modification request
func (g *User) BuildUpdate(curUser *User) {
	if len(g.Username) == 0 {
		g.Username = curUser.Username
	}
	if len(g.FirstName) == 0 {
		g.FirstName = curUser.FirstName
	}
	if len(g.LastName) == 0 {
		g.LastName = curUser.LastName
	}
	if len(g.Email) == 0 {
		g.Email = curUser.Email
	}
	if len(g.GroupId) == 0 {
		g.GroupId = curUser.GroupId
	}
	if len(g.ImageId) == 0 {
		g.ImageId = curUser.ImageId
	}
	if len(g.Role) == 0 {
		g.Role = curUser.Role
	}
}

// UsersToFiles converts an input slice of user to a slice of file
func UsersToFiles(users []*User) []*File {
	var files []*File
	for _, u := range users {
		if u.CheckID("image_id") {
			files = append(files, &File{OwnerId: u.Id, OwnerType: "user"})
		}
	}
	return files
}

// UsersRes Multiple Users in a paginated response
type UsersRes struct {
	TotalCount int64   `json:"total_count"`
	TotalPages int64   `json:"total_pages"`
	Page       int64   `json:"page"`
	Size       int64   `json:"size"`
	HasMore    bool    `json:"has_more"`
	Users      []*User `json:"users"`
}

// ToProto convert UsersRes to proto
func (p *UsersRes) ToProto() []*usersService.User {
	uList := make([]*usersService.User, 0, len(p.Users))
	for _, u := range p.Users {
		uList = append(uList, u.ToProto())
	}
	return uList
}
