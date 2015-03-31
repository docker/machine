package share

import (
	"fmt"
	"syscall"

	dockerutils "github.com/docker/docker/utils"
	"github.com/docker/machine/drivers"
)

type ShareOptions struct {
	Name              string
	Type              string
	SrcPath, DestPath string
	SrcUid, DestUid   int
	SrcGid, DestGid   int
}

type Share interface {
	ContractFulfilled(d drivers.Driver) (bool, error)
	Create(d drivers.Driver) error
	Mount(d drivers.Driver) error
	Destroy(d drivers.Driver) error
	GetOptions() ShareOptions
}

type ShareWithType struct {
	Options ShareOptions
}

func ParseShares(shares []ShareWithType) []Share {
	parsedShares := []Share{}
	for _, s := range shares {
		parsedShares = append(parsedShares, NewShareWithOptions(s.Options))
	}
	return parsedShares
}

func NewShareWithOptions(options ShareOptions) Share {
	switch options.Type {
	case "vboxsf":
		return VBoxSharedFolder{
			Options: options,
		}
	case "nfs":
		return NfsSharedFolder{
			Options: options,
		}
	}
	return nil
}

func NewShare(shareType, absPath string) (Share, error) {
	switch shareType {
	case "vboxsf":
		return VBoxSharedFolder{
			Options: ShareOptions{
				Name:     dockerutils.GenerateRandomID(),
				SrcUid:   syscall.Getuid(),
				DestUid:  1000,
				DestGid:  50,
				SrcPath:  absPath,
				DestPath: absPath,
				Type:     "vboxsf",
			},
		}, nil
	case "nfs":
		return NfsSharedFolder{
			Options: ShareOptions{
				Name:     dockerutils.GenerateRandomID(),
				SrcUid:   syscall.Getuid(),
				DestUid:  1000,
				DestGid:  50,
				SrcPath:  absPath,
				DestPath: absPath,
				Type:     "nfs",
			},
		}, nil
	}
	return nil, fmt.Errorf("Driver type not recognized: %s", shareType)
}
