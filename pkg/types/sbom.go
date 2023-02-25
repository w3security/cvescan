package types

import (
	stypes "github.com/spdx/tools-golang/spdx"

	"github.com/w3security/cvescan/pkg/fanal/types"
)

type SBOM struct {
	OS           types.OS
	Packages     []types.PackageInfo
	Applications []types.Application

	CycloneDX *types.CycloneDX
	SPDX      *stypes.Document2_2
}

type SBOMSource = string

const (
	SBOMSourceRekor = SBOMSource("rekor")
)

var (
	SBOMSources = []string{
		SBOMSourceRekor,
	}
)
