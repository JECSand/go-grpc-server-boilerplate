package database

import (
	"context"
	"errors"
	"github.com/JECSand/go-grpc-server-boilerplate/models"
	"github.com/JECSand/go-grpc-server-boilerplate/utilities"
	"time"
)

// TaskService is used by the app to manage all Task related controllers and functionality
type TaskService struct {
	collection   DBCollection
	db           DBClient
	taskHandler  *DBHandler[*taskModel]
	userHandler  *DBHandler[*userModel]
	groupHandler *DBHandler[*groupModel]
}

// NewTaskService is an exported function used to initialize a new TaskService struct
func NewTaskService(db DBClient, tHandler *DBHandler[*taskModel], uHandler *DBHandler[*userModel], gHandler *DBHandler[*groupModel]) *TaskService {
	collection := db.GetCollection("tasks")
	return &TaskService{collection, db, tHandler, uHandler, gHandler}
}

// checkLinkedRecords ensures the userId and groupId in the models.Task is correct
func (p *TaskService) checkLinkedRecords(g *groupModel, u *userModel) error {
	ctx, cancel := context.WithCancel(context.Background())
	groupChan := make(chan *groupModel)
	userChan := make(chan *userModel)
	errChan := make(chan error)
	defer func() {
		cancel()
		close(groupChan)
		close(userChan)
		close(errChan)
	}()
	go func() {
		out, err := p.groupHandler.FindOne(g)
		select {
		case <-ctx.Done():
			return
		default:
		}
		groupChan <- out
		errChan <- err
	}()
	go func() {
		out, err := p.userHandler.FindOne(u)
		select {
		case <-ctx.Done():
			return
		default:
		}
		userChan <- out
		errChan <- err
	}()
	var err error
	for i := 0; i < 4; i++ {
		select {
		case group := <-groupChan:
			g = group
		case user := <-userChan:
			u = user
		case err = <-errChan:
			if err != nil {
				err = errors.New("invalid user id")
				break
			}
		}
	}
	if g.Id != u.GroupId {
		return errors.New("task user is not in task group")
	}
	return err
}

// TaskCreate is used to create a new user Task
func (p *TaskService) TaskCreate(g *models.Task) (*models.Task, error) {
	err := g.Validate("create")
	if err != nil {
		return nil, err
	}
	gm, err := newTaskModel(g)
	if err != nil {
		return nil, err
	}
	err = p.checkLinkedRecords(&groupModel{Id: gm.GroupId}, &userModel{Id: gm.UserId})
	if err != nil {
		return nil, err
	}
	gm.Status = models.NOT_STARTED
	gm, err = p.taskHandler.InsertOne(gm)
	if err != nil {
		return nil, err
	}
	return gm.toRoot(), err
}

// TasksFind is used to find all Task docs in a MongoDB Collection
func (p *TaskService) TasksFind(g *models.Task) ([]*models.Task, error) {
	var tasks []*models.Task
	tm, err := newTaskModel(g)
	if err != nil {
		return tasks, err
	}
	gms, err := p.taskHandler.FindMany(tm)
	if err != nil {
		return tasks, err
	}
	for _, gm := range gms {
		tasks = append(tasks, gm.toRoot())
	}
	return tasks, nil
}

// TaskFind is used to find a specific Task doc
func (p *TaskService) TaskFind(g *models.Task) (*models.Task, error) {
	gm, err := newTaskModel(g)
	if err != nil {
		return nil, err
	}
	gm, err = p.taskHandler.FindOne(gm)
	if err != nil {
		return nil, err
	}
	return gm.toRoot(), err
}

// TaskDelete is used to delete a Task doc
func (p *TaskService) TaskDelete(g *models.Task) (*models.Task, error) {
	gm, err := newTaskModel(g)
	if err != nil {
		return nil, err
	}
	gm, err = p.taskHandler.DeleteOne(gm)
	if err != nil {
		return nil, err
	}
	return gm.toRoot(), err
}

// TaskDeleteMany is used to delete many Tasks
func (p *TaskService) TaskDeleteMany(g *models.Task) (*models.Task, error) {
	gm, err := newTaskModel(g)
	if err != nil {
		return nil, err
	}
	gm, err = p.taskHandler.DeleteMany(gm)
	if err != nil {
		return nil, err
	}
	return gm.toRoot(), err
}

// TaskUpdate is used to update an existing Task
func (p *TaskService) TaskUpdate(g *models.Task) (*models.Task, error) {
	var filter models.Task
	err := g.Validate("update")
	if err != nil {
		return nil, err
	}
	filter.Id = g.Id
	f, err := newTaskModel(&filter)
	if err != nil {
		return nil, err
	}
	cur, TaskErr := p.taskHandler.FindOne(f)
	if TaskErr != nil {
		return nil, errors.New("task not found")
	}
	g.BuildUpdate(cur.toRoot())
	gm, err := newTaskModel(g)
	if err != nil {
		return nil, err
	}
	err = p.checkLinkedRecords(&groupModel{Id: gm.GroupId}, &userModel{Id: gm.UserId})
	if err != nil {
		return nil, err
	}
	gm, err = p.taskHandler.UpdateOne(f, gm)
	if err != nil {
		return nil, err
	}
	return gm.toRoot(), err
}

// TaskDocInsert is used to insert a Task doc directly into mongodb for testing purposes
func (p *TaskService) TaskDocInsert(g *models.Task) (*models.Task, error) {
	insertTask, err := newTaskModel(g)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_, err = p.collection.InsertOne(ctx, insertTask)
	if err != nil {
		return nil, err
	}
	return insertTask.toRoot(), nil
}

// TasksQuery is used for a paginated tasks search
func (p *TaskService) TasksQuery(ctx context.Context, g *models.Task, pagination *utilities.Pagination) (*models.TasksRes, error) {
	um, err := newTaskModel(g)
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
		return &models.TasksRes{
			TotalCount: 0,
			TotalPages: 0,
			Page:       0,
			Size:       0,
			HasMore:    false,
			Tasks:      make([]*models.Task, 0),
		}, nil
	}
	ums, err := p.taskHandler.PaginatedFind(ctx, um, pagination)
	if err != nil {
		return nil, err
	}
	tasks := rootTasks(ums)
	return &models.TasksRes{
		TotalCount: count,
		TotalPages: int64(pagination.GetTotalPages(int(count))),
		Page:       int64(pagination.GetPage()),
		Size:       int64(pagination.GetSize()),
		HasMore:    pagination.GetHasMore(int(count)),
		Tasks:      tasks,
	}, nil
}
