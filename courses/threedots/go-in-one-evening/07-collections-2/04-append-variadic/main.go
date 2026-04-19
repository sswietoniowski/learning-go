package main

func Merge(a []string, b []string) []string {
	return append(a, b...)
}

func Remove(a []string, index int) []string {
	return append(a[:index], a[index+1:]...)
}

func RemoveLast(a []string) []string {
	return a[:len(a)-1]
}
