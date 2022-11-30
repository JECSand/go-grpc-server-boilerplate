package services

import (
	"context"
	"errors"
	"github.com/JECSand/go-grpc-server-boilerplate/models"
	tasksService "github.com/JECSand/go-grpc-server-boilerplate/protos/task"
	"github.com/JECSand/go-grpc-server-boilerplate/utilities"
)

// TaskService gRPC Service
type TaskService struct {
	log          utilities.Logger
	tokenService *TokenService
	userDB       UserDataService
	groupDB      GroupDataService
	taskDB       TaskDataService
	fileDB       FileDataService
}

// NewTaskService constructs a TaskService for controller gRPC service Task requests
func NewTaskService(log utilities.Logger, ts *TokenService, u UserDataService, g GroupDataService, t TaskDataService, f FileDataService) *TaskService {
	return &TaskService{
		log:          log,
		tokenService: ts,
		userDB:       u,
		groupDB:      g,
		taskDB:       t,
		fileDB:       f,
	}
}

// Create a New Task
func (u *TaskService) Create(ctx context.Context, req *tasksService.CreateReq) (*tasksService.CreateRes, error) {
	task := models.LoadTaskCreateProto(req)
	userScope, err := models.VerifyRequestScope(ctx, "create")
	if err != nil {
		u.log.Errorf("models.VerifyRequestScope: %v", err)
		return nil, utilities.ErrorResponse(err, err.Error())
	}
	task.LoadScope(userScope)
	task.Id = utilities.GenerateObjectID()
	if !task.CheckID("user_id") || !task.CheckID("group_id") {
		tokenClaims, err := models.LoadTokenFromContext(ctx)
		if err != nil {
			u.log.Errorf("models.LoadTokenFromContext: %v", err)
			return nil, utilities.ErrorResponse(err, err.Error())
		}
		if !task.CheckID("user_id") {
			task.UserId = tokenClaims.UserId
		}
		if !task.CheckID("group_id") {
			task.GroupId = tokenClaims.GroupId
		}
	}
	task, err = u.taskDB.TaskCreate(task)
	if err != nil {
		u.log.Errorf("taskDB.TaskCreate: %v", err)
		return nil, utilities.ErrorResponse(err, err.Error())
	}
	return &tasksService.CreateRes{Task: task.ToProto()}, nil
}

// Update a Task
func (u *TaskService) Update(ctx context.Context, req *tasksService.UpdateReq) (*tasksService.UpdateRes, error) {
	var err error
	if !utilities.CheckObjectID(req.GetId()) {
		err = errors.New(req.GetId() + " is an invalid taskId")
		u.log.Errorf("utilities.CheckObjectID: %v", err)
		return nil, utilities.ErrorResponse(err, err.Error())
	}
	task := models.LoadTaskUpdateProto(req)
	task, err = u.taskDB.TaskUpdate(task)
	if err != nil {
		u.log.Errorf("taskDB.TaskUpdate: %v", err)
		return nil, utilities.ErrorResponse(err, err.Error())
	}
	return &tasksService.UpdateRes{Task: task.ToProto()}, nil
}

// Get a specific Task
func (u *TaskService) Get(ctx context.Context, req *tasksService.GetReq) (*tasksService.GetRes, error) {
	if !utilities.CheckObjectID(req.GetId()) {
		err := errors.New(req.GetId() + " is an invalid taskId")
		u.log.Errorf("utilities.CheckObjectID: %v", err)
		return nil, utilities.ErrorResponse(err, err.Error())
	}
	filter := models.Task{Id: req.GetId()}
	userScope, err := models.VerifyRequestScope(ctx, "find")
	if err != nil {
		u.log.Errorf("models.VerifyRequestScope: %v", err)
		return nil, utilities.ErrorResponse(err, err.Error())
	}
	filter.LoadScope(userScope)
	task, err := u.taskDB.TaskFind(&filter)
	if err != nil {
		u.log.Errorf("taskDB.TaskFind: %v", err)
		return nil, utilities.ErrorResponse(err, err.Error())
	}
	return &tasksService.GetRes{Task: task.ToProto()}, nil
}

// Find Tasks from an input query
func (u *TaskService) Find(ctx context.Context, req *tasksService.FindReq) (*tasksService.FindRes, error) {
	// TODO NEXT FIX - valid req.GetQuery() authenticity / scope
	tasks, err := u.taskDB.TasksQuery(ctx, models.LoadTaskFindProto(req), utilities.NewPaginationQuery(int(req.GetSize()), int(req.GetPage())))
	if err != nil {
		u.log.Errorf("taskDB.TasksQuery: %v", err)
		return nil, utilities.ErrorResponse(err, err.Error())
	}
	return &tasksService.FindRes{
		TotalCount: tasks.TotalCount,
		TotalPages: tasks.TotalPages,
		Page:       tasks.Page,
		Size:       tasks.Size,
		HasMore:    tasks.HasMore,
		Tasks:      tasks.ToProto(),
	}, nil
}

// GetGroupTasks returns the tasks for a given groupId
func (u *TaskService) GetGroupTasks(ctx context.Context, req *tasksService.GetGroupTasksReq) (*tasksService.GetGroupTasksRes, error) {
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
	tasks, err := u.taskDB.TasksQuery(ctx, &models.Task{GroupId: groupId}, utilities.NewPaginationQuery(int(req.GetSize()), int(req.GetPage())))
	if err != nil {
		u.log.Errorf("taskDB.TasksQuery: %v", err)
		return nil, utilities.ErrorResponse(err, err.Error())
	}
	return &tasksService.GetGroupTasksRes{
		TotalCount: tasks.TotalCount,
		TotalPages: tasks.TotalPages,
		Page:       tasks.Page,
		Size:       tasks.Size,
		HasMore:    tasks.HasMore,
		Tasks:      tasks.ToProto(),
	}, nil
}

// GetUserTasks returns the tasks for a given userId
func (u *TaskService) GetUserTasks(ctx context.Context, req *tasksService.GetUserTasksReq) (*tasksService.GetUserTasksRes, error) {
	if !utilities.CheckObjectID(req.GetUserId()) {
		err := errors.New("invalid user id")
		u.log.Errorf("utilities.CheckObjectID: %v", err)
		return nil, utilities.ErrorResponse(err, err.Error())
	}
	user, err := u.userDB.UserFind(&models.User{Id: req.GetUserId()})
	if err != nil {
		u.log.Errorf("userDB.UserFind: %v", err)
		return nil, utilities.ErrorResponse(err, err.Error())
	}
	userScope, err := models.VerifyRequestScope(ctx, "find")
	if err != nil {
		u.log.Errorf("models.VerifyRequestScope: %v", err)
		return nil, utilities.ErrorResponse(err, err.Error())
	}
	if userScope.GroupId != "" && userScope.GroupId != user.GroupId {
		u.log.Errorf("taskDB.TasksQuery: %v", err)
		return nil, utilities.ErrorResponse(err, err.Error())
	}
	tasks, err := u.taskDB.TasksQuery(ctx, &models.Task{UserId: user.Id}, utilities.NewPaginationQuery(int(req.GetSize()), int(req.GetPage())))
	if err != nil {
		u.log.Errorf("taskDB.TasksQuery: %v", err)
		return nil, utilities.ErrorResponse(err, err.Error())
	}
	return &tasksService.GetUserTasksRes{
		TotalCount: tasks.TotalCount,
		TotalPages: tasks.TotalPages,
		Page:       tasks.Page,
		Size:       tasks.Size,
		HasMore:    tasks.HasMore,
		Tasks:      tasks.ToProto(),
	}, nil
}

// Delete is the handler function that deletes a task
func (u *TaskService) Delete(ctx context.Context, req *tasksService.DeleteReq) (*tasksService.DeleteRes, error) {
	if !utilities.CheckObjectID(req.GetId()) {
		err := errors.New(req.GetId() + " is an invalid taskId")
		u.log.Errorf("utilities.CheckObjectID: %v", err)
		return nil, utilities.ErrorResponse(err, err.Error())
	}
	filter := models.Task{Id: req.GetId()}
	userScope, err := models.VerifyRequestScope(ctx, "update")
	if err != nil {
		u.log.Errorf("models.VerifyRequestScope: %v", err)
		return nil, utilities.ErrorResponse(err, err.Error())
	}
	filter.LoadScope(userScope)
	task, err := u.taskDB.TaskDelete(&filter)
	if err != nil {
		u.log.Errorf("taskDB.TaskDelete: %v", err)
		return nil, utilities.ErrorResponse(err, err.Error())
	}
	return &tasksService.DeleteRes{Task: task.ToProto()}, nil
}
