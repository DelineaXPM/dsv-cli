//go:build endtoend
// +build endtoend

package e2e

import (
	"fmt"
	"testing"
	"time"
)

func TestUsage(t *testing.T) {
	now := time.Now()
	current := now.Format("2006-01-02")
	pastSevenDays := now.AddDate(0, 0, -7).Format("2006-01-02")
	futureSevenDays := now.AddDate(0, 0, 7).Format("2006-01-02")

	output := runWithProfile(t, fmt.Sprintf("usage --enddate %s", futureSevenDays))
	requireContains(t, output, "error: must specify --startdate")

	// If `--enddate` is not used, today's date should be assumed.
	output = runWithProfile(t, fmt.Sprintf("usage --startdate %s", pastSevenDays))
	requireContains(t, output, fmt.Sprintf(`"startDate": "%s"`, pastSevenDays))
	requireContains(t, output, fmt.Sprintf(`"endDate": "%s"`, current))

	// Test when `--enddate` is used.
	output = runWithProfile(t, fmt.Sprintf(
		"usage --startdate %s --enddate %s", pastSevenDays, futureSevenDays,
	))
	requireContains(t, output, fmt.Sprintf(`"startDate": "%s"`, pastSevenDays))
	requireContains(t, output, fmt.Sprintf(`"endDate": "%s"`, futureSevenDays))
}
