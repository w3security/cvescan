package binary

import (
	"context"
	"errors"
	"os"

	"golang.org/x/xerrors"

	"github.com/aquasecurity/go-dep-parser/pkg/rust/binary"
	"github.com/w3security/cvescan/pkg/fanal/analyzer"
	"github.com/w3security/cvescan/pkg/fanal/analyzer/language"
	"github.com/w3security/cvescan/pkg/fanal/types"
	"github.com/w3security/cvescan/pkg/fanal/utils"
)

func init() {
	analyzer.RegisterAnalyzer(&rustBinaryLibraryAnalyzer{})
}

const version = 1

type rustBinaryLibraryAnalyzer struct{}

func (a rustBinaryLibraryAnalyzer) Analyze(_ context.Context, input analyzer.AnalysisInput) (*analyzer.AnalysisResult, error) {
	p := binary.NewParser()
	libs, deps, err := p.Parse(input.Content)
	if errors.Is(err, binary.ErrUnrecognizedExe) || errors.Is(err, binary.ErrNonRustBinary) {
		return nil, nil
	} else if err != nil {
		return nil, xerrors.Errorf("rust binary parse error: %w", err)
	}

	return language.ToAnalysisResult(types.RustBinary, input.FilePath, "", libs, deps), nil
}

func (a rustBinaryLibraryAnalyzer) Required(_ string, fileInfo os.FileInfo) bool {
	return utils.IsExecutable(fileInfo)
}

func (a rustBinaryLibraryAnalyzer) Type() analyzer.Type {
	return analyzer.TypeRustBinary
}

func (a rustBinaryLibraryAnalyzer) Version() int {
	return version
}
