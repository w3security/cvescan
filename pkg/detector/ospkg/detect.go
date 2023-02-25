package ospkg

import (
	"time"

	"golang.org/x/xerrors"

	"github.com/w3security/cvescan/pkg/detector/ospkg/alma"
	"github.com/w3security/cvescan/pkg/detector/ospkg/alpine"
	"github.com/w3security/cvescan/pkg/detector/ospkg/amazon"
	"github.com/w3security/cvescan/pkg/detector/ospkg/debian"
	"github.com/w3security/cvescan/pkg/detector/ospkg/mariner"
	"github.com/w3security/cvescan/pkg/detector/ospkg/oracle"
	"github.com/w3security/cvescan/pkg/detector/ospkg/photon"
	"github.com/w3security/cvescan/pkg/detector/ospkg/redhat"
	"github.com/w3security/cvescan/pkg/detector/ospkg/rocky"
	"github.com/w3security/cvescan/pkg/detector/ospkg/suse"
	"github.com/w3security/cvescan/pkg/detector/ospkg/ubuntu"
	"github.com/w3security/cvescan/pkg/detector/ospkg/wolfi"
	fos "github.com/w3security/cvescan/pkg/fanal/analyzer/os"
	ftypes "github.com/w3security/cvescan/pkg/fanal/types"
	"github.com/w3security/cvescan/pkg/log"
	"github.com/w3security/cvescan/pkg/types"
)

var (
	// ErrUnsupportedOS defines error for unsupported OS
	ErrUnsupportedOS = xerrors.New("unsupported os")

	drivers = map[string]Driver{
		fos.Alpine:       alpine.NewScanner(),
		fos.Alma:         alma.NewScanner(),
		fos.Amazon:       amazon.NewScanner(),
		fos.CBLMariner:   mariner.NewScanner(),
		fos.Debian:       debian.NewScanner(),
		fos.Ubuntu:       ubuntu.NewScanner(),
		fos.RedHat:       redhat.NewScanner(),
		fos.CentOS:       redhat.NewScanner(),
		fos.Rocky:        rocky.NewScanner(),
		fos.Oracle:       oracle.NewScanner(),
		fos.OpenSUSELeap: suse.NewScanner(suse.OpenSUSE),
		fos.SLES:         suse.NewScanner(suse.SUSEEnterpriseLinux),
		fos.Photon:       photon.NewScanner(),
		fos.Wolfi:        wolfi.NewScanner(),
	}
)

// RegisterDriver is defined for extensibility and not supposed to be used in cvescan.
func RegisterDriver(name string, driver Driver) {
	drivers[name] = driver
}

// Driver defines operations for OS package scan
type Driver interface {
	Detect(string, *ftypes.Repository, []ftypes.Package) ([]types.DetectedVulnerability, error)
	IsSupportedVersion(string, string) bool
}

// Detector implements Operation
type Detector struct{}

// Detect detects the vulnerabilities
func (d Detector) Detect(_, osFamily, osName string, repo *ftypes.Repository, _ time.Time, pkgs []ftypes.Package) ([]types.DetectedVulnerability, bool, error) {
	driver, err := newDriver(osFamily)
	if err != nil {
		return nil, false, ErrUnsupportedOS
	}

	eosl := !driver.IsSupportedVersion(osFamily, osName)

	vulns, err := driver.Detect(osName, repo, pkgs)
	if err != nil {
		return nil, false, xerrors.Errorf("failed detection: %w", err)
	}

	return vulns, eosl, nil
}

func newDriver(osFamily string) (Driver, error) {
	if driver, ok := drivers[osFamily]; ok {
		return driver, nil
	}

	log.Logger.Warnf("unsupported os : %s", osFamily)
	return nil, ErrUnsupportedOS
}
