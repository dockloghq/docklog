package models

import "github.com/golang-jwt/jwt/v5"

type Container struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Image        string   `json:"image"`
	State        string   `json:"state"`
	Created      int64    `json:"created"`
	Status       string   `json:"status"`
	CPULimit     float64  `json:"cpu_limit"`
	MemLimit     int64    `json:"mem_limit"`
	CPU          float64  `json:"cpu"`
	Memory       int64    `json:"memory"`
	IsSwarm      bool     `json:"is_swarm"`
	SwarmService string   `json:"swarm_service,omitempty"`
	Ports        []string `json:"ports,omitempty"`
}

type UserClaims struct {
	ID                 int    `json:"id"`
	Username           string `json:"username"`
	IsAdmin            bool   `json:"is_admin"`
	IsRestrictedAccess bool   `json:"is_restricted_access"`
	CanStart           bool   `json:"can_start"`
	CanStop            bool   `json:"can_stop"`
	CanRestart         bool   `json:"can_restart"`
	CanDelete          bool   `json:"can_delete"`
	CanShell           bool   `json:"can_shell"`
	AllowedContainers  string `json:"allowed_containers"`
	IsActive           bool   `json:"is_active"`
	PasswordVersion    int    `json:"password_version"`
	TokenType          string `json:"token_type,omitempty"`
	jwt.RegisteredClaims
}

type User struct {
	ID                 int    `json:"id"`
	Username           string `json:"username"`
	IsAdmin            bool   `json:"is_admin"`
	PasswordChanged    bool   `json:"password_changed"`
	CanStart           bool   `json:"can_start"`
	CanStop            bool   `json:"can_stop"`
	CanRestart         bool   `json:"can_restart"`
	CanDelete          bool   `json:"can_delete"`
	CanShell           bool   `json:"can_shell"`
	IsRestrictedAccess bool   `json:"is_restricted_access"`
	AllowedContainers  string `json:"allowed_containers"`
	IsActive           bool   `json:"is_active"`
}
