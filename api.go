package api

import (
	"github.com/labstack/echo"

	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
)

const (
	GET  = "GET"
	PUT  = "PUT"
	POST = "POST"
)

type Context interface {
	http.ResponseWriter
	Req() *http.Request
}

type context struct {
	http.ResponseWriter
	req *http.Request
}

func (c *context) Req() *http.Request {
	return c.req
}

type ServiceSpec struct {
	Name     string
	Path     string
	Methods  map[string]*MethodSpec
	Receiver reflect.Value
	Type     reflect.Type
}

type MethodSpec struct {
	Service   *ServiceSpec
	Path      string
	Verb      string
	Method    *reflect.Method
	ReqType   reflect.Type
	RespType  reflect.Type
	parseBody bool
	in        int
	out       int
}

var (
	errType     = reflect.TypeOf((*error)(nil)).Elem()
	contextType = reflect.TypeOf((*Context)(nil)).Elem()
)

func handleError(c Context, err error) {
	panic(err)
}

func writeJSONResponse(c Context, v interface{}) error {
	enc := json.NewEncoder(c)
	return enc.Encode(v)
}

func (ms *MethodSpec) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := &context{w, r}
	err := ms.runMethodInContext(c)
	if err != nil {
		handleError(c, err)
	}
}

func (ms *MethodSpec) runMethodInContext(c Context) error {
	v := reflect.ValueOf(c)

	var out []reflect.Value

	if ms.parseBody {
		reqValue := reflect.New(ms.ReqType)

		if c.Req().Body != nil {
			d := json.NewDecoder(c.Req().Body)
			err := d.Decode(reqValue.Interface())
			if err != nil && err != io.EOF {
				return err
			}
		}

		in := []reflect.Value{ms.Service.Receiver, v, reqValue}

		out = ms.Method.Func.Call(in)
	} else {

		in := []reflect.Value{ms.Service.Receiver, v}

		out = ms.Method.Func.Call(in)
	}

	// checked at service registration that the last return argument is an error

	if errif := out[ms.out-1].Interface(); errif != nil {
		return errif.(error)
	}

	if len(out) == 2 {
		if err := writeJSONResponse(c, out[0].Interface()); err != nil {
			return err
		}
	}

	return nil
}

func (s *ServiceSpec) Route(verb, path, methName string) {
	m, ok := s.Methods[methName]
	if !ok {
		panic(fmt.Errorf("Service: %s has no method named: %s", s.Name, methName))
	}

	m.Path = path
	m.Verb = verb
}

func (s *ServiceSpec) Get(path, methName string) {
	s.Route(GET, path, methName)
}

func (s *ServiceSpec) Post(path, methName string) {
	s.Route(POST, path, methName)
}

func (s *ServiceSpec) Put(path, methName string) {
	s.Route(PUT, path, methName)
}

func (s *ServiceSpec) Run(path string, e *echo.Echo) {
	router := e.Group(path)
	for _, m := range s.Methods {
		//log.Printf("%s %s%s %s [in: %d, reqT: %v, repT: %v]\n", m.Verb, path, m.Path, name, m.in, m.ReqType, m.RespType)
		switch m.Verb {
		case GET:
			router.Get(m.Path, m)
		case POST:
			router.Post(m.Path, m)
		case PUT:
			router.Put(m.Path, m)
		default:
			router.Get(m.Path, m)
		}
	}
}

func NewService(s interface{}) (*ServiceSpec, error) {

	receiver := reflect.ValueOf(s)
	receiverType := reflect.TypeOf(s)
	service := &ServiceSpec{Receiver: receiver, Type: receiverType}
	service.Name = reflect.Indirect(receiver).Type().Name()
	service.Path = service.Name

	service.Methods = make(map[string]*MethodSpec, receiverType.NumMethod())
	for i := 0; i < receiverType.NumMethod(); i++ {
		method := receiverType.Method(i)
		mType := method.Type
		in, out := mType.NumIn(), mType.NumOut()

		if out < 1 || out > 2 {
			return nil, fmt.Errorf("Invalid number of return arguments for api method in %s.%s: %d", service.Name, method.Name, out)
		}

		// check that last return type is error
		if mType.Out(out-1) != errType {
			return nil, fmt.Errorf("%s.%s should return an error as last return argument", service.Name, method.Name)
		}

		if in < 2 || in > 3 {
			return nil, fmt.Errorf("Invalid number of arguments for api method in %s.%s: %d", service.Name, method.Name, in)
		}

		if mType.In(1) != contextType {
			return nil, fmt.Errorf("%s.%s should have first argument of type Context.  Instead: %s", service.Name, method.Name, mType.In(1).Name())
		}

		ms := MethodSpec{
			Service: service,
			Path:    method.Name,
			Method:  &method,
			Verb:    GET,
			in:      in,
			out:     out,
		}

		if in > 2 {
			ms.ReqType = mType.In(2).Elem()
			ms.parseBody = true
		} else {
			ms.parseBody = false
		}

		if out > 1 {
			ms.RespType = mType.Out(0).Elem()
		}

		service.Methods[method.Name] = &ms
	}

	return service, nil
}
