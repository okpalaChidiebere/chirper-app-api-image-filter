package api

import (
	"net/http"

	image_filterv1connect "github.com/okpalaChidiebere/chirper-app-gen-protos/image_filter/v1/image_filterv1connect"
)

type ServerConnectHandlers map[string]func() (path string, handlers http.Handler)

type Servers struct {
	ImageFilterServer  image_filterv1connect.ImagefilterServiceHandler
}


type APIServer struct {
	// router *mux.Router
	httpMux *http.ServeMux
}

func (a Servers) NewAPIServer(httpMux *http.ServeMux) *APIServer{
	server := &APIServer{ httpMux: httpMux, }
	return server
}

func (s Servers) GetAllServiceHandlers () ServerConnectHandlers {
	return map[string]func() (path string, handlers http.Handler){
		image_filterv1connect.ImagefilterServiceName:  func() (path string, handlers http.Handler) {
			return image_filterv1connect.NewImagefilterServiceHandler(s.ImageFilterServer)
		},
	}
}

func (s *APIServer) RegisterServiceHandlers(sh ServerConnectHandlers){
	for _, f := range sh {
		path, handler := f()
    	s.httpMux.Handle(path, handler)
	}
}