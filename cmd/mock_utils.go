package cmd

import (
	"encoding/json"
	"github.com/JECSand/go-grpc-server-boilerplate/models"
	"time"
)

// createTestToken create a JWT Token for use by the integration tests
func createTestToken(ta *App, user *models.User) (string, error) {
	if !user.CheckID("id") { // generate bad JWT token
		return "111111111111111111111111111", nil
	}
	newToken, err := ta.server.TokenService.GenerateToken(user, "session")
	if err != nil {
		return "", err
	}
	return newToken, nil
}

// CreateTestGroup creates a group doc for test setup
func createTestGroup(ta *App, groupType int) *models.Group {
	group := models.Group{}
	if groupType == 1 {
		group.Id = "000000000000000000000002"
		group.Name = "test2"
		group.RootAdmin = false
		group.LastModified = time.Now().UTC()
		group.CreatedAt = time.Now().UTC()
	} else {
		group.Id = "000000000000000000000003"
		group.Name = "test3"
		group.RootAdmin = false
		group.LastModified = time.Now().UTC()
		group.CreatedAt = time.Now().UTC()
	}
	_, err := ta.server.GroupDataService.GroupDocInsert(&group)
	if err != nil {
		panic(err)
	}
	return &group
}

// createTestUser creates a user doc for test setup
func createTestUser(ta *App, userType int) *models.User {
	user := models.User{}
	if userType == 1 {
		user.Id = "000000000000000000000012"
		user.Username = "test_user"
		user.Password = "abc123"
		user.FirstName = "Jill"
		user.LastName = "Tester"
		user.Email = "test2@email.com"
		user.Role = "member"
		user.RootAdmin = false
		user.GroupId = "000000000000000000000002"
		user.LastModified = time.Now().UTC()
		user.CreatedAt = time.Now().UTC()
	} else {
		user.Id = "000000000000000000000013"
		user.Username = "test_user2"
		user.Password = "abc123"
		user.FirstName = "Bill"
		user.LastName = "Quality"
		user.Email = "test3@email.com.com"
		user.Role = "member"
		user.RootAdmin = false
		user.GroupId = "000000000000000000000003"
		user.LastModified = time.Now().UTC()
		user.CreatedAt = time.Now().UTC()
	}
	_, err := ta.server.UserDataService.UserDocInsert(&user)
	if err != nil {
		panic(err)
	}
	return &user
}

// createTestAdminUser creates an admin user doc for test setup
func createTestAdminUser(ta *App, userType int, root bool) *models.User {
	user := models.User{}
	if userType == 1 {
		user.Id = "000000000000000000000014"
		user.Username = "test_admin"
		user.Password = "abc123"
		user.FirstName = "Jill"
		user.LastName = "Admin"
		user.Email = "admin1@email.com"
		user.Role = "admin"
		user.RootAdmin = root
		user.GroupId = "000000000000000000000002"
		user.LastModified = time.Now().UTC()
		user.CreatedAt = time.Now().UTC()
	} else {
		user.Id = "000000000000000000000015"
		user.Username = "test_admin2"
		user.Password = "abc123"
		user.FirstName = "Bill"
		user.LastName = "Admin"
		user.Email = "admin2@email.com.com"
		user.Role = "admin"
		user.RootAdmin = root
		user.GroupId = "000000000000000000000003"
		user.LastModified = time.Now().UTC()
		user.CreatedAt = time.Now().UTC()
	}
	_, err := ta.server.UserDataService.UserDocInsert(&user)
	if err != nil {
		panic(err)
	}
	return &user
}

// createTestTask creates a task doc for test setup
func createTestTask(ta *App, taskType int) *models.Task {
	task := models.Task{}
	now := time.Now()
	if taskType == 1 {
		task.Id = "000000000000000000000021"
		task.Name = "testTask"
		task.Status = models.NOT_STARTED
		task.Due = now.Add(time.Hour * 24).UTC()
		task.Description = "Updated Task to complete"
		task.UserId = "000000000000000000000014"
		task.GroupId = "000000000000000000000002"
		task.LastModified = now.UTC()
		task.CreatedAt = now.UTC()
	} else if taskType == 2 {
		task.Id = "000000000000000000000022"
		task.Name = "testTask2"
		task.Status = models.NOT_STARTED
		task.Due = now.Add(time.Hour * 24).UTC()
		task.Description = "Updated Task to complete"
		task.UserId = "000000000000000000000014"
		task.GroupId = "000000000000000000000002"
		task.LastModified = now.UTC()
		task.CreatedAt = now.UTC()
	} else {
		task.Id = "000000000000000000000023"
		task.Name = "testTask3"
		task.Status = models.NOT_STARTED
		task.Due = now.Add(time.Hour * 48).UTC()
		task.Description = "Updated Task to complete2"
		task.UserId = "000000000000000000000012"
		task.GroupId = "000000000000000000000002"
		task.LastModified = now.UTC()
		task.CreatedAt = now.UTC()
	}
	_, err := ta.server.TaskDataService.TaskDocInsert(&task)
	if err != nil {
		panic(err)
	}
	return &task
}

// getTestUserPayload
func getTestUserPayload(tCase string) []byte {
	switch tCase {
	case "CREATE":
		return []byte(`{"username":"test_user","password":"abc123","firstname":"test","lastname":"user","email":"test2@email.com","group_id":"000000000000000000000002","role":"member"}`)
	case "UPDATE":
		return []byte(`{"username":"newUserName","password":"newUserPass","email":"new_test@email.com","group_id":"000000000000000000000003","role":"member"}`)
	}
	return nil
}

// getTestPasswordPayload
func getTestPasswordPayload(tCase string) []byte {
	switch tCase {
	case "UPDATE_PASSWORD_ERROR":
		return []byte(`{"current_password":"789test122","new_password":"789test124"}`)
	case "UPDATE_PASSWORD_SUCCESS":
		return []byte(`{"current_password":"abc123","new_password":"789test124"}`)
	}
	return nil
}

// getTestGroupPayload
func getTestGroupPayload(tCase string) []byte {
	switch tCase {
	case "CREATE":
		return []byte(`{"name":"testingGroup"}`)
	case "UPDATE":
		return []byte(`{"name":"newTestingGroup"}`)
	}
	return nil
}

// getTestTaskPayload
func getTestTaskPayload(tCase string) []byte {
	var tTask models.Task
	now := time.Now()
	switch tCase {
	case "CREATE":
		tTask.Name = "testTask"
		tTask.Status = models.NOT_STARTED
		tTask.Due = now.Add(time.Hour * 24).UTC()
		tTask.Description = "Updated Task to complete"
		tTask.UserId = "000000000000000000000012"
		tTask.GroupId = "000000000000000000000002"
		b, _ := json.Marshal(tTask)
		return b
	case "UPDATE":
		tTask.Name = "NewTestTask"
		tTask.Status = models.COMPLETED
		tTask.Description = "Updated Task to complete"
		tTask.UserId = "000000000000000000000012"
		tTask.GroupId = "000000000000000000000002"
		b, _ := json.Marshal(tTask)
		return b
	}
	return nil
}
