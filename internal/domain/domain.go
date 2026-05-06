// Package created to avoid import cycle error
package domain

type Role string

type Category string

// User domain info
const (
	RoleAdmin    Role = "ADMIN"
	RoleOperator Role = "OPERATOR"
	RoleViewer   Role = "VIEWER"
)

// Robot domain info
type Status string

const (
	TypeAssistant  Category = "ASSISTANT"
	TypeSumo       Category = "SUMO"
	TypeRacer      Category = "RACER"
	IdleStatus     Status   = "IDLE"
	InUseStatus    Status   = "IN USE"
	ChargingStatus Status   = "CHARGING"
	OfflineStatus  Status   = "OFFLINE"
)

var TypeMap = map[string]string{
	"A": string(TypeAssistant),
	"S": string(TypeSumo),
	"R": string(TypeRacer),
}
