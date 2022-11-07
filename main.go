package main

import "github.com/JECSand/go-grpc-server-boilerplate/cmd"

func main() {
	var app cmd.App
	err := app.Initialize()
	if err != nil {
		panic(err)
	}
	app.Run()
}
