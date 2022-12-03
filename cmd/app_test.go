package cmd

import (
	"context"
	"fmt"
	"github.com/JECSand/go-grpc-server-boilerplate/models"
	authsService "github.com/JECSand/go-grpc-server-boilerplate/protos/auth"
	usersService "github.com/JECSand/go-grpc-server-boilerplate/protos/user"
	"github.com/JECSand/go-grpc-server-boilerplate/utilities"
	"os"
	"testing"
)

// Setup Tests
func setup() *App {
	err := os.Setenv("ENV", "test")
	if err != nil {
		fmt.Println("\n\n-------->ERROR CHECK HERE A, ", err.Error())
	}
	ta := &App{}
	err = ta.Initialize()
	if err != nil {
		fmt.Println("\n\n-------->ERROR CHECK HERE B, ", err.Error())
		panic(err)
	}
	return ta
}

func setupTestUser(ta *App, group bool, tType int) *models.User {
	if group {
		_ = createTestGroup(ta, tType)
	}
	return createTestUser(ta, tType)
}

func setupTestAdminUser(ta *App, root bool, group bool, tType int) *models.User {
	if group {
		_ = createTestGroup(ta, tType)
	}
	return createTestAdminUser(ta, tType, root)
}

func setupTestAuthCtx(ta *App, ctx context.Context, tUser *models.User, scenario string) context.Context {
	var authToken string
	if scenario == "missing" {
		return ctx
	} else if scenario == "invalid" {
		authToken, _ = createTestToken(ta, &models.User{})
	} else {
		authToken, _ = createTestToken(ta, tUser)
	}
	ctx = context.Background()
	return utilities.AttachTokenToContext(ctx, authToken)
}

/*
AUTH TESTS
*/

func Test_AuthRegister(t *testing.T) {
	ctx := context.Background()
	ta := setup()
	conn, closer := ta.server.StartTest(ctx)
	client := authsService.NewAuthServiceClient(conn)
	defer closer()
	// Defining our test slice. Each unit test should have the following properties:
	tests := []struct {
		name    string                    // The name of the test
		res     *authsService.RegisterRes // What out instance we want our function to return.
		wantErr bool                      // whether we want an error.
		req     *authsService.RegisterReq // The input of the test
	}{
		// Here we're declaring each unit test input and output data as defined before
		{
			"success",
			&authsService.RegisterRes{User: &authsService.User{Username: "tester123"}},
			false,
			&authsService.RegisterReq{
				FirstName: "Jack",
				LastName:  "Testings",
				Email:     "tester@test.com",
				Username:  "tester123",
				Password:  "321test123",
			},
		},
		{
			"taken email",
			nil,
			true,
			&authsService.RegisterReq{
				FirstName: "Jack",
				LastName:  "Testings",
				Email:     "master@test.com",
				Username:  "tester123",
				Password:  "321test123",
			},
		},
		{
			"missing email",
			nil,
			true,
			&authsService.RegisterReq{
				FirstName: "Jack",
				LastName:  "Testings",
				Email:     "",
				Username:  "tester123",
				Password:  "321test123",
			},
		},
		{
			"missing password",
			nil,
			true,
			&authsService.RegisterReq{
				FirstName: "Jack",
				LastName:  "Testings",
				Email:     "tester",
				Username:  "tester123",
				Password:  "",
			},
		},
		{
			"missing username",
			nil,
			true,
			&authsService.RegisterReq{
				FirstName: "Jack",
				LastName:  "Testings",
				Email:     "tester",
				Username:  "",
				Password:  "321test123",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := client.Register(ctx, tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("authsService.Register() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			switch tt.name {
			case "success":
				if out.User.Username != tt.res.User.Username || out.User.Id == "" {
					t.Errorf("authsService.Register() \nWant: %q\nGot: %q\n", out.User.Username, tt.res.User.Username)
				}
			default:
				if out != tt.res { // Asserting whether we get the correct wanted value
					t.Errorf("authsService.Register() \nWant: %q\\nGot: %q\n", out, tt.res)
				}
			}
		})
	}
}

func Test_AuthLogin(t *testing.T) {
	ctx := context.Background()
	ta := setup()
	conn, closer := ta.server.StartTest(ctx)
	client := authsService.NewAuthServiceClient(conn)
	defer closer()
	// Defining our test slice. Each unit test should have the following properties:
	tests := []struct {
		name    string                 // The name of the test
		res     *authsService.LoginRes // What out instance we want our function to return.
		wantErr bool                   // whether we want an error.
		req     *authsService.LoginReq // The input of the test
	}{
		// Here we're declaring each unit test input and output data as defined before
		{
			"success",
			&authsService.LoginRes{User: &authsService.User{Username: "MasterAdmin"}},
			false,
			&authsService.LoginReq{Email: "master@test.com", Password: "321test123"},
		},
		{
			"incorrect password",
			nil,
			true,
			&authsService.LoginReq{Email: "master@test.com", Password: "wrong"},
		},
		{
			"incorrect email",
			nil,
			true,
			&authsService.LoginReq{Email: "wrong", Password: "wrong"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := client.Login(ctx, tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("authsService.Login() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			switch tt.name {
			case "success":
				if out.User.Username != tt.res.User.Username {
					t.Errorf("authsService.Login() \nWant: %q\nGot: %q\n", out.User.Username, tt.res.User.Username)
				}
			case "incorrect password":
				if out != tt.res { // Asserting whether we get the correct wanted value
					t.Errorf("authsService.Login() \nWant: %q\\nGot: %q\n", out, tt.res)
				}
			case "incorrect email":
				if out != tt.res { // Asserting whether we get the correct wanted value
					t.Errorf("authsService.Login() \nWant: %q\\nGot: %q\n", out, tt.res)
				}
			}
		})
	}
}

func Test_AuthLogout(t *testing.T) {
	ctx := context.Background()
	ta := setup()
	conn, closer := ta.server.StartTest(ctx)
	client := authsService.NewAuthServiceClient(conn)
	defer closer()
	tUser := setupTestUser(ta, true, 1)
	// Defining our test slice. Each unit test should have the following properties:
	tests := []struct {
		name    string                  // The name of the test
		res     *authsService.LogoutRes // What out instance we want our function to return.
		wantErr bool                    // whether we want an error.
		req     *authsService.Empty     // The input of the test
	}{
		// Here we're declaring each unit test input and output data as defined before
		{
			"missing token",
			nil,
			true,
			&authsService.Empty{},
		},
		{
			"invalid token",
			nil,
			true,
			&authsService.Empty{},
		},
		{
			"success",
			&authsService.LogoutRes{Status: 200},
			false,
			&authsService.Empty{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.name {
			case "success":
				ctx = setupTestAuthCtx(ta, ctx, tUser, "")
			case "invalid token":
				ctx = setupTestAuthCtx(ta, ctx, tUser, "invalid")
			default:
				ctx = setupTestAuthCtx(ta, ctx, tUser, "missing")
			}
			out, err := client.Logout(ctx, tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("authsService.Logout() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			switch tt.name {
			case "success":
				if out.Status != tt.res.Status {
					t.Errorf("authsService.Logout() \nWant: %q\nGot: %q\n", out.Status, tt.res.Status)
				}
			default:
				if out != tt.res { // Asserting whether we get the correct wanted value
					t.Errorf("authsService.Logout() \nWant: %q\\nGot: %q\n", out, tt.res)
				}
			}
		})
	}
}

func Test_AuthRefresh(t *testing.T) {
	ctx := context.Background()
	ta := setup()
	conn, closer := ta.server.StartTest(ctx)
	client := authsService.NewAuthServiceClient(conn)
	defer closer()
	tUser := setupTestUser(ta, true, 1)
	// Defining our test slice. Each unit test should have the following properties:
	tests := []struct {
		name    string                   // The name of the test
		res     *authsService.RefreshRes // What out instance we want our function to return.
		wantErr bool                     // whether we want an error.
		req     *authsService.Empty      // The input of the test
	}{
		// Here we're declaring each unit test input and output data as defined before
		{
			"missing token",
			nil,
			true,
			&authsService.Empty{},
		},
		{
			"invalid token",
			nil,
			true,
			&authsService.Empty{},
		},
		{
			"success",
			&authsService.RefreshRes{AccessToken: ""},
			false,
			&authsService.Empty{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.name {
			case "success":
				ctx = setupTestAuthCtx(ta, ctx, tUser, "")
			case "invalid token":
				ctx = setupTestAuthCtx(ta, ctx, tUser, "invalid")
			default:
				ctx = setupTestAuthCtx(ta, ctx, tUser, "missing")
			}
			out, err := client.Refresh(ctx, tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("authsService.Refresh() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			switch tt.name {
			case "success":
				if out.AccessToken == tt.res.AccessToken {
					t.Errorf("authsService.Refresh() \nDon't Want: %q\nGot: %q\n", out.AccessToken, tt.res.AccessToken)
				}
			default:
				if out != tt.res { // Asserting whether we get the correct wanted value
					t.Errorf("authsService.Refresh() \nWant: %q\\nGot: %q\n", out, tt.res)
				}
			}
		})
	}
}

func Test_AuthGenerateKey(t *testing.T) {
	ctx := context.Background()
	ta := setup()
	conn, closer := ta.server.StartTest(ctx)
	client := authsService.NewAuthServiceClient(conn)
	defer closer()
	tUser := setupTestUser(ta, true, 1)
	// Defining our test slice. Each unit test should have the following properties:
	tests := []struct {
		name    string                       // The name of the test
		res     *authsService.GenerateKeyRes // What out instance we want our function to return.
		wantErr bool                         // whether we want an error.
		req     *authsService.Empty          // The input of the test
	}{
		// Here we're declaring each unit test input and output data as defined before
		{
			"missing token",
			nil,
			true,
			&authsService.Empty{},
		},
		{
			"invalid token",
			nil,
			true,
			&authsService.Empty{},
		},
		{
			"success",
			&authsService.GenerateKeyRes{APIKey: ""},
			false,
			&authsService.Empty{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.name {
			case "success":
				ctx = setupTestAuthCtx(ta, ctx, tUser, "")
			case "invalid token":
				ctx = setupTestAuthCtx(ta, ctx, tUser, "invalid")
			default:
				ctx = setupTestAuthCtx(ta, ctx, tUser, "missing")
			}
			out, err := client.GenerateKey(ctx, tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("authsService.Refresh() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			switch tt.name {
			case "success":
				if out.APIKey == tt.res.APIKey {
					t.Errorf("authsService.Refresh() \nDon't Want: %q\nGot: %q\n", out.APIKey, tt.res.APIKey)
				}
			default:
				if out != tt.res { // Asserting whether we get the correct wanted value
					t.Errorf("authsService.Refresh() \nWant: %q\\nGot: %q\n", out, tt.res)
				}
			}
		})
	}
}

func Test_AuthUpdatePassword(t *testing.T) {
	ctx := context.Background()
	ta := setup()
	conn, closer := ta.server.StartTest(ctx)
	client := authsService.NewAuthServiceClient(conn)
	defer closer()
	tUser := setupTestUser(ta, true, 1)
	// Defining our test slice. Each unit test should have the following properties:
	tests := []struct {
		name    string                          // The name of the test
		res     *authsService.UpdatePasswordRes // What out instance we want our function to return.
		wantErr bool                            // whether we want an error.
		req     *authsService.UpdatePasswordReq // The input of the test
	}{
		// Here we're declaring each unit test input and output data as defined before
		{
			"invalid token",
			nil,
			true,
			&authsService.UpdatePasswordReq{
				NewPassword:     "abc124",
				CurrentPassword: "abc123",
			},
		},
		{
			"missing currentPassword",
			nil,
			true,
			&authsService.UpdatePasswordReq{
				NewPassword: "abc124",
			},
		},
		{
			"missing newPassword",
			nil,
			true,
			&authsService.UpdatePasswordReq{
				CurrentPassword: "abc123",
			},
		},
		{
			"incorrect password",
			nil,
			true,
			&authsService.UpdatePasswordReq{
				NewPassword:     "abc124",
				CurrentPassword: "abc423",
			},
		},
		{
			"same passwords",
			nil,
			true,
			&authsService.UpdatePasswordReq{
				NewPassword:     "abc123",
				CurrentPassword: "abc123",
			},
		},
		{
			"success",
			&authsService.UpdatePasswordRes{Status: 200},
			false,
			&authsService.UpdatePasswordReq{
				NewPassword:     "abc124",
				CurrentPassword: "abc123",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.name {
			case "missing token":
				ctx = setupTestAuthCtx(ta, ctx, tUser, "missing")
			case "invalid token":
				ctx = setupTestAuthCtx(ta, ctx, tUser, "invalid")
			default:
				ctx = setupTestAuthCtx(ta, ctx, tUser, "")
			}
			out, err := client.UpdatePassword(ctx, tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("authsService.UpdatePassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			switch tt.name {
			case "success":
				if out.Status != tt.res.Status {
					t.Errorf("authsService.UpdatePassword() \nWant: %q\nGot: %q\n", out.Status, tt.res.Status)
				}
			default:
				if out != tt.res { // Asserting whether we get the correct wanted value
					t.Errorf("authsService.UpdatePassword() \nWant: %q\\nGot: %q\n", out, tt.res)
				}
			}
		})
	}
}

/*
USER TESTS
*/

func Test_UserCreate(t *testing.T) {
	ctx := context.Background()
	ta := setup()
	conn, closer := ta.server.StartTest(ctx)
	client := usersService.NewUserServiceClient(conn)
	defer closer()
	tUser := setupTestAdminUser(ta, false, true, 1)
	// Defining our test slice. Each unit test should have the following properties:
	tests := []struct {
		name    string                  // The name of the test
		res     *usersService.CreateRes // What out instance we want our function to return.
		wantErr bool                    // whether we want an error.
		req     *usersService.CreateReq // The input of the test
	}{
		// Here we're declaring each unit test input and output data as defined before
		{
			"success",
			&usersService.CreateRes{User: &usersService.User{Username: "tester123"}},
			false,
			&usersService.CreateReq{
				FirstName: "Jack",
				LastName:  "Testings",
				Email:     "tester@test.com",
				Username:  "tester123",
				Password:  "321test123",
			},
		},
		{
			"taken email",
			nil,
			true,
			&usersService.CreateReq{
				FirstName: "Jack",
				LastName:  "Testings",
				Email:     "master@test.com",
				Username:  "tester123",
				Password:  "321test123",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.name {
			case "missing token":
				ctx = setupTestAuthCtx(ta, ctx, tUser, "missing")
			case "invalid token":
				ctx = setupTestAuthCtx(ta, ctx, tUser, "invalid")
			default:
				ctx = setupTestAuthCtx(ta, ctx, tUser, "")
			}
			out, err := client.Create(ctx, tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("usersService.Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			switch tt.name {
			case "success":
				if out.User.Username != tt.res.User.Username || out.User.Id == "" {
					t.Errorf("usersService.Create() \nWant: %q\nGot: %q\n", out.User.Username, tt.res.User.Username)
				}
			default:
				if out != tt.res { // Asserting whether we get the correct wanted value
					t.Errorf("usersService.Create() \nWant: %q\\nGot: %q\n", out, tt.res)
				}
			}
		})
	}
}

func Test_UserUpdate(t *testing.T) {
	ctx := context.Background()
	ta := setup()
	conn, closer := ta.server.StartTest(ctx)
	client := usersService.NewUserServiceClient(conn)
	defer closer()
	tUser := setupTestUser(ta, true, 1)
	tAdmin := setupTestAdminUser(ta, false, false, 1)
	// Defining our test slice. Each unit test should have the following properties:
	tests := []struct {
		name    string                  // The name of the test
		res     *usersService.UpdateRes // What out instance we want our function to return.
		wantErr bool                    // whether we want an error.
		req     *usersService.UpdateReq // The input of the test
	}{
		// Here we're declaring each unit test input and output data as defined before
		{
			"success",
			&usersService.UpdateRes{User: &usersService.User{Username: "tester1233"}},
			false,
			&usersService.UpdateReq{
				Id:        tUser.Id,
				FirstName: "Jack",
				LastName:  "Testings",
				Username:  "tester1233",
				Password:  "321test123",
			},
		},
		{
			"missing id",
			nil,
			true,
			&usersService.UpdateReq{
				FirstName: "Jack",
				LastName:  "Testings",
				Email:     "master@test.com",
				Username:  "tester123",
				Password:  "321test123",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.name {
			case "missing token":
				ctx = setupTestAuthCtx(ta, ctx, tAdmin, "missing")
			case "invalid token":
				ctx = setupTestAuthCtx(ta, ctx, tAdmin, "invalid")
			default:
				ctx = setupTestAuthCtx(ta, ctx, tAdmin, "")
			}
			out, err := client.Update(ctx, tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("usersService.Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			switch tt.name {
			case "success":
				if out.User.Username != tt.res.User.Username || out.User.Id == "" {
					t.Errorf("usersService.Update() \nWant: %q\nGot: %q\n", out.User.Username, tt.res.User.Username)
				}
			default:
				if out != tt.res { // Asserting whether we get the correct wanted value
					t.Errorf("usersService.Update() \nWant: %q\\nGot: %q\n", out, tt.res)
				}
			}
		})
	}
}

func Test_UserGet(t *testing.T) {
	ctx := context.Background()
	ta := setup()
	conn, closer := ta.server.StartTest(ctx)
	client := usersService.NewUserServiceClient(conn)
	defer closer()
	tUser := setupTestUser(ta, true, 1)
	tAdmin := setupTestAdminUser(ta, false, false, 1)
	// Defining our test slice. Each unit test should have the following properties:
	tests := []struct {
		name    string               // The name of the test
		res     *usersService.GetRes // What out instance we want our function to return.
		wantErr bool                 // whether we want an error.
		req     *usersService.GetReq // The input of the test
	}{
		// Here we're declaring each unit test input and output data as defined before
		{
			"success",
			&usersService.GetRes{User: &usersService.User{Username: tUser.Username}},
			false,
			&usersService.GetReq{
				Id: tUser.Id,
			},
		},
		{
			"missing id",
			nil,
			true,
			&usersService.GetReq{},
		},
		{
			"not found",
			nil,
			true,
			&usersService.GetReq{
				Id: "000000000000000000000092",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.name {
			case "missing token":
				ctx = setupTestAuthCtx(ta, ctx, tAdmin, "missing")
			case "invalid token":
				ctx = setupTestAuthCtx(ta, ctx, tAdmin, "invalid")
			default:
				ctx = setupTestAuthCtx(ta, ctx, tAdmin, "")
			}
			out, err := client.Get(ctx, tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("usersService.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			switch tt.name {
			case "success":
				if out.User.Username != tt.res.User.Username || out.User.Id == "" {
					t.Errorf("usersService.Get() \nWant: %q\nGot: %q\n", out.User.Username, tt.res.User.Username)
				}
			default:
				if out != tt.res { // Asserting whether we get the correct wanted value
					t.Errorf("usersService.Get() \nWant: %q\\nGot: %q\n", out, tt.res)
				}
			}
		})
	}
}

func Test_UserFind(t *testing.T) {
	ctx := context.Background()
	ta := setup()
	conn, closer := ta.server.StartTest(ctx)
	client := usersService.NewUserServiceClient(conn)
	defer closer()
	tUser := setupTestUser(ta, true, 1)
	tAdmin := setupTestAdminUser(ta, false, false, 1)
	// Defining our test slice. Each unit test should have the following properties:
	tests := []struct {
		name    string                // The name of the test
		res     *usersService.FindRes // What out instance we want our function to return.
		wantErr bool                  // whether we want an error.
		req     *usersService.FindReq // The input of the test
	}{
		// Here we're declaring each unit test input and output data as defined before
		{
			"success",
			&usersService.FindRes{
				Users: []*usersService.User{{Username: tUser.Username}, {Username: tAdmin.Username}},
				Page:  1,
				Size:  2,
			},
			false,
			&usersService.FindReq{
				User: &usersService.User{GroupId: tUser.GroupId},
				Page: 1,
				Size: 10,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.name {
			case "missing token":
				ctx = setupTestAuthCtx(ta, ctx, tAdmin, "missing")
			case "invalid token":
				ctx = setupTestAuthCtx(ta, ctx, tAdmin, "invalid")
			default:
				ctx = setupTestAuthCtx(ta, ctx, tAdmin, "")
			}
			out, err := client.Find(ctx, tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("usersService.Find() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			switch tt.name {
			case "success":
				if out.Users[0].Username != tt.res.Users[0].Username {
					t.Errorf("usersService.Find() \nWant: %q\nGot: %q\n", out.Users[0].Username, tt.res.Users[0].Username)
				}
				if out.Users[1].Username != tt.res.Users[1].Username {
					t.Errorf("usersService.Find() \nWant: %q\nGot: %q\n", out.Users[1].Username, tt.res.Users[1].Username)
				}
			default:
				if out != tt.res { // Asserting whether we get the correct wanted value
					t.Errorf("usersService.Find() \nWant: %q\\nGot: %q\n", out, tt.res)
				}
			}
		})
	}
}

func Test_UserGetGroup(t *testing.T) {
	ctx := context.Background()
	ta := setup()
	conn, closer := ta.server.StartTest(ctx)
	client := usersService.NewUserServiceClient(conn)
	defer closer()
	tUser := setupTestUser(ta, true, 1)
	tAdmin := setupTestAdminUser(ta, false, false, 1)
	// Defining our test slice. Each unit test should have the following properties:
	tests := []struct {
		name    string                         // The name of the test
		res     *usersService.GetGroupUsersRes // What out instance we want our function to return.
		wantErr bool                           // whether we want an error.
		req     *usersService.GetGroupUsersReq // The input of the test
	}{
		// Here we're declaring each unit test input and output data as defined before
		{
			"success",
			&usersService.GetGroupUsersRes{
				Users: []*usersService.User{{Username: tUser.Username}, {Username: tAdmin.Username}},
				Page:  1,
				Size:  2,
			},
			false,
			&usersService.GetGroupUsersReq{
				GroupId: tUser.GroupId,
				Page:    1,
				Size:    10,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.name {
			case "missing token":
				ctx = setupTestAuthCtx(ta, ctx, tAdmin, "missing")
			case "invalid token":
				ctx = setupTestAuthCtx(ta, ctx, tAdmin, "invalid")
			default:
				ctx = setupTestAuthCtx(ta, ctx, tAdmin, "")
			}
			out, err := client.GetGroupUsers(ctx, tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("usersService.GetGroupUsers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			switch tt.name {
			case "success":
				if out.Users[0].Username != tt.res.Users[0].Username {
					t.Errorf("usersService.GetGroupUsers() \nWant: %q\nGot: %q\n", out.Users[0].Username, tt.res.Users[0].Username)
				}
				if out.Users[1].Username != tt.res.Users[1].Username {
					t.Errorf("usersService.GetGroupUsers() \nWant: %q\nGot: %q\n", out.Users[1].Username, tt.res.Users[1].Username)
				}
			default:
				if out != tt.res { // Asserting whether we get the correct wanted value
					t.Errorf("usersService.GetGroupUsers() \nWant: %q\\nGot: %q\n", out, tt.res)
				}
			}
		})
	}
}

func Test_UserDelete(t *testing.T) {
	ctx := context.Background()
	ta := setup()
	conn, closer := ta.server.StartTest(ctx)
	client := usersService.NewUserServiceClient(conn)
	defer closer()
	tUser := setupTestUser(ta, true, 1)
	tAdmin := setupTestAdminUser(ta, false, false, 1)
	// Defining our test slice. Each unit test should have the following properties:
	tests := []struct {
		name    string                  // The name of the test
		res     *usersService.DeleteRes // What out instance we want our function to return.
		wantErr bool                    // whether we want an error.
		req     *usersService.DeleteReq // The input of the test
	}{
		// Here we're declaring each unit test input and output data as defined before
		{
			"success",
			&usersService.DeleteRes{User: &usersService.User{Username: tUser.Username}},
			false,
			&usersService.DeleteReq{
				Id: tUser.Id,
			},
		},
		{
			"missing id",
			nil,
			true,
			&usersService.DeleteReq{},
		},
		{
			"not found",
			nil,
			true,
			&usersService.DeleteReq{
				Id: "000000000000000000000092",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.name {
			case "missing token":
				ctx = setupTestAuthCtx(ta, ctx, tAdmin, "missing")
			case "invalid token":
				ctx = setupTestAuthCtx(ta, ctx, tAdmin, "invalid")
			default:
				ctx = setupTestAuthCtx(ta, ctx, tAdmin, "")
			}
			out, err := client.Delete(ctx, tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("usersService.Delete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			switch tt.name {
			case "success":
				if out.User.Username != tt.res.User.Username || out.User.Id == "" {
					t.Errorf("usersService.Delete() \nWant: %q\nGot: %q\n", out.User.Username, tt.res.User.Username)
				}
			default:
				if out != tt.res { // Asserting whether we get the correct wanted value
					t.Errorf("usersService.Delete() \nWant: %q\\nGot: %q\n", out, tt.res)
				}
			}
		})
	}
}

/*
GROUP TESTS
*/

/*
TASK TESTS
*/
