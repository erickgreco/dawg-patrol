package robots

import "fmt"

// Method that executes robot movement
func (r *RobotService) Execute(cmd Command) (string, error) {
	switch cmd.Type {
	case CommandForward:
		fmt.Println(CommandForward)
	case CommandBackward:
		fmt.Println(CommandBackward)
	case CommandTurnRight:
		fmt.Println(CommandTurnRight)
	case CommandTurnLeft:
		fmt.Println(CommandTurnLeft)
	default:
		fmt.Println(CommandStop)
	}
	return "Executing command ", nil
}
