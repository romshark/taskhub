package graph

import (
	"errors"
	"regexp"
)

var RegexpEmail = regexp.MustCompile(`/.+@.+\..+/i`)

func ValidateEmailAddress(s string) error {
	if !RegexpEmail.MatchString(s) {
		return errors.New("malformed email address")
	}
	return nil
}

func ValidateUserPassword(s string) error {
	if len(s) < 8 {
		return errors.New("password too short")
	}
	if len(s) > 1024 {
		return errors.New("password too long")
	}
	return nil
}

func ValidateUserDisplayName(s string) error {
	if len(s) < 4 {
		return errors.New("user displayName too short")
	}
	if len(s) > 256 {
		return errors.New("user displayName too long")
	}
	return nil
}

func ValidateUserRole(s string) error {
	if len(s) < 4 {
		return errors.New("user role too short")
	}
	if len(s) > 1024 {
		return errors.New("user role too long")
	}
	return nil
}

func ValidateUserLocation(s string) error {
	if len(s) > 1024 {
		return errors.New("user location too long")
	}
	return nil
}

func ValidateUserPersonalStatus(s string) error {
	if len(s) > 1024*64 {
		return errors.New("user personalStatus too long")
	}
	return nil
}

func ValidateProjectName(s string) error {
	if len(s) < 4 {
		return errors.New("project name too short")
	}
	if len(s) > 256 {
		return errors.New("project name too long")
	}
	return nil
}

func ValidateProjectDescription(s string) error {
	if len(s) > 1024*64 {
		return errors.New("project description too long")
	}
	return nil
}

func ValidateProjectSlug(s string) error {
	if len(s) > 6 {
		return errors.New("project slug too long")
	}
	return nil
}

func ValidateTaskTitle(s string) error {
	if len(s) > 1024 {
		return errors.New("task title too long")
	}
	return nil
}

func ValidateTaskTag(s string) error {
	if len(s) < 1 {
		return errors.New("tag too short")
	}
	if len(s) > 1024 {
		return errors.New("tag too long")
	}
	return nil
}
