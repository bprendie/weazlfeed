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
		fmt.Printf("%s:\n", label)
		for i, choice := range choices {
			fmt.Printf("  %d) %s\n", i+1, choice)
		}
		answer := askString(reader, "Select", def)
		if answer == "" {
			return def
		}
		if n, err := strconv.Atoi(answer); err == nil && n >= 1 && n <= len(choices) {
			return choices[n-1]
		}
		for _, choice := range choices {
			if strings.EqualFold(answer, choice) {
				return choice
			}
		}
		fmt.Println("Enter a menu number or choice name.")
	}
}

func askModel(reader *bufio.Reader, models []string) string {
	for {
		fmt.Println("Models:")
		for i, model := range models {
			fmt.Printf("  %d) %s\n", i+1, model)
		}
		answer := askString(reader, "Select model", "1")
		n, err := strconv.Atoi(answer)
		if err == nil && n >= 1 && n <= len(models) {
			return models[n-1]
		}
		if answer != "" && contains(models, answer) {
			return answer
		}
		fmt.Println("Enter a menu number or exact model name.")
	}
}

func askContextWindow(reader *bufio.Reader) int {
	choices := []struct {
		Name   string
		Tokens int
		Note   string
	}{
		{Name: "small", Tokens: 8192},
		{Name: "medium", Tokens: 16384},
		{Name: "large", Tokens: 32768},
		{Name: "xl", Tokens: 128000, Note: "may cause out-of-memory errors on smaller local servers"},
	}
	for {
		fmt.Println("Context window:")
		for i, choice := range choices {
			label := fmt.Sprintf("  %d) %s (%d tokens)", i+1, choice.Name, choice.Tokens)
			if choice.Note != "" {
				label += " - " + choice.Note
			}
			fmt.Println(label)
		}
		answer := askString(reader, "Select", "large")
		if answer == "" {
			return 32768
		}
		if n, err := strconv.Atoi(answer); err == nil && n >= 1 && n <= len(choices) {
			return choices[n-1].Tokens
		}
		for _, choice := range choices {
			if strings.EqualFold(answer, choice.Name) {
				return choice.Tokens
			}
		}
		fmt.Println("Enter small, medium, large, xl, or a menu number.")
	}
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
