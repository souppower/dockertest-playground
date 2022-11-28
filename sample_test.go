package dockertest_playground

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/ory/dockertest/v3"
	"log"
	"os"
	"testing"
)

var db *redis.Client

func TestMain(m *testing.M) {
	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	// pulls an image, creates a container based on it and runs it
	resource, err := pool.Run("redis", "latest", nil)
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	if err = pool.Retry(func() error {
		db = redis.NewClient(&redis.Options{
			Addr: fmt.Sprintf("localhost:%s", resource.GetPort("6379/tcp")),
		})
		return db.Ping(context.Background()).Err()
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	code := m.Run()

	// When you're done, kill and remove the container, you can't defer this because os.Exit doesn't care for defer
	if err = pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

func TestSomething(t *testing.T) {
	db.Set(context.Background(), "k0", "v0", 0)
	val, err := db.Get(context.Background(), "k0").Result()
	log.Println(val, err)
}
