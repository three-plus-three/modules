package permissions

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
)

func LoadDirectory(dirname string) PermissionProvider {
	return PermissionGetFunc(func() ([]Permission, error) {
		files, err := ioutil.ReadDir(dirname)
		if err != nil {
			return nil, err
		}

		var allPermissions []Permission
		for _, file := range files {
			permissions, err := ReadFile(filepath.Join(dirname, file.Name()))
			if err != nil {
				return nil, err
			}
			allPermissions = append(allPermissions, permissions...)
		}
		return allPermissions, nil
	})
}

func ReadFile(filename string) ([]Permission, error) {
	out, err := os.Open(filename)
	if err != nil {
		return nil, errors.New("ReadFile: " + err.Error())
	}
	defer out.Close()

	var permissions []Permission
	err = json.NewDecoder(out).Decode(&permissions)
	if err != nil {
		return nil, errors.New("read '" + filename + "' fail: " + err.Error())
	}
	return permissions, nil
}
