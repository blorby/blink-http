package main

import (
	"github.com/blinkops/blink-sdk/plugin"
	"net/http"
	"os"
	"path"

	"github.com/blinkops/blink-http/implementation"
	blinkSdk "github.com/blinkops/blink-sdk"
	"github.com/blinkops/blink-sdk/plugin/config"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetLevel(log.DebugLevel)

	// Get the current directory.
	currentDirectory, err := os.Getwd()
	if err != nil {
		log.Error("Failed getting current directory: ", err)
		panic(err)
	}

	log.Info("Current directory is: ", currentDirectory)

	// Initialize the configuration.
	err = os.Setenv(config.ConfigurationPathEnvVar, path.Join(currentDirectory, "config.yaml"))
	if err != nil {
		log.Error("Failed to set configuration env variable: ", err)
		panic(err)
	}

	supportedActions := map[string]implementation.ActionHandler{
		"get":    executeHTTPGetAction,
		"post":   executeHTTPPostAction,
		"put":    executeHTTPPutAction,
		"delete": executeHTTPDeleteAction,
		"patch":  executeHTTPPatchAction,
	}

	httpPlugin, err := implementation.NewHTTPPlugin(currentDirectory, supportedActions)
	if err != nil {
		log.Error("Failed to create plugin implementation: ", err)
		panic(err)
	}

	err = blinkSdk.Start(httpPlugin)
	if err != nil {
		log.Fatal("Error during server startup: ", err)
	}
}


func executeHTTPGetAction(ctx *plugin.ActionContext, request *plugin.ExecuteActionRequest) ([]byte, error) {
	return implementation.ExecuteCoreHTTPAction(ctx, http.MethodGet, request)
}

func executeHTTPPostAction(ctx *plugin.ActionContext, request *plugin.ExecuteActionRequest) ([]byte, error) {
	return implementation.ExecuteCoreHTTPAction(ctx, http.MethodPost, request)
}

func executeHTTPPutAction(ctx *plugin.ActionContext, request *plugin.ExecuteActionRequest) ([]byte, error) {
	return implementation.ExecuteCoreHTTPAction(ctx, http.MethodPut, request)
}

func executeHTTPDeleteAction(ctx *plugin.ActionContext, request *plugin.ExecuteActionRequest) ([]byte, error) {
	return implementation.ExecuteCoreHTTPAction(ctx, http.MethodDelete, request)
}

func executeHTTPPatchAction(ctx *plugin.ActionContext, request *plugin.ExecuteActionRequest) ([]byte, error) {
	return implementation.ExecuteCoreHTTPAction(ctx, http.MethodPatch, request)
}

