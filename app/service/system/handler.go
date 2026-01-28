package system

import "strings"

func isAvailableName(name string) bool {
	if name == "" {
		return true
	}

	var forbiddenName = []string{"admin", "root", "administrator", "管理员", "超级管理员", "seakee", "super_admin", "superAdmin"}

	name = strings.ToLower(name)
	for _, n := range forbiddenName {
		if strings.HasPrefix(name, n) || strings.HasSuffix(name, n) {
			return false
		}
	}

	return true
}
