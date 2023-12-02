package router

import (
	"fmt"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type HttpRequestContext struct {
	Request  *http.Request
	Params   httprouter.Params
	Metadata map[string]*string
}

type Handler func(ctx *HttpRequestContext) HandlerCallback

type HandlerCallback func(w http.ResponseWriter)

type Endpoint struct {
	Method string
	Path   string
}
type HttpInstance struct {
	Handlers map[Endpoint]Handler
	Filters  []Handler
	Port     int
	router   *httprouter.Router
}

func NewHttpInstance() *HttpInstance {
	router := httprouter.New()
	router.ServeFiles("/dist/*filepath", http.Dir("./dist/"))
	instance := &HttpInstance{map[Endpoint]Handler{}, []Handler{}, 8080, router}

	return instance
}

func (h *HttpInstance) WithPort(port int) *HttpInstance {
	h.Port = port
	return h
}
func (h *HttpInstance) BlockingServe() {
	h.setup()
	log.Printf("Starting server at  %d", h.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", h.Port), h.router))
}
func (h *HttpInstance) RegisterHandler(method string, path string, handler Handler) {
	h.Handlers[Endpoint{method, path}] = handler
}

func (h *HttpInstance) RegisterFilter(handler Handler) {
	h.Filters = append(h.Filters, handler)
}

func (h *HttpInstance) setup() {
	for endpoint, handler := range h.Handlers {
		endpoint := endpoint
		handler := handler
		var foo httprouter.Handle = func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
			ctx := &HttpRequestContext{
				Request:  r,
				Params:   p,
				Metadata: map[string]*string{}}

			for _, f := range h.Filters {
				r := f(ctx)
				if r != nil {
					r(w)
					return
				}
			}
			handler(ctx)(w)
		}
		register(h.router, endpoint, foo)
	}
}

func register(router *httprouter.Router, ep Endpoint, handler httprouter.Handle) {
	router.Handle(ep.Method, ep.Path, handler)
}
