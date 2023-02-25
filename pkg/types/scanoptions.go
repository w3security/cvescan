package types

import (
	"github.com/w3security/cvescan/pkg/fanal/types"
)

// ScanOptions holds the attributes for scanning vulnerabilities
type ScanOptions struct {
	VulnType            []string
	Scanners            Scanners
	ImageConfigScanners Scanners // Scanners for container image configuration
	ScanRemovedPackages bool
	Platform            string
	ListAllPackages     bool
	LicenseCategories   map[types.LicenseCategory][]string
	FilePatterns        []string
}
