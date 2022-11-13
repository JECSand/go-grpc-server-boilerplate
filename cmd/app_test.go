package cmd

import (
	"context"
	authsService "github.com/JECSand/go-grpc-server-boilerplate/protos/auth"
	"os"
	"testing"
)

// Setup Tests
func setup() App {
	os.Setenv("ENV", "test")
	ta := App{}
	err := ta.Initialize()
	if err != nil {
		panic(err)
	}
	return ta
}

/*
AUTH TESTS
*/

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
			//fmt.Println("\n\nCHECK REQ: ", tt.req, os.Getenv("ROOT_EMAIL"), os.Getenv("ROOT_PASSWORD"))
			out, err := client.Login(ctx, tt.req)
			if (err != nil) != tt.wantErr {
				//fmt.Println("\n\nCHECK ERROR: ", err)
				t.Errorf("authsService.Login() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//fmt.Println("\n\nCHECK OUT: ", out)
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
