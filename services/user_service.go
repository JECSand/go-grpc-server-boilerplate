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
	//span, ctx := opentracing.StartSpanFromContext(ctx, "userService.Create")
	//defer span.Finish()
	user := models.LoadUserCreateProto(req)
	// TODO NEXT FIX
	/*
		decodedToken, err := auth.DecodeJWT(r.Header.Get("Auth-Token"))
		if err != nil {
			utilities.RespondWithError(w, http.StatusUnauthorized, utilities.JWTError{Message: err.Error()})
			return
		}
		userScope := decodedToken.GetUsersScope("create")
		user.LoadScope(userScope, "create")
		if user.GroupId == "" {
			user.GroupId = decodedToken.GroupId
		}
	*/
	user, err := u.userDB.UserCreate(user)
	if err != nil {
		u.log.Errorf("userDB.UserCreate: %v", err)
		return nil, utilities.ErrorResponse(err, err.Error())
	}
	user.Password = ""
	return &usersService.CreateRes{User: user.ToProto()}, nil
}

// Update a User
func (u *UserService) Update(ctx context.Context, req *usersService.UpdateReq) (*usersService.UpdateRes, error) {
	//span, ctx := opentracing.StartSpanFromContext(ctx, "userService.Update")
	//defer span.Finish()
	if !utilities.CheckObjectID(req.GetId()) {
		err := errors.New(req.GetId() + " is an invalid userId")
		u.log.Errorf("utilities.CheckObjectID: %v", err)
		return nil, utilities.ErrorResponse(err, err.Error())
	}
	user := models.LoadUserUpdateProto(req)
	// TODO NEXT FIX
	/*
		userScope, err := auth.VerifyUserRequestScope(r, user.Id, "update")
		if err != nil {
			u.log.Errorf("auth.VerifyUserRequestScope: %v", err)
			return nil, utilities.ErrorResponse(err, err.Error())
		}
		user.LoadScope(userScope, "update")
	*/
	user, err := u.userDB.UserUpdate(user)
	if err != nil {
		u.log.Errorf("userDB.UserUpdate: %v", err)
		return nil, utilities.ErrorResponse(err, err.Error())
	}
	return &usersService.UpdateRes{User: user.ToProto()}, nil
}

// Get a specific User
func (u *UserService) Get(ctx context.Context, req *usersService.GetReq) (*usersService.GetRes, error) {
	//span, ctx := opentracing.StartSpanFromContext(ctx, "userService.Get")
	//defer span.Finish()
	if !utilities.CheckObjectID(req.GetId()) {
		err := errors.New(req.GetId() + " is an invalid userId")
		u.log.Errorf("utilities.CheckObjectID: %v", err)
		return nil, utilities.ErrorResponse(err, err.Error())
	}
	filter := models.User{Id: req.GetId()}
	// TODO NEXT FIX
	/*
		userScope, err := auth.VerifyUserRequestScope(r, userId, "find")
		if err != nil {
			utilities.RespondWithError(w, http.StatusUnauthorized, utilities.JWTError{Message: err.Error()})
			return
		}
		filter.LoadScope(userScope, "find")
	*/
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
	//span, ctx := opentracing.StartSpanFromContext(ctx, "userService.Find")
	//defer span.Finish()
	// TODO NEXT FIX - valid req.GetQuery() authenticity / scope
	/*
		decodedToken, err := auth.DecodeJWT(r.Header.Get("Auth-Token"))
		if err != nil {
			utilities.RespondWithError(w, http.StatusUnauthorized, utilities.JWTError{Message: err.Error()})
			return
		}
		userScope := decodedToken.GetUsersScope("find")
		filter.LoadScope(userScope, "find")
	*/
	users, err := u.userDB.UsersQuery(ctx, req.GetQuery(), utilities.NewPaginationQuery(int(req.GetSize()), int(req.GetPage())))
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
	//span, ctx := opentracing.StartSpanFromContext(ctx, "userService.Find")
	//defer span.Finish()
	// TODO NEXT FIX - valid req.GetQuery() authenticity / scope
	//		CHECK GroupID
	/*
		decodedToken, err := auth.DecodeJWT(r.Header.Get("Auth-Token"))
		if err != nil {
			utilities.RespondWithError(w, http.StatusUnauthorized, utilities.JWTError{Message: err.Error()})
			return
		}
		userScope := decodedToken.GetUsersScope("find")
		filter.LoadScope(userScope, "find")
	*/
	users, err := u.userDB.UsersQuery(ctx, req.GetGroupId(), utilities.NewPaginationQuery(int(req.GetSize()), int(req.GetPage())))
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
	//span, ctx := opentracing.StartSpanFromContext(ctx, "userService.Find")
	//defer span.Finish()
	if !utilities.CheckObjectID(req.GetId()) {
		err := errors.New(req.GetId() + " is an invalid userId")
		u.log.Errorf("utilities.CheckObjectID: %v", err)
		return nil, utilities.ErrorResponse(err, err.Error())
	}
	filter := models.User{Id: req.GetId()}
	// TODO NEXT FIX
	/*
		userScope, err := auth.VerifyUserRequestScope(r, userId, "update")
		if err != nil {
			utilities.RespondWithError(w, http.StatusUnauthorized, utilities.JWTError{Message: err.Error()})
			return
		}
		filter.LoadScope(userScope, "find")
	*/
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

// deleteUserAssets asynchronously gets a group and its users from the database
func (u *UserService) deleteUserAssets(user *models.User) error {
	if !user.CheckID("id") {
		return errors.New("filter id cannot be empty for mass delete")
	}
	gErrCh := make(chan error)
	uErrCh := make(chan error)
	go func() {
		if user.CheckID("image_id") {
			_, err := u.fileDB.FileDelete(&models.File{OwnerId: user.Id, OwnerType: "user"})
			gErrCh <- err
		} else {
			gErrCh <- nil
		}
	}()
	go func() {
		_, err := u.taskDB.TaskDeleteMany(&models.Task{UserId: user.Id})
		uErrCh <- err
	}()
	for i := 0; i < 2; i++ {
		select {
		case gErr := <-gErrCh:
			if gErr != nil {
				return gErr
			}
		case uErr := <-uErrCh:
			if uErr != nil {
				return uErr
			}
		}
	}
	return nil
}
