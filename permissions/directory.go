package permissions

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
)

func LoadDirectory(dirname string) PermissionProvider {
	return PermissionProviderFunc{
		ProviderName: "directory",
		Permissions: func() ([]Permission, error) {
			files, err := ioutil.ReadDir(dirname)
			if err != nil {
				return nil, err
			}

			var allPermissions []Permission
			for _, file := range files {
				permissions, _, err := ReadPermissionsFromFile(filepath.Join(dirname, file.Name()))
				if err != nil {
					return nil, err
				}
				allPermissions = append(allPermissions, permissions...)
			}
			return allPermissions, nil
		},
		Groups: func() ([]Group, error) {
			files, err := ioutil.ReadDir(dirname)
			if err != nil {
				return nil, err
			}

			var allGroups []Group
			for _, file := range files {
				_, groups, err := ReadPermissionsFromFile(filepath.Join(dirname, file.Name()))
				if err != nil {
					return nil, err
				}
				allGroups = appendGroups(allGroups, groups)
			}
			return allGroups, nil
		}}
}

func ReadPermissionsFromFile(filename string) ([]Permission, []Group, error) {
	out, err := os.Open(filename)
	if err != nil {
		return nil, nil, errors.New("ReadPermissionsFromFile: " + err.Error())
	}
	defer out.Close()

	var data struct {
		Permissions []Permission
		Groups      []Group
	}
	err = json.NewDecoder(out).Decode(&data)
	if err != nil {
		return nil, nil, errors.New("read '" + filename + "' fail: " + err.Error())
	}
	return data.Permissions, data.Groups, nil
}
