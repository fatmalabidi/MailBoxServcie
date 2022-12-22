package utils

import (
	"errors"
	"fmt"

	"github.com/fatmalabidi/MailBoxServcie/src/utils/helpers"
)

// TTPartition represents a test case for Partition function
type TTPartition struct {
	Name          string
	CollectionLen int
	PartitionSize int
	CheckResult   func(resultLength int, idxRange helpers.IdxRange) error
}

// CreateTTPartition generates different test cases for Partition function
func CreateTTPartition() []TTPartition {
	// check with  CollectionLen=100 and PartitionSize=10
	validateCount10 := func(resultLength int, idxRange helpers.IdxRange) error {
		if resultLength != 10 {
			return errors.New(fmt.Sprintf("Expected size: 10, got %d", resultLength))
		}
		return nil
	}
	// check with CollectionLen=100 and PartitionSize=15
	validateCount15 := func(resultLength int, idxRange helpers.IdxRange) error {
		if resultLength != 7 {
			return errors.New(fmt.Sprintf("Expected size: 7, got %d", resultLength))
		}
		return nil
	}
	// check with CollectionLen=0 and PartitionSize=15
	validateCollection0 := func(resultLength int, idxRange helpers.IdxRange) error {
		if idxRange.High != 0 || resultLength != 0 {
			return errors.New(fmt.Sprintf("Expected size: 0, got %d", resultLength))
		}
		return nil
	}
	// check with CollectionLen=10 and PartitionSize=0
	validatePartition0 := func(resultLength int, idxRange helpers.IdxRange) error {
		if (idxRange != helpers.IdxRange{}) {
			return errors.New(fmt.Sprintf("Expected size: 0, got %d", resultLength))
		}
		return nil
	}
	data := []TTPartition{
		{
			Name:          "Validate partitions counts",
			CollectionLen: 100,
			PartitionSize: 10,
			CheckResult:   validateCount10,
		},
		{
			Name:          "Validate partitions counts",
			CollectionLen: 100,
			PartitionSize: 15,
			CheckResult:   validateCount15,
		},
		{
			Name:          "Validate empty collection",
			CollectionLen: 0,
			PartitionSize: 15,
			CheckResult:   validateCollection0,
		},
		{
			Name:          "Validate empty Partition size",
			CollectionLen: 10,
			PartitionSize: 0,
			CheckResult:   validatePartition0,
		},
	}
	return data
}
