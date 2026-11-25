package main

import "fmt"

const defaultValue = "N/A"

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func printBuildValue(label, value string) {
	if value == "" {
		value = defaultValue
	}
	fmt.Printf("%s: %s\n", label, value)
}

func printBuildInfo() {
	printBuildValue("Build version", buildVersion)
	printBuildValue("Build date", buildDate)
	printBuildValue("Build commit", buildCommit)
}
