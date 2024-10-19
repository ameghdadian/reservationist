package redistest

import (
	"fmt"

	"github.com/ameghdadian/service/foundation/docker"
)

func StartRedis() (*docker.Container, error) {
	image := "redis:7.4.0"
	port := "6379"

	rc, err := docker.StartContainer(image, port, []string{}, []string{})
	if err != nil {
		return nil, fmt.Errorf("starting container: %w", err)
	}

	fmt.Printf("Image:			%s\n", image)
	fmt.Printf("ContainerID:		%s\n", rc.ID)
	fmt.Printf("Host:			%s\n", rc.Host)

	return rc, nil
}

func StopRedis(rc *docker.Container) {
	docker.StopContainer(rc.ID)
	fmt.Println("Stopped:", rc.ID)
}
