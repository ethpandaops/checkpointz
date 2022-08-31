package version

import (
	"fmt"
	"runtime"
)

var (
	Release   = "dev"
	GitCommit = "dev"
)

func Full() string {
	return fmt.Sprintf("Checkpointz/%s", Short())
}

func Short() string {
	return fmt.Sprintf("%s-%s", Release, GitCommit)
}

func FullVWithGOOS() string {
	return fmt.Sprintf("%s/%s", Full(), runtime.GOOS)
}
