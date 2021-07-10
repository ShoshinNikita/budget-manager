package main

import (
	"log"

	"github.com/ShoshinNikita/budget-manager/internal/app"
)

//nolint:gochecknoglobals
var (
	// version is a version of the app. It must be set during the build process with -ldflags flag
	version = "unknown"
	// gitHash is the last commit hash. It must be set during the build process with -ldflags flag
	gitHash = "unknown"
)

// Swagger General Info
//
//nolint:lll
//
// @title Budget Manager API
// @version v0.2
// @description Easy-to-use, lightweight and self-hosted solution to track your finances - [GitHub](https://github.com/ShoshinNikita/budget-manager)
//
// @BasePath /api
//
// @securityDefinitions.basic BasicAuth
//
// @license.name MIT
// @license.url https://github.com/ShoshinNikita/budget-manager/blob/master/LICENSE
//

func main() {
	cfg, err := app.ParseConfig()
	if err != nil {
		log.Fatalln(err)
	}

	// Create a new application
	app := app.NewApp(cfg, version, gitHash)

	// Prepare the application
	if err := app.PrepareComponents(); err != nil {
		log.Fatalln(err)
	}

	// Run the application
	if err := app.Run(); err != nil {
		log.Fatalln(err)
	}
}
