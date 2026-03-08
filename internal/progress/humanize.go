package progress

import "fmt"

func humanizeBytes(n uint64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}

	value := float64(n)

	units := []string{"KiB", "MiB", "GiB", "TiB", "PiB"}
	for i := range units {
		value /= unit
		if value < unit || i == len(units)-1 {
			return fmt.Sprintf("%.2f %s", value, units[i])
		}
	}

	return fmt.Sprintf("%d B", n)
}
