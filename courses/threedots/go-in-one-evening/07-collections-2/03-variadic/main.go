package main

func DebugLog(args ...string) []string {
	return append([]string{"[DEBUG]"}, args...)
}

func InfoLog(args ...string) []string {
	return append([]string{"[INFO]"}, args...)
}

func ErrorLog(args ...string) []string {
	return append([]string{"[ERROR]"}, args...)
}
