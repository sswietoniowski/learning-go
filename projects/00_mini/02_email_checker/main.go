package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("Enter an e-mail address: ")
	scanner.Scan()
	email := scanner.Text()

	parts := strings.Split(email, "@")

	if len(parts) != 2 {
		fmt.Println("Invalid e-mail address")
		return
	}

	domain := parts[1]

	if checkDomain(domain) {
		fmt.Println("Probably a valid e-mail: domain exists")
	} else {
		fmt.Println("Invalid e-mail domain")
	}
}

func checkDomain(domain string) bool {
	var hasSPF, hasDMARC bool
	var spfRecord, dmarcRecord string

	mxRecords, err := net.LookupMX(domain)
	if err != nil {
		fmt.Println("No MX records found")
		return false
	}

	txRecords, err := net.LookupTXT(domain)
	if err != nil {
		fmt.Println("No TXT records found")
		return false
	}

	for _, record := range txRecords {
		if strings.HasPrefix(record, "v=spf1") {
			hasSPF = true
			spfRecord = record
			break
		}
	}

	if !hasSPF {
		fmt.Println("No SPF record found")
		return false
	}

	dmarcRecords, err := net.LookupTXT("_dmarc." + domain)
	if err != nil {
		fmt.Println("No DMARC record found")
		return false
	}

	for _, record := range dmarcRecords {
		if strings.HasPrefix(record, "v=DMARC1") {
			hasDMARC = true
			dmarcRecord = record
			break
		}
	}

	if !hasDMARC {
		fmt.Println("No DMARC record found")
		return false
	}

	fmt.Println("Domain has MX records: ", mxRecords)
	fmt.Println("Domain has SPF record: ", spfRecord)
	fmt.Println("Domain has DMARC record: ", dmarcRecord)

	return true
}
