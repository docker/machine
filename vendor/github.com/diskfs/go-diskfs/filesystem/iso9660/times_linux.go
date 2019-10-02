package iso9660

import (
	"syscall"
	"time"
)

func getFileTimes(stat *syscall.Stat_t) (time.Time, time.Time) {
	return time.Unix(stat.Ctim.Sec, stat.Ctim.Nsec), time.Unix(stat.Atim.Sec, stat.Atim.Nsec)
}
