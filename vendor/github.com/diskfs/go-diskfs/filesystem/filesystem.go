// Package filesystem provides interfaces and constants required for filesystem implementations.
// All interesting implementations are in subpackages, e.g. github.com/diskfs/go-diskfs/filesystem/fat32
package filesystem

import (
	"os"
)

// FileSystem is a reference to a single filesystem on a disk
type FileSystem interface {
	Type() Type
	Mkdir(string) error
	ReadDir(string) ([]os.FileInfo, error)
	OpenFile(string, int) (File, error)
}

// Type represents the type of disk this is
type Type int

const (
	// TypeFat32 is a FAT32 compatible filesystem
	TypeFat32 Type = iota
	// TypeISO9660 is an iso filesystem
	TypeISO9660
)
