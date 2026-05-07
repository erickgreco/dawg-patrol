package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/erickgreco/dawg-patrol/internal/robots"
	"github.com/erickgreco/dawg-patrol/internal/users"
)

var usernames = []string{
	"shadowbyte",
	"neo_runner",
	"pixelstorm",
	"zerocool",
	"ghostkernel",
	"bytehunter",
	"darkterminal",
	"hexdrifter",
	"syntaxerror",
	"rootaccess",
	"packetghost",
	"midnightdev",
	"stackbreaker",
	"silentproxy",
	"kernelpanic",
	"devphantom",
	"binarywolf",
	"nullpointer",
	"cybernomad",
	"gopherx",
	"tcpwizard",
	"segfaultsam",
	"cloudraider",
	"crypticnode",
	"overflowking",
}

var passwords = []string{
	"Shadow#2026",
	"NeoRun!42",
	"Pixel$Storm7",
	"ZeroCool#99",
	"Ghost@Kernel1",
	"ByteHunter!88",
	"DarkTerm#55",
	"HexDrifter@22",
	"Syntax!Error9",
	"RootAccess#77",
	"PacketGhost$5",
	"MidnightDev!3",
	"StackBreaker#1",
	"SilentProxy@66",
	"KernelPanic!404",
	"DevPhantom#12",
	"BinaryWolf@8",
	"NullPointer!500",
	"CyberNomad#23",
	"GopherX@2026",
	"TCPWizard!11",
	"SegFaultSam#7",
	"CloudRaider@14",
	"CrypticNode!64",
	"OverflowKing#2",
}

var robotNames = []string{
	"NoisyS1",
	"BlazeR1",
	"EchoA1",
	"VoltS2",
	"NovaR2",
	"PixelA2",
	"TitanS3",
	"DriftR3",
	"ScoutA3",
	"PhantomS4",
	"RocketR4",
	"BuddyA4",
	"ShadowS5",
	"TurboR5",
	"AtlasA5",
	"CrusherS6",
	"FlashR6",
	"HelperA6",
	"VortexS7",
	"NitroR7",
	"GuideA7",
	"RogueS8",
	"StormR8",
	"PulseA8",
	"HammerS9",
	"VelocityR9",
	"OrbitA9",
	"FuryS10",
	"RapidR10",
	"DeltaA10",
}

var serialNumbers = []string{
	"02:1A:C3:4D:5E:01",
	"02:1A:C3:4D:5E:02",
	"02:1A:C3:4D:5E:03",
	"02:1A:C3:4D:5E:04",
	"02:1A:C3:4D:5E:05",
	"02:1A:C3:4D:5E:06",
	"02:1A:C3:4D:5E:07",
	"02:1A:C3:4D:5E:08",
	"02:1A:C3:4D:5E:09",
	"02:1A:C3:4D:5E:0A",
	"02:1A:C3:4D:5E:0B",
	"02:1A:C3:4D:5E:0C",
	"02:1A:C3:4D:5E:0D",
	"02:1A:C3:4D:5E:0E",
	"02:1A:C3:4D:5E:0F",
	"02:1A:C3:4D:5E:10",
	"02:1A:C3:4D:5E:11",
	"02:1A:C3:4D:5E:12",
	"02:1A:C3:4D:5E:13",
	"02:1A:C3:4D:5E:14",
	"02:1A:C3:4D:5E:15",
	"02:1A:C3:4D:5E:16",
	"02:1A:C3:4D:5E:17",
	"02:1A:C3:4D:5E:18",
	"02:1A:C3:4D:5E:19",
	"02:1A:C3:4D:5E:1A",
	"02:1A:C3:4D:5E:1B",
	"02:1A:C3:4D:5E:1C",
	"02:1A:C3:4D:5E:1D",
	"02:1A:C3:4D:5E:1E",
}

var battery = []int64{
	92,
	78,
	65,
	100,
	54,
	81,
	39,
	73,
	88,
	47,
	95,
	60,
	34,
	69,
	83,
	51,
	97,
	28,
	76,
	58,
	90,
	44,
	71,
	86,
	63,
	31,
	99,
	52,
	79,
	67,
}

func generateUsers(num int) []*users.Registration {
	genUsers := make([]*users.Registration, num)

	for i := range num {
		genUsers[i] = &users.Registration{
			Username: usernames[i%len(usernames)],
			Email:    usernames[i%len(usernames)] + fmt.Sprintf("%d", i) + "@example.com",
			Password: passwords[i%len(passwords)],
		}
	}
	return genUsers
}

func generateRobots(num int) []*robots.RobotRegistration {
	genRobots := make([]*robots.RobotRegistration, num)

	for i := range num {
		genRobots[i] = &robots.RobotRegistration{
			SerialNumber: serialNumbers[i%len(serialNumbers)],
			Name:         robotNames[i%len(robotNames)],
			Battery:      battery[i%len(battery)],
		}
	}
	return genRobots
}

func Seed(userService *users.Service, robotService *robots.Service) {
	ctx := context.Background()

	file, err := os.OpenFile("./cmd/seed/logs/seed.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	logger := slog.New(slog.NewJSONHandler(file, nil))

	genUsers := generateUsers(25)
	for _, user := range genUsers {
		_, err := userService.UserRegistration(ctx, user)
		if err != nil {
			log.Println("error creating user: ", err)
			continue
		}
		logger.Info("user created", "username", user.Username, "email", user.Email, "password", user.Password)
	}

	genRobots := generateRobots(30)
	for _, robot := range genRobots {
		_, err := robotService.RobotRegistration(ctx, robot)
		if err != nil {
			log.Println("error creating robot: ", err)
			continue
		}
	}
}
