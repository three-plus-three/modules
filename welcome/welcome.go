package welcome

import (
	"bytes"
	"encoding/json"
	"errors"
	"html/template"
	"io/ioutil"

	"github.com/three-plus-three/modules/environment"
	"github.com/three-plus-three/modules/urlutil"
)

const FieldName = "welcome_url"

type Config struct {
	Name        string `json:"name"`
	ListURL     string `json:"list_url"`
	RedirectURL string `json:"redirect_url"`
}

func ReadWelcomeConfigs(env *environment.Environment) ([]Config, error) {
	filename := env.Fs.FromConfig("home.json")
	args := map[string]interface{}{
		"urlRoot": env.DaemonUrlPath,
	}

	bs, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, errors.New("ReadHTTPConfigFromFile: " + err.Error())
	}

	t, err := template.New("default").Funcs(template.FuncMap{
		"join": urlutil.Join,
	}).Parse(string(bs))
	if err != nil {
		return nil, errors.New("parse url template in '" + filename + "' fail: " + err.Error())
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, args); err != nil {
		return nil, errors.New("generate url template in '" + filename + "' fail: " + err.Error())
	}
	if buf.Len() == 0 {
		return nil, errors.New("template result in '" + filename + "' is empty.")
	}

	var config struct {
		Applications []Config `json:"applications,omitempty"`
	}

	bs = buf.Bytes()
	err = json.NewDecoder(&buf).Decode(&config)
	if err != nil {
		return nil, errors.New("read '" + filename + "' fail: " + err.Error() + "\r\n" + string(bs))
	}
	return config.Applications, nil
}
