package goboot

import (
	"github.com/lamgor666/goboot-common/util/fsx"
	"os"
)

var redislockLuaFileLock string
var redislockLuaFileUnlock string
var redislockCacheDir string

func RedislockLuaFile(typ string, fpath ...string) string {
	if len(fpath) > 0 {
		if fpath[0] == "" {
			return ""
		}

		s1 := fsx.GetRealpath(fpath[0])

		if stat, err := os.Stat(s1); err == nil && !stat.IsDir() {
			switch typ {
			case "lock":
				redislockLuaFileLock = s1
			case "unlock":
				redislockLuaFileUnlock = s1
			}
		}

		return ""
	}

	var s1 string

	switch typ {
	case "lock":
		s1 = redislockLuaFileLock
	case "unlock":
		s1 = redislockLuaFileUnlock
	}

	if s1 == "" {
		switch typ {
		case "lock":
			s2 := fsx.GetRealpath("datadir:redislock.lock.lua")

			if stat, err := os.Stat(s2); err == nil && !stat.IsDir() {
				s1 = s2
			}
		case "unlock":
			s2 := fsx.GetRealpath("datadir:redislock.unlock.lua")

			if stat, err := os.Stat(s2); err == nil && !stat.IsDir() {
				s1 = s2
			}
		}
	}

	return s1
}

func RedislockCacheDir(dir ...string) string {
	if len(dir) > 0 {
		if dir[0] == "" {
			return ""
		}

		s1 := fsx.GetRealpath(dir[0])

		if stat, err := os.Stat(s1); err == nil && stat.IsDir() {
			redislockCacheDir = s1
		}

		return ""
	}

	s1 := redislockCacheDir

	if s1 == "" {
		s2 := fsx.GetRealpath("datadir:cache")

		if stat, err := os.Stat(s2); err == nil && stat.IsDir() {
			s1 = s2
		}
	}

	return s1
}
