package utils

func NormalizePage(page int) int {
	if page < 1 {
		return 1
	}

	return page
}

func NormalizeLimit(limit int) int {
	switch {
	case limit < 1:
		return 10
	case limit > 100:
		return 100
	default:
		return limit
	}
}
