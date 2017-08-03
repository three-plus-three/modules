package permissions

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"text/template"

	"github.com/three-plus-three/modules/urlutil"
)

type HTTPConfig struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

func LoadHTTP(dirname string, args map[string]interface{}) (PermissionProvider, error) {
	files, err := ioutil.ReadDir(dirname)
	if err != nil {
		return nil, err
	}

	var configs []HTTPConfig

	for _, file := range files {
		config, err := ReadHTTPConfig(filepath.Join(dirname, file.Name()), args)
		if err != nil {
			return nil, err
		}
		configs = append(configs, *config)
	}

	return PermissionGetFunc(func() ([]Permission, error) {
		var allPermissions []Permission
		for _, config := range configs {
			permissions, err := ReadHTTP(config.Name, config.URL)
			if err != nil {
				return nil, err
			}
			allPermissions = append(allPermissions, permissions...)
		}
		return allPermissions, nil
	}), nil
}

func ReadHTTPConfig(filename string, args map[string]interface{}) (*HTTPConfig, error) {
	out, err := os.Open(filename)
	if err != nil {
		return nil, errors.New("ReadFile: " + err.Error())
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

func ReadHTTP(filename, url string) ([]Permission, error) {
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

	var permissions []Permission
	err = json.NewDecoder(&buf).Decode(&permissions)
	if err != nil {
		return nil, errors.New("read web in '" + filename + "' fail: " + err.Error() + "\r\n" + buf.String())
	}
	return permissions, nil
}