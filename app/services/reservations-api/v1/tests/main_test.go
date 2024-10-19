package tests

import (
	"fmt"
	"testing"

	"github.com/ameghdadian/service/business/data/dbtest"
	"github.com/ameghdadian/service/business/data/redistest"
	"github.com/ameghdadian/service/foundation/docker"
)

var c *docker.Container
var rc *docker.Container

func TestMain(m *testing.M) {
	var err error
	c, err = dbtest.StartDB()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer dbtest.StopDB(c)

	rc, err = redistest.StartRedis()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer redistest.StopRedis(rc)
	m.Run()
}
