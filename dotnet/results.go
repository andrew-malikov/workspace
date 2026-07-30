package dotnet

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// TestSummary contains the aggregate counters emitted by dotnet test.
type TestSummary struct {
	Total   int
	Passed  int
	Failed  int
	Skipped int
}

type trxDocument struct {
	Summary trxResultSummary `xml:"ResultSummary"`
}

type trxResultSummary struct {
	Counters *trxCounters `xml:"Counters"`
}

type trxCounters struct {
	Total       int `xml:"total,attr"`
	Passed      int `xml:"passed,attr"`
	Failed      int `xml:"failed,attr"`
	NotExecuted int `xml:"notExecuted,attr"`
}

func loadTestSummary(resultsDir string) (TestSummary, error) {
	var paths []string
	if err := filepath.WalkDir(resultsDir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && strings.EqualFold(filepath.Ext(entry.Name()), ".trx") {
			paths = append(paths, path)
		}
		return nil
	}); err != nil {
		return TestSummary{}, fmt.Errorf("discover TRX results: %w", err)
	}
	if len(paths) == 0 {
		return TestSummary{}, errors.New("no TRX result files found")
	}

	var aggregate TestSummary
	for _, path := range paths {
		summary, err := loadTRXSummary(path)
		if err != nil {
			return TestSummary{}, err
		}
		aggregate.Total += summary.Total
		aggregate.Passed += summary.Passed
		aggregate.Failed += summary.Failed
		aggregate.Skipped += summary.Skipped
	}
	return aggregate, nil
}

func loadTRXSummary(path string) (TestSummary, error) {
	file, err := os.Open(path)
	if err != nil {
		return TestSummary{}, fmt.Errorf("open TRX result %s: %w", path, err)
	}
	defer file.Close()

	var document trxDocument
	if err := xml.NewDecoder(file).Decode(&document); err != nil {
		return TestSummary{}, fmt.Errorf("decode TRX result %s: %w", path, err)
	}
	if document.Summary.Counters == nil {
		return TestSummary{}, fmt.Errorf("TRX result %s has no result counters", path)
	}

	counters := document.Summary.Counters
	if counters.Total < 0 || counters.Passed < 0 || counters.Failed < 0 || counters.NotExecuted < 0 {
		return TestSummary{}, fmt.Errorf("TRX result %s has negative result counters", path)
	}
	return TestSummary{
		Total:   counters.Total,
		Passed:  counters.Passed,
		Failed:  counters.Failed,
		Skipped: counters.NotExecuted,
	}, nil
}
