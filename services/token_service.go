package services

import (
	"errors"
	"github.com/JECSand/go-grpc-server-boilerplate/auth"
	"github.com/JECSand/go-grpc-server-boilerplate/models"
	"time"
)

// TokenService is used by the app to manage db auth functionality
type TokenService struct {
	uService UserDataService
	gService GroupDataService
	bService BlacklistDataService
}

// NewTokenService is an exported function used to initialize a new authService struct
func NewTokenService(uService UserDataService, gService GroupDataService, bService BlacklistDataService) *TokenService {
	return &TokenService{uService, gService, bService}
}

// verifyTokenUser verifies Token's User
func (a *TokenService) verifyTokenUser(decodedToken *auth.TokenData) (bool, string) {
	tUser := decodedToken.ToUser()
	checkUser, err := a.uService.UserFind(tUser)
	if err != nil {
		return false, err.Error()
	}
	checkGroup, err := a.gService.GroupFind(&models.Group{Id: tUser.GroupId})
	if err != nil {
		return false, err.Error()
	}
	// validate the Group id of the User and the associated User's Group
	if checkUser.GroupId != checkGroup.Id {
		return false, "Incorrect group id"
	}
	return true, "No Error"
}

// tokenVerifyMiddleWare inputs the route handler function along with User roleType to verify User token and permissions
func (a *TokenService) tokenVerifyMiddleWare(roleType string, authToken string) (bool, error) {
	if a.bService.CheckTokenBlacklist(authToken) {
		return false, errors.New("invalid token")
	}
	decodedToken, err := auth.DecodeJWT(authToken)
	if err != nil {
		return false, err
	}
	verified, verifyMsg := a.verifyTokenUser(decodedToken)
	if verified {
		if roleType == "Root" && decodedToken.RootAdmin {
			return false, nil
		} else if roleType == "Admin" && decodedToken.Role == "admin" {
			return false, nil
		} else if roleType == "Member" {
			return false, nil
		} else {
			return false, errors.New("invalid token")
		}
	} else {
		return false, errors.New(verifyMsg)
	}
}

// GenerateToken outputs an auth token string for an inputted User
func (a *TokenService) GenerateToken(u *models.User, tType string) (string, error) {
	expDT := time.Now().Add(time.Hour * 1).Unix() // Default 1 hour expiration for session token
	if tType == "api" {
		expDT = time.Now().Add(time.Hour * 4380).Unix() // 6 month expiration for api key
	}
	tData, err := auth.InitUserToken(u)
	if err != nil {
		return "", err
	}
	return tData.CreateToken(expDT)
}

// RootAdminTokenVerifyMiddleWare is used to verify that the requester is a valid admin
func (a *TokenService) RootAdminTokenVerifyMiddleWare(authToken string) (bool, error) {
	return a.tokenVerifyMiddleWare("Root", authToken)
}

// AdminTokenVerifyMiddleWare is used to verify that the requester is a valid admin
func (a *TokenService) AdminTokenVerifyMiddleWare(authToken string) (bool, error) {
	return a.tokenVerifyMiddleWare("Admin", authToken)
}

// MemberTokenVerifyMiddleWare is used to verify that a requester is authenticated
func (a *TokenService) MemberTokenVerifyMiddleWare(authToken string) (bool, error) {
	return a.tokenVerifyMiddleWare("Member", authToken)
}

// BlacklistAuthToken is used to blacklist an unexpired token
func (a *TokenService) BlacklistAuthToken(authToken string) error {
	return a.bService.BlacklistAuthToken(authToken)
}
