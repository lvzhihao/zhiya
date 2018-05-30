package goutils

import (
	"fmt"
	"testing"
)

func TestLog(t *testing.T) {
	logger := DefaultLogger()
	logger.Sync()
	logger.Info("test", "error", fmt.Errorf("test err"), "string", "abcdefg")
	logger.Set(&DefaultLogInstance{})
	logger.Info("test", "error", fmt.Errorf("test err"), "string", "abcdefg")
}
