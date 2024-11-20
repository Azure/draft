package config

import (
	"fmt"
	"slices"
	"strings"

	"github.com/blang/semver/v4"
)

// GetValidTemplateVersions converts a version range into a list of valid versions.
// expects bounded patch versions eg: >=0.0.1 <=0.0.5 || >0.1.0 <0.1.5 and not unbounded patch versions eg: >=0.0.1 <0.1.5
func GetValidTemplateVersions(inputVersion string) ([]string, error) {
	validRanges := strings.Split(inputVersion, "||")
	if len(validRanges) == 0 {
		return nil, fmt.Errorf("no versions found for input %s", inputVersion)
	}

	allowedVersions := make([]semver.Version, 0)
	invalidVersions := map[string]any{}
	for _, validRange := range validRanges {
		versions := strings.Split(strings.Trim(validRange, " "), " ")
		lessThanVersions := make([]semver.Version, 0)
		greaterthanVersions := make([]semver.Version, 0)
		for _, version := range versions {
			versionCondition := getVersionCondition(version)
			semVersion, err := semver.Parse(strings.Trim(version, versionCondition))
			if err != nil {
				return nil, err
			}

			switch versionCondition {
			case "<=":
				lessThanVersions = append(lessThanVersions, semVersion)
			case ">=":
				greaterthanVersions = append(greaterthanVersions, semVersion)
			case "<":
				if semVersion.Patch == 0 {
					return nil, fmt.Errorf("unbounded version range: %s", validRange)
				}
				semVersion.Patch--
				lessThanVersions = append(lessThanVersions, semVersion)
			case ">":
				err = semVersion.IncrementPatch()
				if err != nil {
					return nil, fmt.Errorf("failed to increment patch version: %w", err)
				}
				greaterthanVersions = append(greaterthanVersions, semVersion)
			case "=", "==":
				allowedVersions = append(allowedVersions, semVersion)
			case "!=", "!":
				invalidVersions[semVersion.String()] = true
			default:
				allowedVersions = append(allowedVersions, semVersion)
			}
		}

		for _, greaterThanVersion := range greaterthanVersions {
			hasBound := false
			var maxVersionDiff *semver.Version
			for _, lessThanVersion := range lessThanVersions {
				if greaterThanVersion.GT(lessThanVersion) {
					continue
				}

				versionDiff, err := semver.Parse(fmt.Sprintf("%d.%d.%d", lessThanVersion.Major-greaterThanVersion.Major, lessThanVersion.Minor-greaterThanVersion.Minor, lessThanVersion.Patch-greaterThanVersion.Patch))
				if err != nil {
					return nil, fmt.Errorf("failed to parse version difference: %w", err)
				}

				if versionDiff.Major != 0 || versionDiff.Minor != 0 {
					continue
				}

				if maxVersionDiff == nil {
					maxVersionDiff = &versionDiff
				}

				if versionDiff.Patch > maxVersionDiff.Patch {
					maxVersionDiff = &versionDiff
				}
				hasBound = true
			}
			if !hasBound {
				return nil, fmt.Errorf("unbounded version range: %s", validRange)
			}

			allowedVersions = append(allowedVersions, greaterThanVersion)
			for i := 0; i < int(maxVersionDiff.Patch); i++ {
				nextPatch, err := semver.Parse(fmt.Sprintf("%d.%d.%d", greaterThanVersion.Major, greaterThanVersion.Minor, int(greaterThanVersion.Patch)+i+1))
				if err != nil {
					return nil, fmt.Errorf("failed to parse next patch version: %w", err)
				}
				allowedVersions = append(allowedVersions, nextPatch)
			}

		}
	}

	if len(allowedVersions) == 0 {
		return nil, fmt.Errorf("no valid versions found for versions %s", inputVersion)
	}

	// sort versions
	slices.SortFunc(allowedVersions, func(i, j semver.Version) int {
		if i.GT(j) {
			return 1
		}
		if i.LT(j) {
			return -1
		}
		return 0
	})

	allowedVersionStrings := make([]string, 0)
	alreadyHandledVersions := map[string]any{}
	for _, version := range allowedVersions {
		if _, ok := alreadyHandledVersions[version.String()]; !ok {
			if _, ok := invalidVersions[version.String()]; !ok {
				allowedVersionStrings = append(allowedVersionStrings, version.String())
			}
			alreadyHandledVersions[version.String()] = true
		}
	}

	return allowedVersionStrings, nil
}

func getVersionCondition(version string) string {
	if strings.HasPrefix(version, "<=") {
		return "<="
	} else if strings.HasPrefix(version, ">=") {
		return ">="
	} else if strings.HasPrefix(version, "==") {
		return "=="
	} else if strings.HasPrefix(version, "=") {
		return "="
	} else if strings.HasPrefix(version, ">") {
		return ">"
	} else if strings.HasPrefix(version, "<") {
		return "<"
	} else if strings.HasPrefix(version, "!=") {
		return "!="
	} else if strings.HasPrefix(version, "!") {
		return "!"
	}

	return ""
}
