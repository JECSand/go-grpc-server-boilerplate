package database

import (
	"bytes"
	"context"
	"errors"
	"github.com/JECSand/go-grpc-server-boilerplate/models"
	"github.com/JECSand/go-grpc-server-boilerplate/utilities"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"sync"
)

// FileService is used by the app to manage all File related controllers and functionality
type FileService struct {
	collection   DBCollection
	db           DBClient
	fileHandler  *DBHandler[*fileModel]
	userHandler  *DBHandler[*userModel]
	groupHandler *DBHandler[*groupModel]
}

// NewFileService is an exported function used to initialize a new FileService struct
func NewFileService(db DBClient, fHandler *DBHandler[*fileModel], uHandler *DBHandler[*userModel], gHandler *DBHandler[*groupModel]) *FileService {
	collection := db.GetCollection("files")
	return &FileService{
		collection,
		db,
		fHandler,
		uHandler,
		gHandler,
	}
}

// deleteBucket deletes an existing GridFS bucket
func (p *FileService) deleteBucket(bucketName string) error {
	bucket, err := p.db.GetBucket(bucketName)
	if err != nil {
		return err
	}
	return bucket.Drop()
}

// uploadFileToBucket uploads a file to a bucket
func (p *FileService) uploadFileToBucket(g *fileModel, fileContent []byte) (primitive.ObjectID, error) {
	bucket, err := p.db.GetBucket(g.BucketName)
	if err != nil {
		return primitive.NewObjectID(), err
	}
	fileId, err := bucket.UploadFromStream(g.Name, bytes.NewBuffer(fileContent))
	if err != nil {
		return fileId, err
	}
	return fileId, nil
}

// downloadFileFromBucket gets a file from a bucket
func (p *FileService) downloadFileFromBucket(g *fileModel) (*bytes.Buffer, error) {
	bucket, err := p.db.GetBucket(g.BucketName)
	w := bytes.NewBuffer(make([]byte, 0))
	_, err = bucket.DownloadToStream(g.GridFSId, w)
	if err != nil {
		return w, err
	}
	return w, nil
}

// deleteFileFromBucket deletes a file from a bucket
func (p *FileService) deleteFileFromBucket(g *fileModel) error {
	bucket, err := p.db.GetBucket(g.BucketName)
	if err != nil {
		return err
	}
	return bucket.Delete(g.GridFSId)
}

// checkFileOwner queries an OwnerId to verify the record is legit
func (p *FileService) checkFileOwner(g *fileModel) error {
	if g.OwnerType == "group" {
		gm, err := p.groupHandler.FindOne(&groupModel{Id: g.OwnerId})
		if err != nil {
			return err
		}
		if gm.toRoot().CheckID("id") {
			return nil
		}
	} else if g.OwnerType == "user" {
		gm, err := p.userHandler.FindOne(&userModel{Id: g.OwnerId})
		if err != nil {
			return err
		}
		if gm.toRoot().CheckID("id") {
			return nil
		}
	}
	return errors.New("invalid file owner")
}

// FilesFind is used to find many files
func (p *FileService) FilesFind(g *models.File) ([]*models.File, error) {
	var files []*models.File
	tm, err := newFileModel(g)
	if err != nil {
		return files, err
	}
	gms, err := p.fileHandler.FindMany(tm)
	if err != nil {
		return files, err
	}
	for _, gm := range gms {
		files = append(files, gm.toRoot())
	}
	return files, nil
}

// FileFind is used to find a specific file
func (p *FileService) FileFind(g *models.File) (*models.File, error) {
	gm, err := newFileModel(g)
	if err != nil {
		return nil, err
	}
	gm, err = p.fileHandler.FindOne(gm)
	if err != nil {
		return nil, err
	}
	return gm.toRoot(), nil
}

// FileCreate creates a new GridFS File
func (p *FileService) FileCreate(g *models.File, content []byte) (*models.File, error) {
	err := g.Validate("create")
	if err != nil {
		return nil, err
	}
	err = g.BuildBucketName()
	if err != nil {
		return nil, err
	}
	g.Size = len(content)
	gm, err := newFileModel(g)
	if err != nil {
		return nil, err
	}
	err = p.checkFileOwner(gm) // verify that the owner of the new file is a valid db record
	if err != nil {
		return nil, err
	}
	gridFSId, err := p.uploadFileToBucket(gm, content)
	if err != nil {
		return nil, err
	}
	gm.GridFSId = gridFSId
	gm, err = p.fileHandler.InsertOne(gm)
	if err != nil {
		err = p.deleteFileFromBucket(gm)
		if err != nil {
			panic("unable to delete orphaned file: " + gm.GridFSId.Hex() + " from GridFS bucket! msg: " + err.Error())
		}
		return nil, err
	}
	return gm.toRoot(), nil
}

// FileUpdate is used to update an existing File
func (p *FileService) FileUpdate(g *models.File, content []byte) (*models.File, error) {
	var filter models.File
	err := g.Validate("update")
	if err != nil {
		return nil, err
	}
	filter.Id = g.Id
	f, err := newFileModel(&filter)
	if err != nil {
		return nil, err
	}
	cur, err := p.fileHandler.FindOne(f)
	if err != nil {
		return nil, errors.New("file not found")
	}
	g.BuildUpdate(cur.toRoot())
	gm, err := newFileModel(g)
	if err != nil {
		return nil, err
	}
	if gm.BucketName != cur.BucketName { // if new file owner and type in update, then verify the new owner
		err = p.checkFileOwner(gm)
		if err != nil {
			return nil, err
		}
	}
	if len(content) > 0 && cur.Size != len(content) {
		err = p.deleteFileFromBucket(cur)
		if err != nil {
			return nil, err
		}
		gridFSId, err := p.uploadFileToBucket(gm, content)
		if err != nil {
			return nil, err
		}
		gm.GridFSId = gridFSId
		gm.Size = len(content)
	}
	gm, err = p.fileHandler.UpdateOne(f, gm)
	if err != nil {
		return nil, err
	}
	return gm.toRoot(), err
}

// FileDelete is used to delete a GridFS File
func (p *FileService) FileDelete(g *models.File) (*models.File, error) {
	gm, err := newFileModel(g)
	if err != nil {
		return nil, err
	}
	gm, err = p.fileHandler.DeleteOne(gm)
	if err != nil {
		return nil, err
	}
	err = p.deleteFileFromBucket(gm)
	if err != nil {
		return nil, err
	}
	return gm.toRoot(), nil
}

// FileDeleteMany is used to delete a GridFS File
func (p *FileService) FileDeleteMany(g []*models.File) error {
	outErrors := make([]error, len(g))
	var wg sync.WaitGroup
	wg.Add(len(g))
	for c, f := range g {
		go func(c int, f *models.File) {
			_, outErrors[c] = p.FileDelete(f)
			wg.Done()
		}(c, f)
	}
	wg.Wait()
	for _, err := range outErrors {
		if err != nil {
			return err
		}
	}
	return nil
}

// RetrieveFile returns the content bytes for a GridFS File
func (p *FileService) RetrieveFile(g *models.File) (*bytes.Buffer, error) {
	err := g.Validate("retrieve")
	if err != nil {
		return nil, err
	}
	gm, err := newFileModel(g)
	if err != nil {
		return nil, err
	}
	if g.CheckID("gridfs_id") {
		return p.downloadFileFromBucket(gm)
	}
	if g.CheckID("id") {
		gm, err = p.fileHandler.FindOne(gm)
		if err != nil {
			return nil, errors.New("file not found")
		}
		return p.downloadFileFromBucket(gm)
	}
	return nil, errors.New("file not found")
}

// FilesQuery is used for a paginated files search
func (p *FileService) FilesQuery(ctx context.Context, g *models.File, pagination *utilities.Pagination) (*models.FilesRes, error) {
	um, err := newFileModel(g)
	if err != nil {
		return nil, err
	}
	f, err := um.bsonFilter()
	if err != nil {
		return nil, err
	}
	count, err := p.collection.CountDocuments(ctx, f)
	if err != nil {
		return nil, err
	}
	if count == 0 {
		return &models.FilesRes{
			TotalCount: 0,
			TotalPages: 0,
			Page:       0,
			Size:       0,
			HasMore:    false,
			Files:      make([]*models.File, 0),
		}, nil
	}
	ums, err := p.fileHandler.PaginatedFind(ctx, um, pagination)
	if err != nil {
		return nil, err
	}
	files := rootFiles(ums)
	return &models.FilesRes{
		TotalCount: count,
		TotalPages: int64(pagination.GetTotalPages(int(count))),
		Page:       int64(pagination.GetPage()),
		Size:       int64(pagination.GetSize()),
		HasMore:    pagination.GetHasMore(int(count)),
		Files:      files,
	}, nil
}
