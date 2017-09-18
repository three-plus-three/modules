package weaver

/*
import (
	"log"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/three-plus-three/modules/environment"

	"github.com/cheekybits/genny/generic"
	"github.com/three-plus-three/modules/environment/env_tests"
)

// WeaveType 用于泛型替换的类型
type TestWeaveType generic.Type

type testWeaver struct {
	group string
	value TestWeaveType
}

func (w *testWeaver) Update(group string, value TestWeaveType) error {
	w.group = group
	w.value = value
}

func (w *testWeaver) Generate() (TestWeaveType, error) {
	return w.value, nil
}

func TestServerSimple(t *testing.T) {
	env := env_tests.Clone()
	srv := NewServer(env, &testWeaver{}, log.New(os.Stderr, "[menus]", log.LstdFlags))
	hsrv := httptest.NewServer(srv)
	defer hsrv.Close()

	client := Connect(env, environment.ENV_AM_PROXY_ID, Callback(func() (interface{}, error) {
		return "aaabc", nil
	}), "apart", "abc", "/", log.New(os.Stderr, "[abc]", log.LstdFlags))

}
*/
