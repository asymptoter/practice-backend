/*Package docker wraps common used docker utilities*/
package docker

import (
	"fmt"

	"github.com/asymptoter/practice-backend/base/db"
	"github.com/asymptoter/practice-backend/base/mongo"
	"github.com/asymptoter/practice-backend/base/redis"

	"github.com/jmoiron/sqlx"
	"github.com/ory/dockertest"
	dc "github.com/ory/dockertest/docker"
	log "github.com/sirupsen/logrus"
	"gopkg.in/mgo.v2"
)

var Pool *dockertest.Pool
var (
	resources = map[string]*dockertest.Resource{}
	envs      = map[string][]string{
		"postgres": []string{"POSTGRES_USER=a", "POSTGRES_PASSWORD=b", "POSTGRES_DB=practice"},
		"redis":    []string{},
		"mongo":    []string{"MONGO_INITDB_ROOT_USERNAME=a", "MONGO_INITDB_ROOT_PASSWORD=b"},
	}
)

func GetRedis() redis.Service {
	initPool()

	addr := fmt.Sprintf("localhost:%s", getResource("redis").GetPort("6379/tcp"))
	return redis.NewService(addr)
}

func GetPostgreSQL() *sqlx.DB {
	initPool()
	s := fmt.Sprintf("postgres://a:b@localhost:%s/practice?sslmode=disable", getResource("postgres").GetPort("5432/tcp"))
	return db.MustNew("postgres", s)
}

func GetMongoDB() *mgo.Session {
	initPool()

	s := fmt.Sprintf("mongodb://localhost:%s/practice", getResource("mongo").GetPort("27017/tcp"))
	return mongo.MustNew(s)
}

func getResource(containerName string) *dockertest.Resource {
	var err error
	var res *dockertest.Resource
	cID := getContainerID(containerName)
	if len(cID) != 0 {
		res = &dockertest.Resource{}
		container, err := Pool.Client.InspectContainer(cID)
		if err != nil {
			log.Fatalf("Could not get containers: %s", err)
		}
		res.Container = container
	} else {
		res, err = Pool.RunWithOptions(&dockertest.RunOptions{Name: "unittest_" + containerName, Repository: containerName, Tag: "latest", Env: envs[containerName]})
		if err != nil {
			log.Fatalf("Could not start resource: %s", err)
		}
	}
	return res
}

func Purge(containerName string) {
	if err := Pool.Purge(resources[containerName]); err != nil {
		log.Fatalf("Purge %s failed", containerName)
	}
}

// TODO: deprecate
func PurgePostgreSQL() {
	if err := Pool.Purge(resources["postgres"]); err != nil {
		log.Fatalf("Purge postgresql failed")
	}
}

// TODO: deprecate
func PurgeRedis() {
	if err := Pool.Purge(resources["redis"]); err != nil {
		log.Error("Purge redis failed")
	}
}

// TODO: deprecate
func PurgeMongo() {
	if err := Pool.Purge(resources["mongo"]); err != nil {
		log.Error("Purge redis failed")
	}
}

func initPool() {
	if Pool == nil {
		var err error
		Pool, err = dockertest.NewPool("")
		if err != nil {
			log.Fatalf("Could not connect to docker: %s", err)
		}
	}
}

func getContainerID(imageName string) string {
	cs, err := Pool.Client.ListContainers(dc.ListContainersOptions{})
	if err != nil {
		log.Fatalf("Could not list containers: %s", err)
	}

	name := "/unittest_" + imageName
	for _, c := range cs {
		if c.Names[0] == name {
			return c.ID
		}
	}
	return ""
}
