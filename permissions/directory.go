package permissions

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/three-plus-three/modules/errors"
)

func LoadDirectory(dirname string) error {
	RegisterPermissions("directory", PermissionProviderFunc(func() (*PermissionData, error) {
		files, err := ioutil.ReadDir(dirname)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, nil
			}
			return nil, errors.Wrap(err, "从 File 载入 Permissions 失败")
		}

		all := &PermissionData{}
		for _, file := range files {
			data, err := ReadPermissionsFromFile(filepath.Join(dirname, file.Name()))
			if err != nil {
				return nil, errors.Wrap(err, "从 File 载入 Permissions 失败")
			}
			appendPermissionData(all, data)
		}
		return all, nil
	}))
	return nil
}

func ReadPermissionsFromFile(filename string) (*PermissionData, error) {
	out, err := os.Open(filename)
	if err != nil {
		return nil, errors.New("ReadPermissionsFromFile: " + err.Error())
	}
	defer out.Close()

	var data = &PermissionData{}
	err = json.NewDecoder(out).Decode(data)
	if err != nil {
		return nil, errors.New("read web in '" + filename + "' fail: " + err.Error())
	}
	return data, nil
}
