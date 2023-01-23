package services

import (
	"context"
	"errors"
	"github.com/JECSand/go-grpc-server-boilerplate/models"
	usersService "github.com/JECSand/go-grpc-server-boilerplate/protos/user"
	"github.com/JECSand/go-grpc-server-boilerplate/utilities"
)

// UserService gRPC Service
type UserService struct {
	log          utilities.Logger
	tokenService *TokenService
	userDB       UserDataService
	groupDB      GroupDataService
	taskDB       TaskDataService
	fileDB       FileDataService
}

// NewUserService constructs a UserService for controller gRPC service User requests
func NewUserService(log utilities.Logger, ts *TokenService, u UserDataService, g GroupDataService, t TaskDataService, f FileDataService) *UserService {
	return &UserService{
		log:          log,
		tokenService: ts,
		userDB:       u,
		groupDB:      g,
		taskDB:       t,
		fileDB:       f,
	}
}

// Create is a New User
func (u *UserService) Create(ctx context.Context, req *usersService.CreateReq) (*usersService.CreateRes, error) {
	user := models.LoadUserCreateProto(req)
	err := user.Validate("create")
	if err != nil {
		u.log.Errorf("user.Validate: %v", err)
		return nil, utilities.ErrorResponse(err, err.Error())
	}
	accessToken, err := utilities.GetTokenFromContext(ctx)
	if err != nil {
		u.log.Errorf("utilities.GetTokenFromContext: %v", err)
		return nil, utilities.ErrorResponse(err, err.Error())
	}
	decodedToken, err := models.DecodeJWT(accessToken)
	if err != nil {
		u.log.Errorf("utilities.GetTokenFromContext: %v", err)
		return nil, utilities.ErrorResponse(err, err.Error())
	}
	userScope := decodedToken.GetUsersScope("create")
	user.LoadScope(userScope, "create")
	if user.GroupId == "" {
		user.GroupId = decodedToken.GroupId
	}
	user, err = u.userDB.UserCreate(user)
	if err != nil {
		u.log.Errorf("userDB.UserCreate: %v", err)
		return nil, utilities.ErrorResponse(err, err.Error())
	}
	user.Password = ""
	return &usersService.CreateRes{User: user.ToProto()}, nil
}

// Update a User
func (u *UserService) Update(ctx context.Context, req *usersService.UpdateReq) (*usersService.UpdateRes, error) {
	if !utilities.CheckObjectID(req.GetId()) {
		err := errors.New(req.GetId() + " is an invalid userId")
		u.log.Errorf("utilities.CheckObjectID: %v", err)
		return nil, utilities.ErrorResponse(err, err.Error())
	}
	user := models.LoadUserUpdateProto(req)
	userScope, err := models.VerifyUserRequestScope(ctx, user.Id, "update")
	if err != nil {
		u.log.Errorf("models.VerifyUserRequestScope: %v", err)
		return nil, utilities.ErrorResponse(err, err.Error())
	}
	user.LoadScope(userScope, "update")
	user, err = u.userDB.UserUpdate(user)
	if err != nil {
		u.log.Errorf("userDB.UserUpdate: %v", err)
		return nil, utilities.ErrorResponse(err, err.Error())
	}
	return &usersService.UpdateRes{User: user.ToProto()}, nil
}

// Get a specific User
func (u *UserService) Get(ctx context.Context, req *usersService.GetReq) (*usersService.GetRes, error) {
	if !utilities.CheckObjectID(req.GetId()) {
		err := errors.New(req.GetId() + " is an invalid userId")
		u.log.Errorf("utilities.CheckObjectID: %v", err)
		return nil, utilities.ErrorResponse(err, err.Error())
	}
	filter := models.User{Id: req.GetId()}
	userScope, err := models.VerifyUserRequestScope(ctx, req.GetId(), "find")
	if err != nil {
		u.log.Errorf("models.VerifyUserRequestScope: %v", err)
		return nil, utilities.ErrorResponse(err, err.Error())
	}
	filter.LoadScope(userScope, "find")
	user, err := u.userDB.UserFind(&filter)
	if err != nil {
		u.log.Errorf("userDB.UserFind: %v", err)
		return nil, utilities.ErrorResponse(err, err.Error())
	}
	user.Password = ""
	return &usersService.GetRes{User: user.ToProto()}, nil
}

// Find Users from an input query
func (u *UserService) Find(ctx context.Context, req *usersService.FindReq) (*usersService.FindRes, error) {
	// TODO NEXT FIX - valid req.GetQuery() authenticity / scope
	users, err := u.userDB.UsersQuery(ctx, models.LoadUserFindProto(req), utilities.NewPaginationQuery(int(req.GetSize()), int(req.GetPage())))
	if err != nil {
		u.log.Errorf("userDB.UsersQuery: %v", err)
		return nil, utilities.ErrorResponse(err, err.Error())
	}
	return &usersService.FindRes{
		TotalCount: users.TotalCount,
		TotalPages: users.TotalPages,
		Page:       users.Page,
		Size:       users.Size,
		HasMore:    users.HasMore,
		Users:      users.ToProto(),
	}, nil
}

// GetGroupUsers returns the users for a given groupId
func (u *UserService) GetGroupUsers(ctx context.Context, req *usersService.GetGroupUsersReq) (*usersService.GetGroupUsersRes, error) {
	if !utilities.CheckObjectID(req.GetGroupId()) {
		err := errors.New("invalid group id")
		u.log.Errorf("utilities.CheckObjectID: %v", err)
		return nil, utilities.ErrorResponse(err, err.Error())
	}
	groupId, err := models.VerifyGroupRequestScope(ctx, req.GetGroupId())
	if err != nil {
		u.log.Errorf("models.VerifyGroupRequestScope: %v", err)
		return nil, utilities.ErrorResponse(err, err.Error())
	}
	users, err := u.userDB.UsersQuery(ctx, &models.User{GroupId: groupId}, utilities.NewPaginationQuery(int(req.GetSize()), int(req.GetPage())))
	if err != nil {
		u.log.Errorf("userDB.UsersQuery: %v", err)
		return nil, utilities.ErrorResponse(err, err.Error())
	}
	return &usersService.GetGroupUsersRes{
		TotalCount: users.TotalCount,
		TotalPages: users.TotalPages,
		Page:       users.Page,
		Size:       users.Size,
		HasMore:    users.HasMore,
		Users:      users.ToProto(),
	}, nil
}

// Delete is the handler function that deletes a user
func (u *UserService) Delete(ctx context.Context, req *usersService.DeleteReq) (*usersService.DeleteRes, error) {
	if !utilities.CheckObjectID(req.GetId()) {
		err := errors.New(req.GetId() + " is an invalid userId")
		u.log.Errorf("utilities.CheckObjectID: %v", err)
		return nil, utilities.ErrorResponse(err, err.Error())
	}
	filter := models.User{Id: req.GetId()}
	userScope, err := models.VerifyUserRequestScope(ctx, req.GetId(), "find")
	if err != nil {
		u.log.Errorf("models.VerifyUserRequestScope: %v", err)
		return nil, utilities.ErrorResponse(err, err.Error())
	}
	filter.LoadScope(userScope, "find")
	user, err := u.userDB.UserFind(&filter)
	if err != nil {
		u.log.Errorf("userDB.UserFind: %v", err)
		return nil, utilities.ErrorResponse(err, err.Error())
	}
	err = u.deleteUserAssets(user)
	if err != nil {
		u.log.Errorf("userDB.deleteUserAssets: %v", err)
		return nil, utilities.ErrorResponse(err, err.Error())
	}
	user, err = u.userDB.UserDelete(&filter)
	if err != nil {
		u.log.Errorf("userDB.UserDelete: %v", err)
		return nil, utilities.ErrorResponse(err, err.Error())
	}
	user.Password = ""
	return &usersService.DeleteRes{User: user.ToProto()}, nil
}

// deleteUserAssets asynchronously gets a group and its users from the database
func (u *UserService) deleteUserAssets(user *models.User) error {
	if !user.CheckID("id") {
		return errors.New("filter id cannot be empty for mass delete")
	}
	ctx, cancel := context.WithCancel(context.Background())
	errChan := make(chan error)
	defer func() {
		cancel()
		close(errChan)
	}()
	go func() {
		if user.CheckID("image_id") {
			_, err := u.fileDB.FileDelete(&models.File{OwnerId: user.Id, OwnerType: "user"})
			select {
			case <-ctx.Done():
				return
			default:
			}
			errChan <- err
		} else {
			errChan <- nil
		}
	}()
	go func() {
		_, err := u.taskDB.TaskDeleteMany(&models.Task{UserId: user.Id})
		select {
		case <-ctx.Done():
			return
		default:
		}
		errChan <- err
	}()
	var err error
	for i := 0; i < 2; i++ {
		select {
		case err = <-errChan:
			if err != nil {
				break
			}
		}
	}
	return err
}

/*
// UploadImage allows for a user image to be associated with the User record
func (u *UserService) UploadImage(stream *usersService.UploadImageReq) error {
	name := "fixLater.txt"
	file := models.NewInFile(name)
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			if err := s.storage.Store(file); err != nil {
				return status.Error(codes.Internal, err.Error())
			}
			return stream.SendAndClose(&usersService.UploadImageRes{Name: name})
		}
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}

		if err := file.Write(req.GetChunk()); err != nil {
			return status.Error(codes.Internal, err.Error())
		}
	}

		filter := models.User{Id: "00000000000000"}
		userScope, err := auth.VerifyUserRequestScope(r, userId, "update")
		if err != nil {
			utilities.RespondWithError(w, http.StatusUnauthorized, utilities.JWTError{Message: err.Error()})
			return
		}
		filter.LoadScope(userScope, "find")
		user, err := ur.uService.UserFind(&filter)
		if err != nil {
			utilities.RespondWithError(w, http.StatusNotFound, utilities.JWTError{Message: "user not found"})
			return
		}
		newImage := false
		if !user.CheckID("image_id") {
			user.ImageId = utilities.GenerateObjectID()
			newImage = true
		}
		f := &models.File{Id: user.ImageId, OwnerType: "user", OwnerId: user.Id, BucketType: "user-images", Name: handler.Filename}
		buf := bytes.NewBuffer(nil)
		if _, err = io.Copy(buf, file); err != nil {
			utilities.RespondWithError(w, http.StatusInternalServerError, utilities.JWTError{Message: err.Error()})
			return
		}
		if newImage {
			f, err = ur.fService.FileCreate(f, buf.Bytes())
			if err != nil {
				utilities.RespondWithError(w, http.StatusInternalServerError, utilities.JWTError{Message: err.Error()})
				return
			}
			user, err = ur.uService.UserUpdate(&models.User{Id: user.Id, ImageId: user.ImageId})
			if err != nil {
				utilities.RespondWithError(w, http.StatusInternalServerError, utilities.JWTError{Message: err.Error()})
				return
			}
		} else {
			f, err = ur.fService.FileUpdate(f, buf.Bytes())
			if err != nil {
				utilities.RespondWithError(w, http.StatusInternalServerError, utilities.JWTError{Message: err.Error()})
				return
			}
		}
		user.Password = ""
		w = utilities.SetResponseHeaders(w, "", "")
		w.WriteHeader(http.StatusOK)
		if err = json.NewEncoder(w).Encode(user); err != nil {
			return
		}
		return
}
*/
