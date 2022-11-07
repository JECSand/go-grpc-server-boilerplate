package models

import (
	"bytes"
	"errors"
	"github.com/JECSand/go-grpc-server-boilerplate/utilities"
	"strings"
	"time"
)

// InFile is used for streaming bytes into
type InFile struct {
	name   string
	buffer *bytes.Buffer
}

// NewInFile constructs an InFile
func NewInFile(name string) *InFile {
	return &InFile{
		name:   name,
		buffer: &bytes.Buffer{},
	}
}

// Write bytes unto an InFile
func (f *InFile) Write(chunk []byte) error {
	_, err := f.buffer.Write(chunk)
	return err
}

// File is a root struct that is used to store the json encoded data for/from a mongodb file doc.
type File struct {
	Id           string    `json:"id,omitempty"`
	OwnerId      string    `json:"owner_id,omitempty"`
	OwnerType    string    `json:"owner_type,omitempty"`
	GridFSId     string    `json:"gridfs_id,omitempty"`
	BucketName   string    `json:"bucket_name,omitempty"`
	BucketType   string    `json:"bucket_type,omitempty"`
	Name         string    `json:"name,omitempty"`
	FileType     string    `json:"file_type,omitempty"`
	Size         int       `json:"size,omitempty"`
	LastModified time.Time `json:"last_modified,omitempty"`
	CreatedAt    time.Time `json:"created_at,omitempty"`
	DeletedAt    time.Time `json:"deleted_at,omitempty"`
}

// BuildBucketName returns a current name for the bucket of a GridFS File
func (g *File) BuildBucketName() error {
	if g.CheckID("owner_id") && g.OwnerType != "" {
		g.BucketName = g.OwnerType + "_" + g.OwnerId + "_bucket"
		return nil
	}
	return errors.New("file missing owner_id")
}

// CheckID determines whether a specified ID is set or not
func (g *File) CheckID(chkId string) bool {
	switch chkId {
	case "id":
		if !utilities.CheckObjectID(g.Id) {
			return false
		}
	case "owner_id":
		if !utilities.CheckObjectID(g.OwnerId) {
			return false
		}
	case "gridfs_id":
		if !utilities.CheckObjectID(g.GridFSId) {
			return false
		}
	}
	return true
}

// Validate a File for different scenarios such as creating new File, or updating a File
func (g *File) Validate(valCase string) (err error) {
	var missingFields []string
	switch valCase {
	case "create":
		if !g.CheckID("owner_id") {
			missingFields = append(missingFields, "owner_id")
		}
		if g.OwnerType == "" {
			missingFields = append(missingFields, "owner_type")
		}
		if g.Name == "" {
			missingFields = append(missingFields, "name")
		}
		if g.FileType == "" {
			missingFields = append(missingFields, "file_type")
		}
	case "update":
		if !g.CheckID("id") {
			missingFields = append(missingFields, "id")
		}
	case "retrieve":
		if !g.CheckID("id") && !g.CheckID("owner_id") && !g.CheckID("gridfs_id") {
			missingFields = append(missingFields, "id")
		}
	default:
		return errors.New("unrecognized validation case")
	}
	if len(missingFields) > 0 {
		return errors.New("missing the following file fields: " + strings.Join(missingFields, ", "))
	}
	return
}

// BuildFilter is a function that setups the base user struct during a File modification request
func (g *File) BuildFilter() (*File, error) {
	var filter File
	if g.CheckID("id") {
		filter.Id = g.Id
	} else if g.CheckID("owner_id") {
		filter.OwnerId = g.OwnerId
	} else if g.CheckID("gridfs_id") {
		filter.GridFSId = g.GridFSId
	} else {
		return nil, errors.New("file is missing a valid query filter")
	}
	return &filter, nil
}

// BuildUpdate is a function that setups the base file struct during a file modification request
func (g *File) BuildUpdate(cur *File) {
	c := 0
	if len(g.OwnerId) == 0 {
		g.OwnerId = cur.OwnerId
		c++
	}
	if len(g.OwnerType) == 0 {
		g.OwnerType = cur.OwnerType
		c++
	}
	if len(g.BucketType) == 0 {
		g.BucketType = cur.BucketType
	}
	if len(g.Name) == 0 {
		g.Name = cur.Name
	}
	if len(g.FileType) == 0 {
		g.FileType = cur.FileType
	}
	if c == 0 {
		_ = g.BuildBucketName()
	} else {
		g.BucketName = cur.BucketName
	}
}
