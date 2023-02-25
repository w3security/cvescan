package debian

import (
	"bufio"
	"context"
	"os"

	"golang.org/x/xerrors"

	"github.com/w3security/cvescan/pkg/fanal/analyzer"
	aos "github.com/w3security/cvescan/pkg/fanal/analyzer/os"

	"github.com/w3security/cvescan/pkg/fanal/types"
	"github.com/w3security/cvescan/pkg/fanal/utils"
)

func init() {
	analyzer.RegisterAnalyzer(&debianOSAnalyzer{})
}

const version = 1

var requiredFiles = []string{"etc/debian_version"}

type debianOSAnalyzer struct{}

func (a debianOSAnalyzer) Analyze(_ context.Context, input analyzer.AnalysisInput) (*analyzer.AnalysisResult, error) {
	scanner := bufio.NewScanner(input.Content)
	for scanner.Scan() {
		line := scanner.Text()
		return &analyzer.AnalysisResult{
			OS: types.OS{Family: aos.Debian, Name: line},
		}, nil
	}
	return nil, xerrors.Errorf("debian: %w", aos.AnalyzeOSError)
}

func (a debianOSAnalyzer) Required(filePath string, _ os.FileInfo) bool {
	return utils.StringInSlice(filePath, requiredFiles)
}

func (a debianOSAnalyzer) Type() analyzer.Type {
	return analyzer.TypeDebian
}

func (a debianOSAnalyzer) Version() int {
	return version
}
