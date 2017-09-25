package spi

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"text/template"

	"github.com/three-plus-three/modules/errors"
	"github.com/three-plus-three/modules/permissions"
	"github.com/three-plus-three/modules/urlutil"
)

var (
	mu        sync.Mutex
	privoders = map[string]permissions.PermissionProvider{}
)

func Read() (*permissions.PermissionData, error) {
	mu.Lock()
	defer mu.Unlock()

	switch len(privoders) {
	case 0:
		return nil, nil
	case 1:
		for _, p := range privoders {
			return p.Read()
		}
		fallthrough
	default:
		var all = &permissions.PermissionData{}
		for _, p := range privoders {
			data, err := p.Read()
			if err != nil {
				return nil, err
			}
			permissions.MergePermissionData(all, data)
		}
		return all, nil
	}
}

func ClearRegisters() {
	mu.Lock()
	defer mu.Unlock()
	privoders = map[string]permissions.PermissionProvider{}
}

func Register(name string, privoder permissions.PermissionProvider) {
	mu.Lock()
	defer mu.Unlock()

	if privoder == nil {
		panic("provider is nil")
	}
	if _, ok := privoders[name]; ok {
		panic("privoder '" + name + "' is already exists.")
	}
	privoders[name] = privoder
}

type HTTPConfig struct {
	file string
	Name string `json:"name"`
	URL  string `json:"url"`
}

func LoadHTTP(dirname string, args map[string]interface{}) error {
	files, err := ioutil.ReadDir(dirname)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return errors.Wrap(err, "载入 PermissionProvider 失败")
	}

	var configs []HTTPConfig

	for _, file := range files {
		config, err := readHTTPConfigFromFile(filepath.Join(dirname, file.Name()), args)
		if err != nil {
			return errors.Wrap(err, "载入 PermissionProvider 失败")
		}
		config.file = file.Name()
		configs = append(configs, *config)
	}

	for _, config := range configs {
		Register("directory:"+config.file+":"+config.Name,
			permissions.PermissionProviderFunc(func() (*permissions.PermissionData, error) {
				data, err := readPermissionsFromHTTP(config.Name, config.URL)
				if err != nil {
					return nil, errors.Wrap(err, "从 HTTP 载入 Permissions 失败")
				}
				return data, nil
			}))
	}
	return nil
}

func readHTTPConfigFromFile(filename string, args map[string]interface{}) (*HTTPConfig, error) {
	out, err := os.Open(filename)
	if err != nil {
		return nil, errors.New("ReadHTTPConfigFromFile: " + err.Error())
	}
	defer out.Close()

	var config HTTPConfig
	err = json.NewDecoder(out).Decode(&config)
	if err != nil {
		return nil, errors.New("read '" + filename + "' fail: " + err.Error())
	}

	t, err := template.New("default").Funcs(template.FuncMap{
		"join": urlutil.Join,
	}).Parse(config.URL)
	if err != nil {
		return nil, errors.New("parse url template in '" + filename + "' fail: " + err.Error())
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, args); err != nil {
		return nil, errors.New("generate url in '" + filename + "' fail: " + err.Error())
	}

	config.Name = filename + ":" + config.Name
	config.URL = buf.String()
	return &config, nil
}

func readPermissionsFromHTTP(filename, url string) (*permissions.PermissionData, error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, errors.New("read web in '" + filename + "' fail: " + err.Error())
	}
	if response.StatusCode != http.StatusOK {
		bs, _ := ioutil.ReadAll(response.Body)
		if len(bs) == 0 {
			return nil, errors.New("read web in '" + filename + "' fail: " + response.Status)
		}
		return nil, errors.New("read web in '" + filename + "' fail: " + response.Status + "\r\n" + string(bs))
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, response.Body)
	if err != nil {
		return nil, errors.New("read web in '" + filename + "' fail: " + err.Error())
	}

	var data = &permissions.PermissionData{}
	err = json.NewDecoder(&buf).Decode(data)
	if err != nil {
		return nil, errors.New("read web in '" + filename + "' fail: " + err.Error() + "\r\n" + buf.String())
	}
	return data, nil
}
