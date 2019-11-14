package spi

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/three-plus-three/modules/errors"
	"github.com/three-plus-three/modules/permissions"
)

func LoadDirectory(dirname string) error {
	Register("directory", permissions.PermissionProviderFunc(func() (*permissions.PermissionData, error) {
		files, err := ioutil.ReadDir(dirname)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, nil
			}
			return nil, errors.Wrap(err, "从 File 载入 Permissions 失败")
		}

		switch len(files) {
		case 0:
			return nil, nil
		case 1:
			return readPermissionsFromFile(filepath.Join(dirname, files[0].Name()))
		default:
			all := &permissions.PermissionData{}
			for _, file := range files {
				data, err := readPermissionsFromFile(filepath.Join(dirname, file.Name()))
				if err != nil {
					return nil, errors.Wrap(err, "从 File 载入 Permissions 失败")
				}
				permissions.MergePermissionData(all, data)
			}
			return all, nil
		}
	}))
	return nil
}

func readPermissionsFromFile(filename string) (*permissions.PermissionData, error) {
	out, err := os.Open(filename)
	if err != nil {
		return nil, errors.New("ReadPermissionsFromFile: " + err.Error())
	}
	defer out.Close()

	var data = &permissions.PermissionData{}
	err = json.NewDecoder(out).Decode(data)
	if err != nil {
		return nil, errors.New("read web in '" + filename + "' fail: " + err.Error())
	}
	return data, nil
}
