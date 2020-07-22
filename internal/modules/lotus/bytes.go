package lotus

import (
	"strconv"
	"strings"
)

var byteSizeUnits = []string{"B", "KiB", "MiB", "GiB", "TiB", "PiB", "EiB", "ZiB"}

var deciUnits = []string{"", "Ki", "Mi", "Gi", "Ti", "Pi", "Ei", "Zi"}

func DescSizeStr(sizeStr string) uint64 {
	return DescSizeUnit(byteSizeUnits, sizeStr)
}

func DescDeciStr(deciStr string) uint64 {
	return DescSizeUnit(deciUnits, deciStr)
}

func DescSizeUnit(units []string, value string) uint64 {
	result := float64(1)
	for i, unit := range units {
		if i > 0 {
			result = result * 1024
			if strings.HasSuffix(value, unit) {
				value = strings.ReplaceAll(value, unit, "")
				vi, err := strconv.ParseFloat(value, 64)
				if err != nil {
					continue
				} else {
					result = result * vi
					break
				}
			}
		}
	}
	return uint64(result)
}
