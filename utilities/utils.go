package utilities

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// JsonErr structures a standard error to return
type JsonErr struct {
	Code int    `json:"code"`
	Text string `json:"text"`
}

// JWTError is a struct that is used to contain a json encoded error message for any JWT related errors
type JWTError struct {
	Message string `json:"message"`
}

// GenerateObjectID for index keying records of data
func GenerateObjectID() string {
	newId := primitive.NewObjectID()
	return newId.Hex()
}

// CheckObjectID checks whether a hexID is null or now
func CheckObjectID(hexID string) bool {
	if hexID == "" || hexID == "000000000000000000000000" {
		return false
	}
	return true
}

// StrToBool converts a string to a bool
func StrToBool(in string) bool {
	if in == "true" || in == "TRUE" {
		return true
	}
	return false
}

// BoolToStr converts a bool to a string value
func BoolToStr(in bool) string {
	if in {
		return "true"
	}
	return "false"
}
