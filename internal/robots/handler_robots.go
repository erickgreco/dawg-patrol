package robots

type RobotHandler struct {
	service *RobotService
}

func NewRobotHandler(service *RobotService) *RobotHandler {
	return &RobotHandler{
		service: service,
	}
}
