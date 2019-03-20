package environment

import (
	"path/filepath"

	"github.com/three-plus-three/modules/util"
)

// FileSystem 运行环境中文件系统的抽象
type FileSystem interface {
	FromRun(s ...string) string
	FromInstallRoot(s ...string) string
	FromWebConfig(s ...string) string
	//FromBin(s ...string) string
	FromLib(s ...string) string
	FromRuntimeEnv(s ...string) string
	FromData(s ...string) string
	FromTMP(s ...string) string
	FromConfig(s ...string) string
	FromLogDir(s ...string) string
	FromDataConfig(s ...string) string
	SearchConfig(s ...string) []string
}

type linuxFs struct {
	installDir string
	binDir     string
	logDir     string
	dataDir    string
	confDir    string
	tmpDir     string
	runDir     string

	// PACKAGE_NAME = "tpt"
	// INSTALL_ROOT_DIR = "/usr/local/tpt"
	// LOG_DIR = "/var/log/tpt"
	// DATA_DIR = "/var/lib/tpt"
	// SCRIPT_DIR = "/usr/lib/tpt/scripts"
	// CONFIG_DIR = "/etc/tpt"
	// Run_DIR = "/var/run/tpt/"
}

func (fs *linuxFs) FromInstallRoot(s ...string) string {
	return filepath.Join(fs.installDir, filepath.Join(s...))
}

func (fs *linuxFs) FromRun(s ...string) string {
	return filepath.Join(fs.runDir, filepath.Join(s...))
}

func (fs *linuxFs) FromWebConfig(s ...string) string {
	return filepath.Join(fs.confDir, "web", filepath.Join(s...))
}

func (fs *linuxFs) FromBin(s ...string) string {
	return filepath.Join(fs.binDir, filepath.Join(s...))
}

func (fs *linuxFs) FromLib(s ...string) string {
	return filepath.Join(fs.installDir, "lib", filepath.Join(s...))
}

func (fs *linuxFs) FromRuntimeEnv(s ...string) string {
	return filepath.Join(fs.installDir, "runtime_env", filepath.Join(s...))
}

func (fs *linuxFs) FromData(s ...string) string {
	return filepath.Join(fs.dataDir, filepath.Join(s...))
}

func (fs *linuxFs) FromTMP(s ...string) string {
	return filepath.Join(fs.tmpDir, filepath.Join(s...))
}

func (fs *linuxFs) FromConfig(s ...string) string {
	return filepath.Join(fs.installDir, "conf", filepath.Join(s...))
}

func (fs *linuxFs) FromDataConfig(s ...string) string {
	return filepath.Join(fs.confDir, filepath.Join(s...))
}

func (fs *linuxFs) FromLogDir(s ...string) string {
	return filepath.Join(fs.logDir, filepath.Join(s...))
}

func (fs *linuxFs) SearchConfig(s ...string) []string {
	var files []string
	for _, nm := range []string{fs.FromConfig(filepath.Join(s...)),
		fs.FromDataConfig(filepath.Join(s...))} {
		if util.FileExists(nm) {
			files = append(files, nm)
		}
	}
	return files
}
