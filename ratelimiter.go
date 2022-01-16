package goboot

import (
	"github.com/lamgor666/goboot-common/util/fsx"
	"os"
)

var ratelimiterLuaFile string
var ratelimiterCacheDir string

func RatelimiterLuaFile(fpath ...string) string {
	if len(fpath) > 0 {
		if fpath[0] == "" {
			return ""
		}

		s1 := fsx.GetRealpath(fpath[0])

		if stat, err := os.Stat(s1); err == nil && !stat.IsDir() {
			ratelimiterLuaFile = s1
		}

		return ""
	}

	return ratelimiterLuaFile
}

func RatelimiterCacheDir(dir ...string) string {
	if len(dir) > 0 {
		if dir[0] == "" {
			return ""
		}

		s1 := fsx.GetRealpath(dir[0])

		if stat, err := os.Stat(s1); err == nil && stat.IsDir() {
			ratelimiterCacheDir = s1
		}

		return ""
	}

	return ratelimiterCacheDir
}
