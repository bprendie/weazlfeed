package main

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
)

func askString(reader *bufio.Reader, label, def string) string {
	if def != "" {
		fmt.Printf("%s [%s]: ", label, def)
	} else {
		fmt.Printf("%s: ", label)
	}
	value, _ := reader.ReadString('\n')
	value = strings.TrimSpace(value)
	if value == "" {
		return def
	}
	return value
}

func askChoice(reader *bufio.Reader, label string, choices []string, def string) string {
	for {
		value := askString(reader, label+" ("+strings.Join(choices, "/")+")", def)
		for _, choice := range choices {
			if value == choice {
				return value
			}
		}
		fmt.Println("Choose one of: " + strings.Join(choices, ", "))
	}
}

func askContextWindow(reader *bufio.Reader) int {
	value := askString(reader, "Context window", "32768")
	n, err := strconv.Atoi(value)
	if err != nil || n <= 0 {
		return 32768
	}
	return n
}

func defaultModel(providerType string) string {
	if providerType == "ollama" {
		return "llama3.1"
	}
	return "local-model"
}

func normalizeServerURL(value string) string {
	return strings.TrimRight(strings.TrimSpace(value), "/")
}
