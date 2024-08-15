package startup

import (
	"log"

	"github.com/kardianos/service"
)

// Program
type Program struct{}

// Program.Start
func (p *Program) Start(s service.Service) error {
	go Startup()
	return nil
}

// Program.Stop
func (p *Program) Stop(s service.Service) error {
	return nil
}

// Run
func Run(command *string) {
	s, err := service.New(&Program{}, &service.Config{
		Name:        "Tunnel Proxy Service",
		DisplayName: "Tunnel Proxy Service",
		Description: "Tunnel Proxy Service",
	})
	if err != nil {
		log.Println(err)
	}

	if *command != "" {
		err = service.Control(s, *command)
		if err != nil {
			log.Println(err)
		}
		return
	}

	if err = s.Run(); err != nil {
		log.Println(err)
	}
}
