package tests

import (
	"fmt"
	"testing"

	"github.com/ameghdadian/service/business/data/dbtest"
	"github.com/ameghdadian/service/foundation/docker"
)

var c *docker.Container

func Test_Main(m *testing.M) {
	var err error
	c, err = dbtest.StartDB()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer dbtest.StopDB(c)

	m.Run()
}
