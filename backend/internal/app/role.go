// Package app provides application-level configuration for the web/worker/all
// deployment role split. The role is read once from the APP_ROLE environment
// variable during init and exposed through package-level helpers so that
// provider functions can skip Start() calls for components that do not belong
// to the current role.
package app

import (
	"log"
	"os"
	"strings"
)

// Role represents the deployment role of this process.
type Role string

const (
	RoleAll    Role = "all"
	RoleWeb    Role = "web"
	RoleWorker Role = "worker"
)

// currentRole is set once during ParseRoleFromEnv and never mutated afterwards.
var currentRole = RoleAll

// ParseRoleFromEnv reads APP_ROLE from the environment and stores the result.
// Invalid values fall back to RoleAll. Call this once at the very beginning of
// main() before any dependency injection runs.
func ParseRoleFromEnv() Role {
	raw := strings.TrimSpace(strings.ToLower(os.Getenv("APP_ROLE")))
	switch Role(raw) {
	case RoleWeb:
		currentRole = RoleWeb
	case RoleWorker:
		currentRole = RoleWorker
	case RoleAll, "":
		currentRole = RoleAll
	default:
		log.Printf("[Role] Unknown APP_ROLE=%q, falling back to \"all\"", raw)
		currentRole = RoleAll
	}
	log.Printf("[Role] Running as %q", currentRole)
	return currentRole
}

// CurrentRole returns the role that was determined by ParseRoleFromEnv.
func CurrentRole() Role {
	return currentRole
}

// IsWebEnabled returns true when the process should serve HTTP traffic
// (role == web or role == all).
func IsWebEnabled() bool {
	return currentRole == RoleAll || currentRole == RoleWeb
}

// IsWorkerEnabled returns true when the process should run background
// workers / cron / cleanup tasks (role == worker or role == all).
func IsWorkerEnabled() bool {
	return currentRole == RoleAll || currentRole == RoleWorker
}
