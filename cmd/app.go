package cmd

import (
	"context"
	"github.com/JECSand/go-grpc-server-boilerplate/config"
	"github.com/JECSand/go-grpc-server-boilerplate/database"
	"github.com/JECSand/go-grpc-server-boilerplate/models"
	"github.com/JECSand/go-grpc-server-boilerplate/server"
	"github.com/JECSand/go-grpc-server-boilerplate/services"
	"github.com/JECSand/go-grpc-server-boilerplate/utilities"
	"go.mongodb.org/mongo-driver/bson"
	"os"
	"time"
)

// App is the highest level struct of the rest_api application. Stores the server, client, and config settings.
type App struct {
	server *server.Server
	db     database.DBClient
}

// Initialize is a function used to initialize a new instantiation of the API Application
func (a *App) Initialize() error {
	var err error
	// 1) Initialize config settings & set environmental variables
	var conf *config.Configuration
	if os.Getenv("ENV") != "docker-dev" {
		conf, err = config.GetConfigurations()
		if err != nil {
			return err
		}
		conf.InitializeEnvironmentalVars()
	} else {
		conf, err = config.GetDevConfigurations()
		if err != nil {
			return err
		}
	}
	appLogger := utilities.NewAPILogger(conf)
	appLogger.InitLogger()
	appLogger.Info("Starting user server")
	appLogger.Infof(
		"LogLevel: %s, Environment: %s",
		conf.Logger.Level,
		conf.ENV,
	)
	// 2) Initialize & Connect DB Client
	a.db, err = database.InitializeNewClient()
	if err != nil {
		return err
	}
	err = a.db.Connect()
	if err != nil {
		return err
	}
	// 3) Initial DB Services
	gHandler := a.db.NewGroupHandler()
	uHandler := a.db.NewUserHandler()
	blHandler := a.db.NewBlacklistHandler()
	tHandler := a.db.NewTaskHandler()
	fHandler := a.db.NewFileHandler()
	gService := database.NewGroupService(a.db, gHandler)
	uService := database.NewUserService(a.db, uHandler, gHandler)
	bService := database.NewBlacklistService(a.db, blHandler)
	tService := services.NewTokenService(uService, gService, bService)
	ttService := database.NewTaskService(a.db, tHandler, uHandler, gHandler)
	fService := database.NewFileService(a.db, fHandler, uHandler, gHandler)
	// 4) Create RootAdmin user if database is empty
	var group models.Group
	var adminUser models.User
	group.Name = os.Getenv("ROOT_GROUP")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	docCount, err := a.db.GetCollection("groups").CountDocuments(ctx, bson.M{})
	if err != nil {
		return err
	}
	if docCount == 0 {
		group.RootAdmin = true
		group.Id = utilities.GenerateObjectID()
		if os.Getenv("ENV") == "test" {
			group.Id = "000000000000000000002222"
		}
		adminGroup, err := gService.GroupCreate(&group)
		if err != nil {
			return err
		}
		if os.Getenv("ENV") == "test" {
			adminUser.Id = "000000000000000000002221"
		}
		adminUser.Username = os.Getenv("ROOT_ADMIN")
		adminUser.Email = os.Getenv("ROOT_EMAIL")
		adminUser.Password = os.Getenv("ROOT_PASSWORD")
		adminUser.FirstName = "root"
		adminUser.LastName = "admin"
		adminUser.GroupId = adminGroup.Id
		_, err = uService.UserCreate(&adminUser)
		if err != nil {
			return err
		}
	}
	// 5) Initialize Server
	a.server = server.NewServer(appLogger, conf, uService, gService, ttService, fService, tService)
	return nil
}

// Run is a function used to run a previously initialized API Application
func (a *App) Run() {
	defer a.db.Close()
	a.server.Start()
}
