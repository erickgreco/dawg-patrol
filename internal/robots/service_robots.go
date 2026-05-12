package robots

import (
	"context"
	"errors"
	"math/rand"
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
	typeRegex = regexp.MustCompile(`([A-Z])\d+$`)
)

const (
	highLevel    = "Battery level under 80%"
	mediumLevel  = "Battery level under 50%"
	lowLevel     = "Battery level under 20%"
	invalidLevel = "Battery level insufficient"
	validLevel   = "Battery level above 80%"
)

type RobotsRepo interface {
	RegisterRobot(context.Context, *Robot) error
	RegisterEvent(context.Context, *RobotEvents) error
	GetIdleRobots(context.Context) ([]*RobotSummary, error)
	GetUnavailableRobots(context.Context) ([]*RobotSummary, error)
	GetByID(ctx context.Context, robotID uuid.UUID) (*RobotSummary, error)
	ReserveRobot(ctx context.Context, reservationID, userID, robotID uuid.UUID) (*RobotReservation, error)
	CleanExpiredReservations(ctx context.Context) error
	ValidateReservation(ctx context.Context, reservationID, userID, robotID uuid.UUID) (uuid.UUID, error)
	GetReservationByID(ctx context.Context, reservationID uuid.UUID) (*RobotReservation, error)
	ExtendReservation(ctx context.Context, reservationID uuid.UUID) error
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

	category, err := AssignTypeFromName(robot.Name)
	if err != nil {
		return nil, err
	}

	validatedRobot := &Robot{
		ID:           uuid.New(),
		SerialNumber: serialNumber,
		Name:         robot.Name,
		Category:     category,
		Battery:      robot.Battery,
		Status:       string(RandomStatus()),
		LastSeenAt:   time.Now(),
	}

	if err := serv.store.RegisterRobot(ctx, validatedRobot); err != nil {
		return nil, err
	}

	return &RobotSummary{
		ID:           validatedRobot.ID,
		SerialNumber: validatedRobot.SerialNumber,
		Name:         validatedRobot.Name,
		Category:     validatedRobot.Category,
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
Helper to auto asign type according to value in name
Robot name structure is:
Name + Caps role initial + Number of robot
Ex: NoisyA1 (in this case Noisy is assistant)
*/
func AssignTypeFromName(name string) (domain.Category, error) {
	name = strings.TrimSpace(name)

	matches := typeRegex.FindStringSubmatch(name)
	if len(matches) < 2 {
		return "", myerrors.ErrInvalidRobotName
	}

	typeLetter := matches[1]

	category, ok := domain.TypeMap[typeLetter]
	if !ok {
		return "", myerrors.ErrInvalidRobotType
	}

	return domain.Category(category), nil
}

/*
This method will apply business logic as it grows, will work with roles
*/
func (serv *Service) IdleRobots(ctx context.Context) (*IdleRobots, error) {
	idleRobots, err := serv.store.GetIdleRobots(ctx)
	if err != nil {
		return nil, err
	}

	var assistants []*RobotSummary
	var sumos []*RobotSummary
	var racers []*RobotSummary

	for _, robot := range idleRobots {
		switch robot.Category {
		case domain.TypeAssistant:
			assistants = append(assistants, robot)
		case domain.TypeSumo:
			sumos = append(sumos, robot)
		case domain.TypeRacer:
			racers = append(racers, robot)
		}
	}
	return &IdleRobots{
		AssistantRobots: assistants,
		SumoRobots:      sumos,
		RacerRobots:     racers,
	}, nil
}

// This helper was created to be able to random add type while creating robot (used for seed)
func RandomStatus() domain.Status {
	categories := []domain.Status{domain.IdleStatus, domain.InUseStatus, domain.ChargingStatus, domain.OfflineStatus}

	robotType := categories[rand.Intn(len(categories))]

	return robotType
}

/*
Method implements battery validation before robot reserve,
previous to starting ws conn
*/
func (serv *Service) RobotByID(ctx context.Context, robotID uuid.UUID) (*RobotSummary, error) {
	robot, err := serv.store.GetByID(ctx, robotID)
	if err != nil {
		if errors.Is(err, myerrors.ErrDataNotFound) {
			return nil, myerrors.ErrUnavailableRobot
		}
		return nil, err
	}

	batteryLevel := BatteryStatus(robot.Battery)

	return &RobotSummary{
		ID:           robot.ID,
		SerialNumber: robot.SerialNumber,
		Name:         robot.Name,
		Category:     robot.Category,
		Status:       batteryLevel,
		Battery:      robot.Battery,
		LastSeenAt:   robot.LastSeenAt,
	}, nil
}

func BatteryStatus(level int64) string {
	switch {
	case level >= 80:
		return validLevel
	case level >= 50:
		return highLevel
	case level >= 20:
		return mediumLevel
	case level >= 10:
		return lowLevel
	default:
		return invalidLevel
	}
}

func (serv *Service) RobotReservation(ctx context.Context, robotID, userID uuid.UUID) (*RobotReservation, error) {
	robot, err := serv.store.GetByID(ctx, robotID)
	if err != nil {
		if errors.Is(err, myerrors.ErrDataNotFound) {
			return nil, myerrors.ErrRobotNotFound
		}
		return nil, err
	}

	if robot.Status != string(domain.IdleStatus) {
		return nil, myerrors.ErrUnavailableRobot
	}

	batteryLevel := BatteryStatus(robot.Battery)

	if batteryLevel == lowLevel {
		return nil, myerrors.ErrLowBatteryLevel
	}

	reservationID := uuid.New()

	reservation, err := serv.store.ReserveRobot(ctx, reservationID, userID, robot.ID)
	if err != nil {
		if errors.Is(err, myerrors.ErrUnavailableRobot) {
			return nil, myerrors.ErrUnavailableRobot
		}
		return nil, err
	}

	return reservation, nil
}

func (serv *Service) ReservationByID(ctx context.Context, reservationID uuid.UUID) (*RobotReservation, error) {
	reservation, err := serv.store.GetReservationByID(ctx, reservationID)
	if err != nil {
		if errors.Is(err, myerrors.ErrInvalidReservation) {
			return nil, myerrors.ErrInvalidReservation
		}
		return nil, err
	}
	return reservation, nil
}
