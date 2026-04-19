package main

func Direction(d string) string {
	switch d {
	case "N":
		return "north"
	case "S":
		return "south"
	case "E":
		return "east"
	case "W":
		return "west"
	default:
		return "unknown"
	}
}
