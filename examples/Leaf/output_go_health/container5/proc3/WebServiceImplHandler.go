// Blueprint: auto-generated by WebServer plugin
package proc3

import (
	"github.com/gorilla/mux"
	"net/http"
	"os"
	"context"
	"encoding/json"
	"errors"
	"log"
)

type WebServiceImplHandler struct {
	service *WebServiceImplHealthChecker
	url string
}
type WebServiceImpl_Leaf_WebResponse struct {
	Ret0 int64
}
type WebServiceImpl_Hello_WebResponse struct {
	Ret0 string
}
type WebServiceImpl_Health_WebResponse struct {
	Ret0 string
}
func NewWebServiceImplHandler(old_handler *WebServiceImplHealthChecker,framework string) *WebServiceImplHandler {
	handler := &WebServiceImplHandler{service: old_handler, url: ""}
	return handler
}

func (webhandler *WebServiceImplHandler) Leaf(w http.ResponseWriter, r *http.Request)  {
	var err error
	ctx := context.Background()
	var a int64
	arg1 := r.FormValue("a")
	if arg1 != "" {
		err = json.Unmarshal([]byte(arg1), &a)
		if err != nil {
			http.Error(w, err.Error(), 500)
			log.Println(err)
			log.Println(arg1, "a")
			return
		}
	}
	var jaegerTracer_trace_ctx string
	arg2 := r.FormValue("jaegerTracer_trace_ctx")
	if arg2 != "" {
		err = json.Unmarshal([]byte(arg2), &jaegerTracer_trace_ctx)
		if err != nil {
			http.Error(w, err.Error(), 500)
			log.Println(err)
			log.Println(arg2, "jaegerTracer_trace_ctx")
			return
		}
	}
	ret0, ret1 := webhandler.service.Leaf(ctx, a, jaegerTracer_trace_ctx)
	if ret1 != nil {
		http.Error(w, ret1.Error(), 500)
		return
	}
	response := WebServiceImpl_Leaf_WebResponse{}
	response.Ret0 = ret0
	json.NewEncoder(w).Encode(response)
	
}

func (webhandler *WebServiceImplHandler) Hello(w http.ResponseWriter, r *http.Request)  {
	var err error
	ctx := context.Background()
	var world string
	arg1 := r.FormValue("world")
	if arg1 != "" {
		err = json.Unmarshal([]byte(arg1), &world)
		if err != nil {
			http.Error(w, err.Error(), 500)
			log.Println(err)
			log.Println(arg1, "world")
			return
		}
	}
	var jaegerTracer_trace_ctx string
	arg2 := r.FormValue("jaegerTracer_trace_ctx")
	if arg2 != "" {
		err = json.Unmarshal([]byte(arg2), &jaegerTracer_trace_ctx)
		if err != nil {
			http.Error(w, err.Error(), 500)
			log.Println(err)
			log.Println(arg2, "jaegerTracer_trace_ctx")
			return
		}
	}
	ret0, ret1 := webhandler.service.Hello(ctx, world, jaegerTracer_trace_ctx)
	if ret1 != nil {
		http.Error(w, ret1.Error(), 500)
		return
	}
	response := WebServiceImpl_Hello_WebResponse{}
	response.Ret0 = ret0
	json.NewEncoder(w).Encode(response)
	
}

func (webhandler *WebServiceImplHandler) Health(w http.ResponseWriter, r *http.Request)  {
	ctx := context.Background()
	ret0, ret1 := webhandler.service.Health(ctx)
	if ret1 != nil {
		http.Error(w, ret1.Error(), 500)
		return
	}
	response := WebServiceImpl_Health_WebResponse{}
	response.Ret0 = ret0
	json.NewEncoder(w).Encode(response)
	
}

func (webhandler *WebServiceImplHandler) Run() error {
	addr := os.Getenv("webService_ADDRESS")
	port := os.Getenv("webService_PORT")
	if addr == "" || port == "" {
		return errors.New("Address or Port were not set")
	}
	url := "http://" + addr + ":" + port
	router := mux.NewRouter()
	webhandler.url = url
	router.Path("/Hello").HandlerFunc(webhandler.Hello)
	router.Path("/Health").HandlerFunc(webhandler.Health)
	router.Path("/Leaf").HandlerFunc(webhandler.Leaf)
	log.Println("Launching Server")
	return http.ListenAndServe(addr + ":" + port, router)
}

