package menus

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/three-plus-three/modules/environment"
	"github.com/three-plus-three/modules/toolbox"
)

// Weaver 菜单的组织工具
type Weaver interface {
	Update(group string, menuList interface{}) error
	Generate() (interface{}, error)
}

// Server 菜单的服备
type Server struct {
	env    *environment.Environment
	weaver Weaver
}

func (srv *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer io.Copy(ioutil.Discard, r.Body)
	}

	switch r.Method {
	case "GET":
		srv.read(w, r)
	case "PUT", "POST":
		srv.write(w, r)
	default:
		http.NotFound(w, r)
	}
}

func text(w http.ResponseWriter, code int, txt string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	fmt.Fprintln(w, txt)
}

func (srv *Server) read(w http.ResponseWriter, r *http.Request) {
	// o := srv.value.Load()
	// if o == nil {
	// 	http.Error(w, "server is initialing.", http.StatusServiceUnavailable)
	// 	return
	// }
	// weaver, ok := o.(*menuWeaver)
	// if !ok || weaver == nil {
	// 	http.Error(w, "weaver is initialing.", http.StatusServiceUnavailable)
	// 	return
	// }

	results, err := srv.weaver.Generate()
	if err != nil {
		http.Error(w, "weaver is initialing.", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(results)
	if err != nil {
		log.Println("[menus]", err)
	}
}

func (srv *Server) write(w http.ResponseWriter, r *http.Request) {
	group := r.URL.Query().Get("app")
	if group == "" {
		http.Error(w, "app is missing", http.StatusBadRequest)
		return
	}

	var data []toolbox.Menu
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = srv.weaver.Update(group, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	text(w, http.StatusOK, "OK")
}

// NewServer 创建一个菜单服备
func NewServer(env *environment.Environment, weaver Weaver) (*Server, error) {
	return &Server{
		env:    env,
		weaver: weaver,
	}, nil
}
