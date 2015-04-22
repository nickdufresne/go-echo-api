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

type jobStore struct {
	lock sync.RWMutex
	id   int
	jobs []*Job
}

func (s *jobStore) saveJob(j *Job) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.id++
	j.ID = s.id
	jobCopy := *j
	s.jobs = append(s.jobs, &jobCopy)
}

func (s *jobStore) getJobs() []Job {
	s.lock.RLock()
	defer s.lock.RUnlock()
	result := make([]Job, len(s.jobs))
	for i, j := range s.jobs {
		result[i] = *j
	}
	return result
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
