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

	s1 := ratelimiterLuaFile

	if s1 == "" {
		s2 := fsx.GetRealpath("datadir:ratelimiter.lua")

		if stat, err := os.Stat(s2); err == nil && !stat.IsDir() {
			s1 = s2
		}
	}

	return s1
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

	s1 := ratelimiterCacheDir

	if s1 == "" {
		s2 := fsx.GetRealpath("datadir:cache")

		if stat, err := os.Stat(s2); err == nil && stat.IsDir() {
			s1 = s2
		}
	}

	return s1
}
