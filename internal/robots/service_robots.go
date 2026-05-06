package robots

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/erickgreco/dawg-patrol/internal/domain"
	"github.com/erickgreco/dawg-patrol/pkg/myerrors"
	"github.com/google/uuid"
)

// Variable to define robot structure (default)
var (
	snRegex   = regexp.MustCompile(`^[0-9A-F]{2}(:[0-9A-F]{2}){5}$`)
	nameRegex = regexp.MustCompile(`^[A-Z][a-z]+[A-Z][0-9]+$`)
	roleRegex = regexp.MustCompile(`([A-Z])\d+$`)
)

type RobotsRepo interface {
	RegisterRobot(context.Context, *Robot) error
	RegisterEvent(context.Context, *RobotEvents) error
	GetIdleRobots(context.Context) ([]*RobotSummary, error)
}

type Service struct {
	store RobotsRepo
}

func NewRobotService(store RobotsRepo) *Service {
	return &Service{
		store: store,
	}
}

func (serv *Service) RobotRegistration(ctx context.Context, robot *RobotRegistration) (*RobotSummary, error) {
	serialNumber := NormalizeSN(robot.SerialNumber)

	if !ValidateSN(serialNumber) {
		return nil, myerrors.ErrInvalidSerialNumber
	}

	if !ValidateRobotName(robot.Name) {
		return nil, myerrors.ErrInvalidRobotName
	}

	if !ValidateBattery(robot.Battery) {
		return nil, myerrors.ErrBatteryOutOfRange
	}

	role, err := AssignRoleFromName(robot.Name)
	if err != nil {
		return nil, err
	}

	validatedRobot := &Robot{
		ID:           uuid.New(),
		SerialNumber: serialNumber,
		Name:         robot.Name,
		Role:         role,
		Battery:      robot.Battery,
		Status:       string(domain.IdleStatus),
		LastSeenAt:   time.Now(),
	}

	if err := serv.store.RegisterRobot(ctx, validatedRobot); err != nil {
		return nil, err
	}

	return &RobotSummary{
		ID:           validatedRobot.ID,
		SerialNumber: validatedRobot.SerialNumber,
		Name:         validatedRobot.Name,
		Role:         validatedRobot.Role,
		Status:       validatedRobot.Status,
		Battery:      validatedRobot.Battery,
		LastSeenAt:   validatedRobot.LastSeenAt,
	}, nil
}

// Helper that normalizes robot info
func NormalizeSN(str string) string {
	str = strings.TrimSpace(str)
	str = strings.ToUpper(str)
	str = strings.ReplaceAll(str, "-", ":")

	return str
}

/*
Helper to valide Serial Number
Default is XX:XX:XX:XX:XX:XX
*/
func ValidateSN(sn string) bool {
	return snRegex.MatchString(sn)
}

/*
Helper to validate Robot Name
Structure is:
Name + Role initial caps + robotNumber
*/
func ValidateRobotName(name string) bool {
	return nameRegex.MatchString(strings.TrimSpace(name))
}

/*
Helper to validate battery in case firmware sends invalid info
*/
func ValidateBattery(battery int64) bool {
	return battery >= 0 && battery <= 100
}

/*
Helper to auto asign role according to value in name
*/
func AssignRoleFromName(name string) (domain.Role, error) {
	name = strings.TrimSpace(name)

	matches := roleRegex.FindStringSubmatch(name)
	if len(matches) < 2 {
		return "", myerrors.ErrInvalidRobotName
	}

	roleLetter := matches[1]

	role, ok := domain.RoleMap[roleLetter]
	if !ok {
		return "", myerrors.ErrInvalidRobotRole
	}

	return domain.Role(role), nil
}
