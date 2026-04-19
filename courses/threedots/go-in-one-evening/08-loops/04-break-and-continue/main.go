package main

import "strings"

func CountCreatedEvents(events []string) int {
	count := 0

	for _, event := range events {
		if event == "client_deleted" {
			break
		}

		if event == "order_updated" {
			continue
		}

		if strings.HasSuffix(event, "_created") {
			count++
		}
	}

	return count
}

func main() {
	events := []string{
		"product_created",
		"product_updated",
		"product_assigned",
		"order_created",
		"order_updated",
		"client_created",
		"client_updated",
		"client_refreshed",
		"client_deleted",
		"order_updated",
	}

	CountCreatedEvents(events)
}
