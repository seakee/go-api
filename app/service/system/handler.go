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

func maskCompactValue(value string) string {
	runes := []rune(value)
	switch len(runes) {
	case 0:
		return ""
	case 1:
		return "*"
	case 2:
		return string(runes[:1]) + "*"
	default:
		return string(runes[:1]) + strings.Repeat("*", len(runes)-2) + string(runes[len(runes)-1:])
	}
}

func maskEmail(email string) string {
	email = normalizeEmail(email)
	if email == "" {
		return ""
	}

	local, domain, ok := strings.Cut(email, "@")
	if !ok {
		return maskCompactValue(email)
	}

	maskedLocal := maskCompactValue(local)
	host, suffix, ok := strings.Cut(domain, ".")
	if !ok {
		return maskedLocal + "@" + maskCompactValue(domain)
	}

	return maskedLocal + "@" + maskCompactValue(host) + "." + suffix
}

func maskPhone(phone string) string {
	phone = normalizePhone(phone)
	if phone == "" {
		return ""
	}

	runes := []rune(phone)
	const prefixLen = 3
	const suffixLen = 4

	if len(runes) <= prefixLen+suffixLen {
		return maskCompactValue(phone)
	}

	return string(runes[:prefixLen]) +
		strings.Repeat("*", len(runes)-prefixLen-suffixLen) +
		string(runes[len(runes)-suffixLen:])
}

func restoreMaskedEmailInput(input, current string) string {
	current = normalizeEmail(current)
	if input == "" || current == "" {
		return input
	}
	if input == maskEmail(current) {
		return current
	}

	return input
}

func restoreMaskedPhoneInput(input, current string) string {
	current = normalizePhone(current)
	if input == "" || current == "" {
		return input
	}
	if input == maskPhone(current) {
		return current
	}

	return input
}
