package services

import (
	"context"
	"errors"
	"github.com/JECSand/go-grpc-server-boilerplate/models"
	authService "github.com/JECSand/go-grpc-server-boilerplate/protos/auth"
	"github.com/JECSand/go-grpc-server-boilerplate/utilities"
	"os"
)

// AuthService gRPC Service
type AuthService struct {
	log          utilities.Logger
	tokenService *TokenService
	userDB       UserDataService
	groupDB      GroupDataService
}

// NewAuthService constructs a UserService for controller gRPC service User requests
func NewAuthService(log utilities.Logger, ts *TokenService, u UserDataService, g GroupDataService) *AuthService {
	return &AuthService{
		log:          log,
		tokenService: ts,
		userDB:       u,
		groupDB:      g,
	}
}

// Register handler function that registers a new user
func (u *AuthService) Register(ctx context.Context, req *authService.RegisterReq) (*authService.RegisterRes, error) {
	if os.Getenv("REGISTRATION") == "OFF" {
		err := errors.New("not found")
		u.log.Errorf("AuthService.Register: %v", err)
		return nil, utilities.ErrorResponse(err, err.Error())
	}
	user := models.LoadRegisterProto(req)
	group := &models.Group{
		Id:        utilities.GenerateObjectID(),
		Name:      user.Email + "_group",
		RootAdmin: false,
	}
	g, err := u.groupDB.GroupCreate(group)
	if err != nil {
		u.log.Errorf("groupDB.GroupCreate: %v", err)
		return nil, utilities.ErrorResponse(err, err.Error())
	}
	user.Role = "admin"
	user.GroupId = g.Id
	user, err = u.userDB.UserCreate(user)
	if err != nil {
		u.log.Errorf("userDB.UserCreate: %v", err)
		return nil, utilities.ErrorResponse(err, err.Error())
	}
	newToken, err := u.tokenService.GenerateToken(user, "session")
	if err != nil {
		u.log.Errorf("tokenService.GenerateToken: %v", err)
		return nil, utilities.ErrorResponse(err, err.Error())
	}
	user.Password = ""
	return &authService.RegisterRes{User: user.ToAuthProto(), AccessToken: newToken}, nil
}

// Login is the handler function that manages the user SignIn process
func (u *AuthService) Login(ctx context.Context, req *authService.LoginReq) (*authService.LoginRes, error) {
	user := models.LoadLoginProto(req)
	err := user.Validate("login")
	if err != nil {
		u.log.Errorf("AuthService.Login: %v", err)
		return nil, utilities.ErrorResponse(err, err.Error())
	}
	user, err = u.userDB.AuthenticateUser(user)
	if err != nil {
		u.log.Errorf("userDB.AuthenticateUser: %v", err)
		return nil, utilities.ErrorResponse(err, err.Error())
	}
	sessionToken, err := u.tokenService.GenerateToken(user, "session")
	if err != nil {
		u.log.Errorf("tokenService.GenerateToken: %v", err)
		return nil, utilities.ErrorResponse(err, err.Error())
	}
	user.Password = ""
	return &authService.LoginRes{User: user.ToAuthProto(), AccessToken: sessionToken}, nil
}

// Logout is the handler function that ends a users session
func (u *AuthService) Logout(ctx context.Context, req *authService.Empty) (*authService.LogoutRes, error) {
	accessToken, err := utilities.GetTokenFromContext(ctx)
	if err != nil {
		u.log.Errorf("utilities.GetTokenFromContext: %v", err)
		return nil, utilities.ErrorResponse(err, err.Error())
	}
	err = u.tokenService.BlacklistAuthToken(accessToken)
	if err != nil {
		u.log.Errorf("tokenService.BlacklistAuthToken: %v", err)
		return nil, utilities.ErrorResponse(err, err.Error())
	}
	return &authService.LogoutRes{Status: 200}, nil
}

// Refresh is the handler function that refreshes a users JWT token
func (u *AuthService) Refresh(ctx context.Context, req *authService.Empty) (*authService.RefreshRes, error) {
	tokenClaims, err := models.LoadTokenFromContext(ctx)
	if err != nil {
		u.log.Errorf("models.LoadTokenFromContext: %v", err)
		return nil, utilities.ErrorResponse(err, err.Error())
	}
	user, err := u.userDB.UserFind(tokenClaims.ToUser())
	if err != nil {
		u.log.Errorf("userDB.UserFind: %v", err)
		return nil, utilities.ErrorResponse(err, err.Error())
	}
	sessionToken, err := u.tokenService.GenerateToken(user, "session")
	if err != nil {
		u.log.Errorf("tokenService.GenerateToken: %v", err)
		return nil, utilities.ErrorResponse(err, err.Error())
	}
	return &authService.RefreshRes{AccessToken: sessionToken}, nil
}

// GenerateKey is the handler function that generates 6 month API Key for a given user
func (u *AuthService) GenerateKey(ctx context.Context, req *authService.Empty) (*authService.GenerateKeyRes, error) {
	tokenClaims, err := models.LoadTokenFromContext(ctx)
	if err != nil {
		u.log.Errorf("models.LoadTokenFromContext: %v", err)
		return nil, utilities.ErrorResponse(err, err.Error())
	}
	user, err := u.userDB.UserFind(tokenClaims.ToUser())
	if err != nil {
		u.log.Errorf("userDB.UserFind: %v", err)
		return nil, utilities.ErrorResponse(err, err.Error())
	}
	apiKey, err := u.tokenService.GenerateToken(user, "api")
	if err != nil {
		u.log.Errorf("tokenService.GenerateToken: %v", err)
		return nil, utilities.ErrorResponse(err, err.Error())
	}
	return &authService.GenerateKeyRes{APIKey: apiKey}, nil
}

// UpdatePassword is the handler function that manages the user password update process
func (u *AuthService) UpdatePassword(ctx context.Context, req *authService.UpdatePasswordReq) (*authService.UpdatePasswordRes, error) {
	tokenClaims, err := models.LoadTokenFromContext(ctx)
	if err != nil {
		u.log.Errorf("models.LoadTokenFromContext: %v", err)
		return nil, utilities.ErrorResponse(err, err.Error())
	}
	pw := models.LoadPasswordUpdateProto(req)
	user := tokenClaims.ToUser()
	_, err = u.userDB.UpdatePassword(user, pw.CurrentPassword, pw.NewPassword)
	if err != nil {
		u.log.Errorf("userDB.UpdatePassword: %v", err)
		return nil, utilities.ErrorResponse(err, err.Error())
	}
	return &authService.UpdatePasswordRes{Status: 200}, nil
}
