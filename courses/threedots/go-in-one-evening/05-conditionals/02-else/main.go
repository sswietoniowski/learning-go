package main

func DescribeNumber(x int) string {
	if x < 0 {
		return "negative"
	} else if x > 0 {
		return "positive"
	} else {
		return "zero"
	}
}
