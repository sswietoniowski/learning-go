package api

// Version is the version of the application.
const Version = "1.0.0"

// Config is the configuration for the application.
type Config struct {
	ServerPort      int
	EnvironmentName string
	DatabaseType    string
	FrontendOrigin  string
}
