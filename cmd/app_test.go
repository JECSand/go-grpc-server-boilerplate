package cmd

import (
	"context"
	authsService "github.com/JECSand/go-grpc-server-boilerplate/protos/auth"
	"os"
	"testing"
)

var ta App

// Setup Tests
func setup() {
	os.Setenv("ENV", "test")
	ta = App{}
	err := ta.Initialize()
	if err != nil {
		panic(err)
	}
}

/*
AUTH TESTS
*/

// TODO MAKE TEST TABLE DRIVEN
// User TestLogin Test
func TestLogin(t *testing.T) {
	ctx := context.Background()
	setup()
	conn, closer := ta.server.StartTest(ctx)
	defer closer()
	client := authsService.NewAuthServiceClient(conn)
	testReq := &authsService.LoginReq{Email: os.Getenv("ROOT_EMAIL"), Password: os.Getenv("ROOT_PASSWORD")}
	out, err := client.Login(ctx, testReq)
	if err != nil {
		t.Errorf("Err -> \nWant: %q\nGot: %q\n", "", err)
	}
	if out.AccessToken == "" {
		t.Errorf("Err -> \nWant: %q\nGot: %q\n", "accessToken", "")
	}
}
