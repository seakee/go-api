package system

import (
	"regexp"
	"strings"
)

var (
	emailRegexp = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
	phoneRegexp = regexp.MustCompile(`^\+?[0-9]{6,20}$`)
)

func isAvailableName(name string) bool {
	if name == "" {
		return true
	}

	var forbiddenName = []string{"admin", "root", "administrator", "管理员", "超级管理员", "super_admin", "superAdmin"}

	name = strings.ToLower(name)
	for _, n := range forbiddenName {
		if strings.HasPrefix(name, n) || strings.HasSuffix(name, n) {
			return false
		}
	}

	return true
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func normalizePhone(phone string) string {
	return strings.TrimSpace(phone)
}

func isValidEmail(email string) bool {
	if email == "" {
		return false
	}

	return emailRegexp.MatchString(email)
}

func isValidPhone(phone string) bool {
	if phone == "" {
		return false
	}

	return phoneRegexp.MatchString(phone)
}

func isEmailIdentifier(identifier string) bool {
	return isValidEmail(normalizeEmail(identifier))
}

func isPhoneIdentifier(identifier string) bool {
	return isValidPhone(normalizePhone(identifier))
}
