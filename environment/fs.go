package environment

import "path/filepath"

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

type winFs struct {
	rootDir string
}

func (wfs *winFs) FromRun(s ...string) string {
	return filepath.Join(wfs.rootDir, filepath.Join(s...))
}

func (wfs *winFs) FromInstallRoot(s ...string) string {
	return filepath.Join(wfs.rootDir, filepath.Join(s...))
}

func (wfs *winFs) FromWebConfig(s ...string) string {
	return filepath.Join(wfs.rootDir, "web", "conf", filepath.Join(s...))
}

func (wfs *winFs) FromBin(s ...string) string {
	return filepath.Join(wfs.rootDir, "bin", filepath.Join(s...))
}

func (wfs *winFs) FromLib(s ...string) string {
	return filepath.Join(wfs.rootDir, "lib", filepath.Join(s...))
}

func (wfs *winFs) FromLogDir(s ...string) string {
	return filepath.Join(wfs.rootDir, "logs", filepath.Join(s...))
}

func (wfs *winFs) FromRuntimeEnv(s ...string) string {
	return filepath.Join(wfs.rootDir, "runtime_env", filepath.Join(s...))
}

func (wfs *winFs) FromData(s ...string) string {
	return filepath.Join(wfs.rootDir, "data", filepath.Join(s...))
}

func (wfs *winFs) FromTMP(s ...string) string {
	return filepath.Join(wfs.rootDir, "data", "tmp", filepath.Join(s...))
}

func (wfs *winFs) FromConfig(s ...string) string {
	return filepath.Join(wfs.rootDir, "conf", filepath.Join(s...))
}

func (wfs *winFs) FromDataConfig(s ...string) string {
	return filepath.Join(wfs.rootDir, "data", "conf", filepath.Join(s...))
}

func (wfs *winFs) SearchConfig(s ...string) []string {
	var files []string
	for _, nm := range []string{filepath.Join(wfs.rootDir, "conf", filepath.Join(s...)),
		filepath.Join(wfs.rootDir, "etc", filepath.Join(s...)),
		filepath.Join(wfs.rootDir, "data", "conf", filepath.Join(s...)),
		filepath.Join(wfs.rootDir, "data", "etc", filepath.Join(s...))} {
		if FileExists(nm) {
			files = append(files, nm)
		}
	}
	return files
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
		if FileExists(nm) {
			files = append(files, nm)
		}
	}
	return files
}
