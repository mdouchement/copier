package util

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/mdouchement/copier"
)

var (
	speedPattern   = regexp.MustCompile(`(?i)^(-?\d+)([kmg]bps|[kmg])$`)
	timeoutPattern = regexp.MustCompile(`(?i)^(-?\d+)([smh])$`)
)

// ParseSpeed evluates the speed defined as string.
func ParseSpeed(value string) (i int, err error) {
	value = strings.ToLower(value)
	parts := speedPattern.FindStringSubmatch(value)
	if len(parts) < 3 {
		return 0, fmt.Errorf("error parsing value=%s", value)
	}
	speedString := parts[1]
	multiple := parts[2]
	speed, err := strconv.Atoi(speedString)
	if err != nil {
		return
	}

	switch multiple {
	case "", "bps":
		return speed * copier.Bps, nil
	case "k", "kbps":
		return speed * copier.KBps, nil
	case "m", "mbps":
		return speed * copier.MBps, nil
	case "g", "gbps":
		return speed * copier.GBps, nil
	}

	return
}

// ParseTimeout evluates the timeout defined as string.
func ParseTimeout(value string) (i time.Duration, err error) {
	value = strings.ToLower(value)
	parts := timeoutPattern.FindStringSubmatch(value)
	if len(parts) < 3 {
		return 0, fmt.Errorf("error parsing value=%s", value)
	}
	timeString := parts[1]
	multiple := parts[2]
	timeout, err := strconv.Atoi(timeString)
	if err != nil {
		return
	}

	switch multiple {
	case "s":
		return time.Duration(timeout) * time.Second, nil
	case "m":
		return time.Duration(timeout) * time.Minute, nil
	case "h":
		return time.Duration(timeout) * time.Hour, nil
	case "d":
		return time.Duration(timeout) * time.Hour * 24, nil
	}

	return
}
