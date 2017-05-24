package types

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/three-plus-three/modules/environment"
)

var modelsFile *flag.Flag

func init() {
	modelsFile = flag.Lookup("ds.models")
	if nil == modelsFile {
		flag.String("ds.models", "conf/tpt_models.xml", "the name of models file")
		modelsFile = flag.Lookup("ds.models")
	}
}

// Load 加载元模型数据，它将自动搜索各个目录加载文件
func Load(env *environment.Environment) (*TableDefinitions, error) {
	files := []string{modelsFile.Value.String(),
		filepath.Join("..", modelsFile.Value.String()),
		env.Fs.FromDataConfig("tpt_models.xml"),
		env.Fs.FromLib("models/tpt_models.xml"),
		"data/conf/tpt_models.xml",
		"data/etc/tpt_models.xml",
		"../data/conf/tpt_models.xml",
		"../data/etc/tpt_models.xml",
		"conf/tpt_models.xml",
		"etc/tpt_models.xml",
		"../conf/tpt_models.xml",
		"../etc/tpt_models.xml",
		"lib/models/tpt_models.xml",
		"../lib/models/tpt_models.xml",
		"lib/meta/tpt_models.xml",
		"../lib/meta/tpt_models.xml",
		"../meta/tpt_models.xml",
		"../../meta/tpt_models.xml",
		"../../../meta/tpt_models.xml",
		"../../../../meta/tpt_models.xml",
		"../../../../../meta/tpt_models.xml",
		"../cn/com/hengwei/meta/tpt_models.xml",
		"../../cn/com/hengwei/meta/tpt_models.xml",
		"../../../cn/com/hengwei/meta/tpt_models.xml",
		"../../../../cn/com/hengwei/meta/tpt_models.xml",
		"../../../../../cn/com/hengwei/meta/tpt_models.xml"}
	found := false
	for _, file := range files {
		if environment.FileExists(file) {
			flag.Set("ds.models", file)
			found = true
			break
		}
	}

	if !found {
		buf := bytes.NewBuffer(make([]byte, 0, 1024))
		buf.WriteString("models file is not exists, search path is:\r\n")
		for _, file := range files {
			buf.WriteString("    ")
			buf.WriteString(file)
			buf.WriteString("\r\n")
		}

		return nil, errors.New(buf.String())
	}

	definitions, e := LoadTableDefinitions(modelsFile.Value.String())
	if nil != e {
		return nil, fmt.Errorf("read file '%s' failed, %s", modelsFile.Value.String(), e.Error())
	}
	return definitions, nil
}

func LookupModelsFile(config_name, origin_name string, is_dir bool) (string, error) {
	files := []string{
		"../meta/" + origin_name,
		"../../meta/" + origin_name,
		"../../../meta/" + origin_name,
		"../../../../meta/" + origin_name,
		config_name,
		filepath.Join("..", config_name),
		origin_name,
		filepath.Join("..", origin_name),
		"conf/" + origin_name,
		"etc/" + origin_name,
		"../conf/" + origin_name,
		"../etc/" + origin_name,
		"lib/models/" + origin_name,
		"../lib/models/" + origin_name}

	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			continue
		}

		if is_dir {
			if info.IsDir() {
				return file, nil
			}
		} else {
			if !info.IsDir() {
				return file, nil
			}
		}
	}

	var buf bytes.Buffer
	buf.WriteString("search path is:")
	for _, file := range files {
		buf.WriteString("\r\n    ")
		buf.WriteString(file)
	}
	return "", errors.New(buf.String())
}
