package npm

import (
	"context"
	"os"
	"path/filepath"

	"golang.org/x/xerrors"

	"github.com/aquasecurity/go-dep-parser/pkg/nodejs/npm"
	"github.com/w3security/cvescan/pkg/fanal/analyzer"
	"github.com/w3security/cvescan/pkg/fanal/analyzer/language"
	"github.com/w3security/cvescan/pkg/fanal/types"
	"github.com/w3security/cvescan/pkg/fanal/utils"
)

func init() {
	analyzer.RegisterAnalyzer(&npmLibraryAnalyzer{})
}

const version = 1

var requiredFiles = []string{types.NpmPkgLock}

type npmLibraryAnalyzer struct{}

func (a npmLibraryAnalyzer) Analyze(_ context.Context, input analyzer.AnalysisInput) (*analyzer.AnalysisResult, error) {
	p := npm.NewParser()
	res, err := language.Analyze(types.Npm, input.FilePath, input.Content, p)
	if err != nil {
		return nil, xerrors.Errorf("unable to parse %s: %w", input.FilePath, err)
	}
	return res, nil
}

func (a npmLibraryAnalyzer) Required(filePath string, _ os.FileInfo) bool {
	fileName := filepath.Base(filePath)
	return utils.StringInSlice(fileName, requiredFiles)
}

func (a npmLibraryAnalyzer) Type() analyzer.Type {
	return analyzer.TypeNpmPkgLock
}

func (a npmLibraryAnalyzer) Version() int {
	return version
}
