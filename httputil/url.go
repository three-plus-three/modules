package httputil

import "github.com/three-plus-three/modules/urlutil"

// SplitURLPath 分隔 url path, 取出 url path 的第一部份
func SplitURLPath(pa string) (string, string) {
	return urlutil.Split(pa)
}

// JoinURLPath 拼接 url
func JoinURLPath(paths ...string) string {
	return urlutil.Join(pa...)
}

// JoinURLPathWith 拼接 url
func JoinURLPathWith(base string, paths []string) string {
	return urlutil.JoinWith(base, paths)
}

// NewURLBuilder 创建 url builder
func NewURLBuilder(base string) *urlutil.URLBuilder {
	return urlutil.NewURLBuilder(base)
}

func ConcatURLPaths(paths ...string) string {
	return urlutil.Join(paths...)
}
