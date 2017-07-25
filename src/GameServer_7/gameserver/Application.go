package gameserver

import (
	"github.com/pkg/errors"
	"log"
)

var application *Application = nil

type Application struct {
	staticInfo *StaticInfo
	server     *Server
}

////////////////////////////////////////////////////////////////////////////////////////////

func MakeApp() error {
	if application == nil {
		// Static info
		staticInfo, err := NewStaticInfo()
		if err != nil {
			log.Printf("Failed create static info: %s\n", err)
			return err
		}

		// Server
		server := NewServer()

		application = &Application{
			staticInfo: staticInfo,
			server:     server,
		}
		return nil
	}
	return errors.New("Already have application")
}

func GetApp() *Application {
	return application
}

////////////////////////////////////////////////////////////////////////////////////////////

func (app *Application) RunServer() error {
	return app.server.StartListen()
}

func (app *Application) ExitServer() error {
	return app.server.ExitServer()
}

func (app *Application) GetStaticInfo() *StaticInfo {
	return app.staticInfo
}
