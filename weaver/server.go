package weaver

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/cheekybits/genny/generic"
	"github.com/runner-mei/log"
	"github.com/three-plus-three/modules/environment"
)

// WeaveType 用于泛型替换的类型
type WeaveType generic.Type

// Weaver 菜单的组织工具
type Weaver interface {
	Update(group string, value WeaveType) error
	Generate(ctx string) (WeaveType, error)
	Stats() interface{}
}

// Server 菜单的服备
type Server struct {
	env        *environment.Environment
	weaver     Weaver
	renderHTML func(w http.ResponseWriter, r *http.Request, data WeaveType)
	logger     log.Logger
}

func (srv *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer io.Copy(ioutil.Discard, r.Body)
	}

	switch r.Method {
	case "GET":
		if strings.HasSuffix(r.URL.Path, "/stats") ||
			strings.HasSuffix(r.URL.Path, "/stats/") {
			srv.stats(w, r)
			return
		}
		srv.read(w, r)
	case "PUT", "POST":
		srv.write(w, r)
	default:
		http.NotFound(w, r)
	}
}

func isConsumeHTML(r *http.Request) bool {
	contentType := r.Header.Get("Content-Type")
	if strings.Contains(contentType, "text/html") {
		return true
	}
	accept := r.Header.Get("Accept")
	return strings.Contains(accept, "text/html")
}

func renderTEXT(w http.ResponseWriter, code int, txt string) error {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	_, err := fmt.Fprintln(w, txt)
	return err
}

func renderJSON(w http.ResponseWriter, code int, value interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	return json.NewEncoder(w).Encode(value)
}

func (srv *Server) stats(w http.ResponseWriter, r *http.Request) {
	err := renderJSON(w, http.StatusOK, srv.weaver.Stats())
	if err != nil {
		srv.logger.Error("stats fail", log.Error(err))
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	}
}

func (srv *Server) read(w http.ResponseWriter, r *http.Request) {
	ctx := r.URL.Query().Get("app")
	if ctx == "stats" {
		err := renderJSON(w, http.StatusOK, srv.weaver.Stats())
		if err != nil {
			srv.logger.Error("stats fail", log.String("app", ctx), log.Error(err))
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}
		return
	}

	results, err := srv.weaver.Generate(ctx)
	if err != nil {
		srv.logger.Error("stats fail", log.String("app", ctx), log.Error(err))
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	if srv.renderHTML != nil && isConsumeHTML(r) {
		srv.renderHTML(w, r, results)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(results)
	if err != nil {
		srv.logger.Error("stats fail", log.String("app", ctx), log.Error(err))
	} else {
		srv.logger.Info("query is ok", log.String("app", ctx))
	}
}

func (srv *Server) write(w http.ResponseWriter, r *http.Request) {
	group := r.URL.Query().Get("app")
	if group == "" {
		srv.logger.Error("app is missing")
		http.Error(w, "app is missing", http.StatusBadRequest)
		return
	}

	var data WeaveType
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		srv.logger.Error("update fail", log.String("group", group), log.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = srv.weaver.Update(group, data)
	if err != nil {
		srv.logger.Error("update fail", log.String("group", group), log.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	renderTEXT(w, http.StatusOK, "OK")

	srv.logger.Info("update is successful", log.String("group", group))
}

// NewServer 创建一个菜单服备
func NewServer(env *environment.Environment, weaver Weaver, logger log.Logger,
	renderHTML func(w http.ResponseWriter, r *http.Request, data WeaveType)) (*Server, error) {
	return &Server{
		env:        env,
		weaver:     weaver,
		renderHTML: renderHTML,
		logger:     logger,
	}, nil
}
