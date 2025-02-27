package postgres

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

func ConfigureTestDB(ctx context.Context) (DB, func(), error) {
	var db DB
	pool, err := dockertest.NewPool("")
	if err != nil {
		return db, func() {}, fmt.Errorf("could not construct pool: %w", err)
	}

	err = pool.Client.Ping()
	if err != nil {
		return db, func() {}, fmt.Errorf("could not connect to Docker: %w", err)
	}

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "latest",
		Env: []string{
			"POSTGRES_USER=test",
			"POSTGRES_PASSWORD=test",
			"POSTGRES_DB=test",
		},
		ExposedPorts: []string{"5432/tcp"},
	}, func(hc *docker.HostConfig) {
		hc.AutoRemove = true
		hc.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})

	if err != nil {
		return db, func() {}, fmt.Errorf("could not start resource: %w", err)
	}

	hostPort := resource.GetPort("5432/tcp")
	connStr := "postgres://test:test@localhost:" + hostPort + "/test?sslmode=disable"

	resource.Expire(240) // hard kill after 240s

	pool.MaxWait = 120 * time.Second
	if err = pool.Retry(func() error {
		db, err = Connect(connStr)
		if err != nil {
			return err
		}
		return db.Ping(ctx)
	}); err != nil {
		log.Fatalf("could not connect to docker: %s", err)
	}

	cleanup := func() {
		if err := pool.Purge(resource); err != nil {
			log.Fatalf("Could not purge resource: %s", err)
		}
	}

	if err := db.SetupSchema(ctx); err != nil {
		return db, cleanup, fmt.Errorf("could not setup schema: %w", err)
	}

	return db, cleanup, nil
}
