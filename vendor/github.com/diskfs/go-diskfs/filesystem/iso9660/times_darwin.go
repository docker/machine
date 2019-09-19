package iso9660

import (
	"syscall"
	"time"
)

func getFileTimes(stat *syscall.Stat_t) (time.Time, time.Time) {
	return time.Unix(stat.Ctimespec.Sec, stat.Ctimespec.Nsec), time.Unix(stat.Atimespec.Sec, stat.Atimespec.Nsec)
}
