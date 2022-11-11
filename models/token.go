package models

import (
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"os"
)

// TokenData stores the structured data from a session token for use
type TokenData struct {
	UserId    string
	Role      string
	RootAdmin bool
	GroupId   string
}

// InitUserToken inputs a pointer to a user and returns TokenData
func InitUserToken(u *User) (*TokenData, error) {
	err := u.Validate("auth")
	if err != nil {
		return &TokenData{}, err
	}
	return &TokenData{
		UserId:    u.Id,
		Role:      u.Role,
		RootAdmin: u.RootAdmin,
		GroupId:   u.GroupId,
	}, nil
}

// ToUser creates a new User struct using the TokenData and returns a pointer to it
func (t *TokenData) ToUser() *User {
	return &User{
		Id:        t.UserId,
		Role:      t.Role,
		RootAdmin: t.RootAdmin,
		GroupId:   t.GroupId,
	}
}

// GetGroupsScope returns a scoped Group ID filter based on token User role
func (t *TokenData) GetGroupsScope() *Group {
	g := Group{Id: t.GroupId}
	if t.RootAdmin {
		g.Id = ""
	}
	return &g
}

// GetUsersScope returns a scoped User ID filter based on token User role
func (t *TokenData) GetUsersScope(scopeType string) *User {
	g := User{Id: t.UserId, GroupId: t.GroupId, RootAdmin: t.RootAdmin, Role: t.Role}
	if t.RootAdmin {
		g.Id = ""
		g.GroupId = ""
	} else if t.Role == "admin" && scopeType == "create" || scopeType == "update" {
		g.Id = ""
		g.GroupId = t.GroupId
	} else if scopeType == "find" {
		g.Id = ""
		g.GroupId = t.GroupId
	}
	return &g
}

// AdminRouteRoleCheck checks admin routes JWT tokens to ensure that a group admin does not break scope
func (t *TokenData) AdminRouteRoleCheck() string {
	groupId := ""
	if t.RootAdmin {
		groupId = t.GroupId
	}
	return groupId
}

// CreateToken is used to create a new session JWT token
func (t *TokenData) CreateToken(exp int64) (string, error) {
	if t.UserId == "" || t.GroupId == "" || t.Role == "" {
		return "", errors.New("missing required token claims")
	}
	if exp == 0 {
		return "", errors.New("new token must have a expiration time greater than 0")
	}
	var MySigningKey = []byte(os.Getenv("TOKEN_SECRET"))
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["id"] = t.UserId
	claims["role"] = t.Role
	claims["root"] = t.RootAdmin
	claims["group_id"] = t.GroupId
	claims["exp"] = exp
	return token.SignedString(MySigningKey)
}

// DecodeJWT is used to decode a JWT token
func DecodeJWT(curToken string) (*TokenData, error) {
	var tokenData TokenData
	if curToken == "" {
		return &tokenData, errors.New("unauthorized")
	}
	var MySigningKey = []byte(os.Getenv("TOKEN_SECRET"))
	// Decode token
	token, err := jwt.Parse(curToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("error")
		}
		return []byte(MySigningKey), nil
	})
	if err != nil {
		return &tokenData, err
	}
	// Determine user based on token
	if token.Valid {
		tokenClaims := token.Claims.(jwt.MapClaims)
		tokenData.UserId = tokenClaims["id"].(string)
		tokenData.Role = tokenClaims["role"].(string)
		tokenData.RootAdmin = tokenClaims["root"].(bool)
		tokenData.GroupId = tokenClaims["group_id"].(string)
		return &tokenData, nil
	}
	return &tokenData, errors.New("invalid token")
}

// LoadTokenFromRequest inputs a http request and returns decrypted TokenData or an error
func LoadTokenFromRequest(authToken string) (*TokenData, error) {
	tokenData, err := DecodeJWT(authToken)
	if err != nil {
		return nil, err
	}
	return tokenData, nil
}

// VerifyGroupRequestScope inputs a Group http request and returns decrypted TokenData or an error
func VerifyGroupRequestScope(authToken string, groupId string) (string, error) {
	tokenData, err := LoadTokenFromRequest(authToken)
	if err != nil {
		return "", err
	}
	if tokenData.RootAdmin || tokenData.GroupId == groupId {
		return groupId, nil
	}
	return "", errors.New("unauthorized")
}

// VerifyUserRequestScope inputs User http request and returns decrypted TokenData or an error
func VerifyUserRequestScope(authToken string, userId string, scopeType string) (*User, error) {
	tokenData, err := LoadTokenFromRequest(authToken)
	if err != nil {
		return nil, err
	}
	userScope := tokenData.GetUsersScope(scopeType)
	if tokenData.RootAdmin || tokenData.Role == "admin" { // default scope ok if user is a root admin or group admin
		userScope.Id = userId
		return userScope, nil
	}
	if scopeType == "update" && userScope.Id == userId { // default also ok if user is updating itself
		return userScope, nil
	}
	if scopeType == "find" && userScope.GroupId == tokenData.GroupId { // default also ok if user is finding in group
		return userScope, nil
	}
	return nil, errors.New("unauthorized")
}

// VerifyRequestScope inputs generic http requests and returns decrypted TokenData or an error
func VerifyRequestScope(authToken string, scopeType string) (*User, error) {
	tokenData, err := LoadTokenFromRequest(authToken)
	if err != nil {
		return nil, err
	}
	return tokenData.GetUsersScope(scopeType), nil
}
