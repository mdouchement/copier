package util

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/mdouchement/copier"
)

var (
	speedPattern = regexp.MustCompile(`(?i)^(-?\d+)([kmg]bps|[kmg])$`)
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
