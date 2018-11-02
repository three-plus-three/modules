package web_ext

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/three-plus-three/modules/environment"
)

func ReadVariables(env *environment.Environment, title string) map[string]interface{} {
	hw_root_dir := os.Getenv("hw_root_dir")

	if title == "" {
		title = "IT综合运维管理平台"
	}

	variables := map[string]interface{}{
		"application_catalog": env.Config.StringWithDefault("application.catalog", "all"),
		"version_text": ReadFileWithDefault([]string{
			env.Fs.FromInstallRoot("VERSION")}, "3.3"),

		"head_title_text": ReadFileWithDefault([]string{
			env.Fs.FromDataConfig("resources/profiles/header.txt"),
			env.Fs.FromData("resources/profiles/header.txt"),
			filepath.Join(hw_root_dir, "data/resources/profiles/header.txt")}, title),
		"footer_title_text": ReadFileWithDefault([]string{
			env.Fs.FromDataConfig("resources/profiles/footer.txt"),
			env.Fs.FromData("resources/profiles/footer.txt"),
			filepath.Join(hw_root_dir, "data/resources/profiles/footer.txt")}, "© 2017 恒维信息技术(上海)有限公司, 保留所有版权。"),
	}

	return variables
}

func ReadFileWithDefault(files []string, defaultValue string) string {
	for _, s := range files {
		content, e := ioutil.ReadFile(s)
		if nil == e {
			if content = bytes.TrimSpace(content); len(content) > 0 {
				return string(content)
			}
		}
	}
	return defaultValue
}
