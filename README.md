# go-echo-api

experimental Golang API service to help manage rest resource API in echo http server

from the examples folder:

Built using:

[Echo](http://github.com/labstack/echo)

To get started with this project
=================================

```
go get github.com/labstack/echo
go get github.com/labstack/echo/middleware
go get github.com/nickdufresne/go-echo-api
```

From the examples directory:

```go

// in file: jobs.go

package main

import (
  "github.com/labstack/echo"
  mw "github.com/labstack/echo/middleware"
  "github.com/nickdufresne/go-echo-api"

  "log"
  "sync"
)

type JobsAPI struct{}

type Job struct {
  ID   int    `json:"id"`
  Name string `json:"name"`
}

type JobsList struct {
  Jobs []Job
}

func (JobsAPI) Create(c api.Context, job *Job) (*Job, error) {
  store.saveJob(job)
  return job, nil
}

func (JobsAPI) List(c api.Context) (*JobsList, error) {
  jobs := store.getJobs()
  return &JobsList{jobs}, nil
}

var store = &jobStore{}

func main() {
  e := echo.New()
  e.Use(mw.Logger)
  s, err := api.NewService(&JobsAPI{})
  if err != nil {
    // panic if method signatures aren't correct, because the api won't work as expected
    panic(err)
  }

  // set up default rest urls ...
  s.Get("/", "List")
  s.Post("/", "Create")
  s.Run("/jobs", e)

  log.Println("Launching jobs api server on http://localhost:4000")
  e.Run(":4000")
}
```

Then run:

```
go run jobs.go # => Launching jobs api server on http://localhost:4000

curl -X POST http://localhost:4000/jobs/ -d '{"name": "hello"}'

# => {"id":1,"name":"hello"}

curl -X POST http://localhost:4000/jobs/ -d '{"name": "world"}'

curl -X GET http://localhost:4000/jobs/

# => {"Jobs":[{"id":1,"name":"hello"},{"id":2,"name":"world"}]}
```

