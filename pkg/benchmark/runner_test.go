package benchmark

import (
	"fmt"
	"testing"

	"github.com/kbelludoo/mbvm-db/pkg/mbvmdb"
	"github.com/kbelludoo/swebench-go/pkg/dataset"
)

func TestMBVMDB_Integration(t *testing.T) {
	suite := dataset.GetBenchmarkSuite()
	targets := map[string]string{
		"SWE-GO-001": "Execute",
		"SWE-GO-002": "ReadNext",
		"SWE-GO-003": "Allow",
	}

	for _, task := range suite {
		target := targets[task.ID]
		db := mbvmdb.New()
		err := db.IngestSource(task.SourceFile, task.RawSourceCode)
		if err != nil {
			t.Fatalf("IngestSource failed on %s: %v", task.ID, err)
		}

		slice := db.ExtractSlice(target, 2)
		proj := db.ProjectSlice(slice.Symbols)
		fmt.Printf("[%s] Target '%s': Raw=%d tok | PureMBVM=%d tok | MBVMDB=%d tok (Len=%d bytes)\n",
			task.ID, target, task.RawTokens, task.MBVMTokens, slice.TokensEstimate, len(proj))
		fmt.Println("--- Projected Slice ---")
		fmt.Println(proj)
		fmt.Println("-----------------------")
	}
}
