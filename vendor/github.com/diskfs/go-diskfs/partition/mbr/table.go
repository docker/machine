package mbr

import (
	"bytes"
	"fmt"
	"io"

	"github.com/diskfs/go-diskfs/util"
)

// Table represents an MBR partition table to be applied to a disk or read from a disk
type Table struct {
	Partitions         []*Partition
	LogicalSectorSize  int // logical size of a sector
	PhysicalSectorSize int // physical size of the sector
	initialized        bool
}

const (
	mbrSize               = 512
	partitionEntriesStart = 446
	partitionEntriesCount = 4
	signatureStart        = 510
)

// partitionEntrySize standard size of an MBR partition
const partitionEntrySize = 16

func getMbrSignature() []byte {
	return []byte{0x55, 0xaa}
}

// compare 2 partition arrays
func comparePartitionArray(p1, p2 []*Partition) bool {
	if (p1 == nil && p2 != nil) || (p2 == nil && p1 != nil) {
		return false
	}
	if p1 == nil && p2 == nil {
		return true
	}
	// neither is nil, so now we need to compare
	if len(p1) != len(p2) {
		return false
	}
	matches := true
	for i, p := range p1 {
		if p == nil && p2 != nil || !p.Equal(p2[i]) {
			matches = false
			break
		}
	}
	return matches
}

// ensure that a blank table is initialized
func (t *Table) initTable(size int64) {
	// default settings
	if t.LogicalSectorSize == 0 {
		t.LogicalSectorSize = 512
	}
	if t.PhysicalSectorSize == 0 {
		t.PhysicalSectorSize = 512
	}

	t.initialized = true
}

// Equal check if another table is equal to this one, ignoring CHS start and end for the partitions
func (t *Table) Equal(t2 *Table) bool {
	if t2 == nil {
		return false
	}
	// neither is nil, so now we need to compare
	basicMatch := t.LogicalSectorSize == t2.LogicalSectorSize &&
		t.PhysicalSectorSize == t2.PhysicalSectorSize
	partMatch := comparePartitionArray(t.Partitions, t2.Partitions)
	return basicMatch && partMatch
}

// tableFromBytes read a partition table from a byte slice
func tableFromBytes(b []byte, logicalBlockSize, physicalBlockSize int) (*Table, error) {
	// check length
	if len(b) != mbrSize {
		return nil, fmt.Errorf("Data for partition was %d bytes instead of expected %d", len(b), mbrSize)
	}
	mbrSignature := b[signatureStart:]

	// validate signature
	if bytes.Compare(mbrSignature, getMbrSignature()) != 0 {
		return nil, fmt.Errorf("Invalid MBR Signature %v", mbrSignature)
	}

	parts := make([]*Partition, 0, partitionEntriesCount)
	count := int(partitionEntriesCount)
	for i := 0; i < count; i++ {
		// write the primary partition entry
		start := partitionEntriesStart + i*partitionEntrySize
		end := start + partitionEntrySize
		p, err := partitionFromBytes(b[start:end])
		if err != nil {
			return nil, fmt.Errorf("Error reading partition entry %d: %v", i, err)
		}
		parts = append(parts, p)
	}

	table := &Table{
		Partitions:         parts,
		LogicalSectorSize:  512,
		PhysicalSectorSize: 512,
	}

	return table, nil
}

// Type report the type of table, always the string "mbr"
func (t *Table) Type() string {
	return "mbr"
}

// Read read a partition table from a disk, given the logical block size and physical block size
func Read(f util.File, logicalBlockSize, physicalBlockSize int) (*Table, error) {
	// read the data off of the disk
	b := make([]byte, mbrSize, mbrSize)
	read, err := f.ReadAt(b, 0)
	if err != nil {
		return nil, fmt.Errorf("Error reading MBR from file: %v", err)
	}
	if read != len(b) {
		return nil, fmt.Errorf("Read only %d bytes of MBR from file instead of expected %d", read, len(b))
	}
	return tableFromBytes(b, logicalBlockSize, physicalBlockSize)
}

// ToBytes convert Table to byte slice suitable to be flashed to a disk
// If successful, always will return a byte slice of size exactly 512
func (t *Table) toBytes() ([]byte, error) {
	b := make([]byte, 0, mbrSize-partitionEntriesStart)

	// write the partitions
	for i := 0; i < partitionEntriesCount; i++ {
		if i < len(t.Partitions) {
			btmp, err := t.Partitions[i].toBytes()
			if err != nil {
				return nil, fmt.Errorf("Could not prepare partition %d to write on disk: %v", i, err)
			}
			b = append(b, btmp...)
		} else {
			b = append(b, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}...)
		}
	}

	// signature
	b = append(b, getMbrSignature()...)
	return b, nil
}

// Write writes a given MBR Table to disk.
// Must be passed the util.File to write to and the size of the disk
func (t *Table) Write(f util.File, size int64) error {
	b, err := t.toBytes()
	if err != nil {
		return fmt.Errorf("Error preparing partition table for writing to disk: %v", err)
	}

	written, err := f.WriteAt(b, partitionEntriesStart)
	if err != nil {
		return fmt.Errorf("Error writing partition table to disk: %v", err)
	}
	if written != len(b) {
		return fmt.Errorf("Partition table wrote %d bytes to disk instead of the expected %d", written, len(b))
	}
	return nil
}

// GetPartitionSize returns the size in bytes of a single partition
func (t *Table) GetPartitionSize(partition int) (int64, error) {
	if partition > len(t.Partitions) {
		return 0, fmt.Errorf("Requested partition %d but only have %d partitions in table", partition, len(t.Partitions))
	}

	return int64(t.Partitions[partition-1].Size) * int64(t.LogicalSectorSize), nil
}

// GetPartitionStart returns the start position in bytes of a single partition
func (t *Table) GetPartitionStart(partition int) (int64, error) {
	if partition > len(t.Partitions) {
		return 0, fmt.Errorf("Requested partition %d but only have %d partitions in table", partition, len(t.Partitions))
	}

	return int64(t.Partitions[partition-1].Start) * int64(t.LogicalSectorSize), nil
}

// ReadPartitionContents read the entire contents of a partition into an io.Writer
//
// If successul, returns the number of bytes read.
// If the partition does not exist, returns an error.
func (t *Table) ReadPartitionContents(partition int, f util.File, writer io.Writer) (int64, error) {
	if partition > len(t.Partitions) {
		return 0, fmt.Errorf("Requested partition %d but only have %d partitions in table", partition, len(t.Partitions))
	}
	if partition < 1 {
		return 0, fmt.Errorf("Requested partition %d but first potential partition is %d", partition, 1)
	}
	logicalSectorSize := t.LogicalSectorSize
	if logicalSectorSize == 0 {
		logicalSectorSize = 512
	}
	physicalSectorSize := t.PhysicalSectorSize
	if physicalSectorSize == 0 {
		physicalSectorSize = 512
	}
	return t.Partitions[partition-1].readContents(f, logicalSectorSize, physicalSectorSize, writer)

}

// WritePartitionContents fill a partition with the available data from an io.Reader.
// This command is destructive; it will replace the contents of the partition with the contents read from the
// io.Reader.
//
// If successful, returns the number of bytes written.
// If the partition does not exist, returns an error.
func (t *Table) WritePartitionContents(partition int, f util.File, reader io.Reader) (uint64, error) {
	if partition > len(t.Partitions) {
		return 0, fmt.Errorf("Requested partition %d but only have %d partitions in table", partition, len(t.Partitions))
	}
	if partition < 1 {
		return 0, fmt.Errorf("Requested partition %d but first potential partition is %d", partition, 1)
	}
	logicalSectorSize := t.LogicalSectorSize
	if logicalSectorSize == 0 {
		logicalSectorSize = 512
	}
	physicalSectorSize := t.PhysicalSectorSize
	if physicalSectorSize == 0 {
		physicalSectorSize = 512
	}
	return t.Partitions[partition-1].writeContents(f, logicalSectorSize, physicalSectorSize, reader)
}
