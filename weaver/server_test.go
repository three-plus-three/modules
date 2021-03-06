package weaver

import (
	"fmt"
	"net"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/runner-mei/log"
	"github.com/three-plus-three/modules/environment"
	"github.com/three-plus-three/modules/environment/env_tests"
)

// WeaveType 用于泛型替换的类型
// type TestWeaveType generic.Type

type testWeaver struct {
	group string
	value WeaveType
}

func (w *testWeaver) Stats() interface{} {
	return nil
}

func (w *testWeaver) Update(group string, value WeaveType) error {
	w.group = group
	w.value = value
	return nil
}

func (w *testWeaver) Generate(ctx string) (WeaveType, error) {
	return w.value, nil
}

func TestServerSimple(t *testing.T) {
	env := env_tests.Clone(nil)

	srv, err := NewServer(env, &testWeaver{}, log.Empty(), nil)
	if err != nil {
		t.Error(err)
		return
	}
	//srv.Close()

	hsrv := httptest.NewServer(srv)
	defer hsrv.Close()
	_, port, _ := net.SplitHostPort(strings.TrimPrefix(hsrv.URL, "http://"))
	env.GetServiceConfig(environment.ENV_HOME_PROXY_ID).SetPort(port)

	client := Connect(env, environment.ENV_AM_PROXY_ID, Callback(func() (ValueType, error) {
		return 12, nil
	}), "apart", "abc", "/", log.Empty())

	defer client.Close()

	time.Sleep(2 * time.Second)
	a, err := client.Read()
	if err != nil {
		t.Error(err)
		return
	}

	if fmt.Sprint(a) != "12" {
		t.Errorf("%T %s", a, a)
	}

}
