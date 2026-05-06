// Package created to avoid import cycle error
package domain

type Role string

// User domain info
const (
	RoleAdmin    Role = "ADMIN"
	RoleOperator Role = "OPERATOR"
	RoleViewer   Role = "VIEWER"
)

// Robot domain info
type Status string

const (
	RoleAssistant  Role   = "ASSISTANT"
	RoleSumo       Role   = "SUMO"
	RoleRacer      Role   = "RACER"
	IdleStatus     Status = "IDLE"
	InUseStatus    Status = "IN USE"
	ChargingStatus Status = "CHARGING"
	OfflineStatus  Status = "OFFLINE"
)

var RoleMap = map[string]string{
	"A": string(RoleAssistant),
	"S": string(RoleSumo),
	"R": string(RoleRacer),
}
