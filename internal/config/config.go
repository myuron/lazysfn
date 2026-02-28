// Package config provides parsing of AWS configuration files.
package config

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
)

// Profile represents an AWS CLI profile with its name and region.
type Profile struct {
	Name   string
	Region string
}

// ParseProfiles reads an AWS config file and returns a sorted list of profiles.
func ParseProfiles(path string) ([]Profile, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening config file: %w", err)
	}
	defer func() { _ = f.Close() }()

	var profiles []Profile
	var current *Profile

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			header := line[1 : len(line)-1]
			name := ""
			switch {
			case header == "default":
				name = "default"
			case strings.HasPrefix(header, "profile "):
				name = strings.TrimSpace(strings.TrimPrefix(header, "profile "))
			default:
				current = nil
				continue
			}
			if name != "" {
				profiles = append(profiles, Profile{Name: name})
				current = &profiles[len(profiles)-1]
			}
			continue
		}

		if current == nil {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if key == "region" {
			current.Region = value
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	sort.Slice(profiles, func(i, j int) bool {
		return profiles[i].Name < profiles[j].Name
	})

	return profiles, nil
}
