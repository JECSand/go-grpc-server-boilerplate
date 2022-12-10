package services

import (
	"context"
	"errors"
	"github.com/JECSand/go-grpc-server-boilerplate/models"
	groupsService "github.com/JECSand/go-grpc-server-boilerplate/protos/group"
	"github.com/JECSand/go-grpc-server-boilerplate/utilities"
)

// GroupService gRPC Service
type GroupService struct {
	log          utilities.Logger
	tokenService *TokenService
	userDB       UserDataService
	groupDB      GroupDataService
	taskDB       TaskDataService
	fileDB       FileDataService
}

// NewGroupService constructs a GroupService for controller gRPC service Group requests
func NewGroupService(log utilities.Logger, ts *TokenService, u UserDataService, g GroupDataService, t TaskDataService, f FileDataService) *GroupService {
	return &GroupService{
		log:          log,
		tokenService: ts,
		userDB:       u,
		groupDB:      g,
		taskDB:       t,
		fileDB:       f,
	}
}

// Create is a New Group
func (u *GroupService) Create(ctx context.Context, req *groupsService.CreateReq) (*groupsService.CreateRes, error) {
	group := models.LoadGroupCreateProto(req)
	group.Id = utilities.GenerateObjectID()
	group.RootAdmin = false
	group, err := u.groupDB.GroupCreate(group)
	if err != nil {
		u.log.Errorf("groupDB.GroupCreate: %v", err)
		return nil, utilities.ErrorResponse(err, err.Error())
	}
	return &groupsService.CreateRes{Group: group.ToProto()}, nil
}

// Update a Group
func (u *GroupService) Update(ctx context.Context, req *groupsService.UpdateReq) (*groupsService.UpdateRes, error) {
	if !utilities.CheckObjectID(req.GetId()) {
		err := errors.New(req.GetId() + " is an invalid groupId")
		u.log.Errorf("utilities.CheckObjectID: %v", err)
		return nil, utilities.ErrorResponse(err, err.Error())
	}
	group := models.LoadGroupUpdateProto(req)
	groupId, err := models.VerifyGroupRequestScope(ctx, group.Id)
	if err != nil {
		u.log.Errorf("models.VerifyGroupRequestScope: %v", err)
		return nil, utilities.ErrorResponse(err, err.Error())
	}
	group.Id = groupId
	group, err = u.groupDB.GroupUpdate(group)
	if err != nil {
		u.log.Errorf("groupDB.GroupUpdate: %v", err)
		return nil, utilities.ErrorResponse(err, err.Error())
	}
	return &groupsService.UpdateRes{Group: group.ToProto()}, nil
}

// Get a specific Group
func (u *GroupService) Get(ctx context.Context, req *groupsService.GetReq) (*groupsService.GetRes, error) {
	if !utilities.CheckObjectID(req.GetId()) {
		err := errors.New(req.GetId() + " is an invalid groupId")
		u.log.Errorf("utilities.CheckObjectID: %v", err)
		return nil, utilities.ErrorResponse(err, err.Error())
	}
	groupId, err := models.VerifyGroupRequestScope(ctx, req.GetId())
	if err != nil {
		u.log.Errorf("models.VerifyGroupRequestScope: %v", err)
		return nil, utilities.ErrorResponse(err, err.Error())
	}
	group, err := u.groupDB.GroupFind(&models.Group{Id: groupId})
	if err != nil {
		u.log.Errorf("groupDB.GroupFind: %v", err)
		return nil, utilities.ErrorResponse(err, err.Error())
	}
	return &groupsService.GetRes{Group: group.ToProto()}, nil
}

// Find Groups from an input query
func (u *GroupService) Find(ctx context.Context, req *groupsService.FindReq) (*groupsService.FindRes, error) {
	// TODO NEXT FIX - valid req.GetQuery() authenticity / scope
	groups, err := u.groupDB.GroupsQuery(ctx, models.LoadGroupFindProto(req), utilities.NewPaginationQuery(int(req.GetSize()), int(req.GetPage())))
	if err != nil {
		u.log.Errorf("groupDB.GroupsQuery: %v", err)
		return nil, utilities.ErrorResponse(err, err.Error())
	}
	return &groupsService.FindRes{
		TotalCount: groups.TotalCount,
		TotalPages: groups.TotalPages,
		Page:       groups.Page,
		Size:       groups.Size,
		HasMore:    groups.HasMore,
		Groups:     groups.ToProto(),
	}, nil
}

// Delete is the handler function that deletes a group
func (u *GroupService) Delete(ctx context.Context, req *groupsService.DeleteReq) (*groupsService.DeleteRes, error) {
	if !utilities.CheckObjectID(req.GetId()) {
		err := errors.New("invalid group id")
		u.log.Errorf("utilities.CheckObjectID: %v", err)
		return nil, utilities.ErrorResponse(err, err.Error())
	}
	groupUsers, err := u.getGroupUsers(req.GetId())
	if err != nil {
		u.log.Errorf("GroupService.getGroupUsers: %v", err)
		return nil, utilities.ErrorResponse(err, err.Error())
	}
	err = u.deleteGroupAssets(groupUsers.Group, groupUsers.Users)
	if err != nil {
		u.log.Errorf("GroupService.deleteGroupAssets: %v", err)
		return nil, utilities.ErrorResponse(err, err.Error())
	}
	group, err := u.groupDB.GroupDelete(&models.Group{Id: req.GetId()})
	if err != nil {
		u.log.Errorf("groupDB.GroupDelete: %v", err)
		return nil, utilities.ErrorResponse(err, err.Error())
	}
	return &groupsService.DeleteRes{Group: group.ToProto()}, nil
}

// deleteGroupAssets asynchronously gets a group and its users from the database
func (u *GroupService) deleteGroupAssets(group *models.Group, users []*models.User) error {
	if !group.CheckID("id") {
		return errors.New("filter id cannot be empty for mass delete")
	}
	fErrCh := make(chan error) // Images Files Bulk Delete
	uErrCh := make(chan error) // Delete Group Users
	tErrCh := make(chan error) // Delete Group Tasks
	go func() {
		err := u.fileDB.FileDeleteMany(models.UsersToFiles(users))
		fErrCh <- err
	}()
	go func() {
		_, err := u.userDB.UserDeleteMany(&models.User{GroupId: group.Id})
		uErrCh <- err
	}()
	go func() {
		_, err := u.taskDB.TaskDeleteMany(&models.Task{GroupId: group.Id})
		tErrCh <- err
	}()
	for i := 0; i < 3; i++ {
		select {
		case fErr := <-fErrCh:
			if fErr != nil {
				return fErr
			}
		case uErr := <-uErrCh:
			if uErr != nil {
				return uErr
			}
		case tErr := <-tErrCh:
			if tErr != nil {
				return tErr
			}
		}
	}
	return nil
}

// getGroupUsers asynchronously gets a group and its users from the database
func (u *GroupService) getGroupUsers(groupId string) (*models.GroupUsers, error) {
	m := &models.GroupUsers{Users: []*models.User{}}
	gOutCh := make(chan *models.Group)
	gErrCh := make(chan error)
	uOutCh := make(chan []*models.User)
	uErrCh := make(chan error)
	go func() {
		reG, err := u.groupDB.GroupFind(&models.Group{Id: groupId})
		gOutCh <- reG
		gErrCh <- err
	}()
	go func() {
		reU, err := u.userDB.UsersFind(&models.User{GroupId: groupId})
		uOutCh <- reU
		uErrCh <- err
	}()
	for i := 0; i < 4; i++ {
		select {
		case gOut := <-gOutCh:
			m.Group = gOut
		case gErr := <-gErrCh:
			if gErr != nil {
				return m, gErr
			}
		case uOut := <-uOutCh:
			m.Users = uOut
		case uErr := <-uErrCh:
			if uErr != nil {
				return m, uErr
			}
		}
	}
	return m, nil
}

// getGroupTasks asynchronously gets a Group and its Tasks from the database
func (u *GroupService) getGroupTasks(groupId string) (*models.GroupTasks, error) {
	var m *models.GroupTasks
	gOutCh := make(chan *models.Group)
	gErrCh := make(chan error)
	uOutCh := make(chan []*models.Task)
	uErrCh := make(chan error)
	go func() {
		reG, err := u.groupDB.GroupFind(&models.Group{Id: groupId})
		gOutCh <- reG
		gErrCh <- err
	}()
	go func() {
		reU, err := u.taskDB.TasksFind(&models.Task{GroupId: groupId})
		uOutCh <- reU
		uErrCh <- err
	}()
	for i := 0; i < 4; i++ {
		select {
		case gOut := <-gOutCh:
			m.Group = gOut
		case gErr := <-gErrCh:
			if gErr != nil {
				return m, gErr
			}
		case uOut := <-uOutCh:
			m.Tasks = uOut
		case uErr := <-uErrCh:
			if uErr != nil {
				return m, uErr
			}
		}
	}
	return m, nil
}
