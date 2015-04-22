package api

import (
	"bytes"
	"encoding/json"
	"github.com/labstack/echo"
	"net/http"
	"net/http/httptest"
	"testing"
)

type JobsAPI struct{}

type Job struct {
	Name string
}

type JobResp struct {
	Jobs []Job
}

type JobReq struct {
	Limit int
}

func (JobsAPI) List(c Context, r *JobReq) (*JobResp, error) {
	jobs := []Job{Job{"One"}, Job{"Two"}}
	return &JobResp{jobs}, nil
}

func TestServer(t *testing.T) {
	e := echo.New()
	s, err := NewService(&JobsAPI{})

	if err != nil {
		t.Error(err)
	}

	s.Get("/", "List")
	s.Run("/jobs", e)

	resp := httptest.NewRecorder()

	j := JobReq{1}

	b, err := json.Marshal(&j)
	if err != nil {
		t.Error(err)
	}

	data := bytes.NewReader(b)

	req, err := http.NewRequest("GET", "/jobs/", data)

	if err != nil {
		t.Error(err)
	}

	e.ServeHTTP(resp, req)

	if resp.Code != 200 {
		t.Errorf("Response should generate status OK but got: %d", resp.Code)
	}

	if resp.Body == nil {
		t.Error("response body was nil")
	}

	jResp := new(JobResp)

	if err := json.NewDecoder(resp.Body).Decode(jResp); err != nil {
		t.Errorf("Response should reply with jobs. Error when decoding JSON response body: %s", err.Error())
	}

	if len(jResp.Jobs) != 2 {
		t.Error("Reponse does not match expected")
	}

	//fmt.Println(resp)
	//fmt.Printf("%+v", resp)
}

type JobsAPIFail struct{}

func (JobsAPIFail) ListFail(c Context, r *JobReq) {

}

func TestServerFail(t *testing.T) {
	_, err := NewService(&JobsAPIFail{})
	if err == nil {
		t.Error("Should not let method without any return values")
	}
}

type JobsAPIFail2 struct{}

func (JobsAPIFail2) ListFail(c Context, r *JobReq) *JobResp {
	jobs := []Job{Job{"One"}, Job{"Two"}}
	return &JobResp{jobs}
}

func TestServiceIncorrectReturnSignature(t *testing.T) {
	_, err := NewService(&JobsAPIFail2{})
	if err == nil {
		t.Error("Should not let method without return error")
	}
}

type JobsAPIFail3 struct{}

func (JobsAPIFail3) ListFail(r *JobReq) error {
	return nil
}

func TestServiceIncorrectMethodArgs(t *testing.T) {
	_, err := NewService(&JobsAPIFail3{})
	if err == nil {
		t.Error("Should not allow service method without context type argument")
	}
}
