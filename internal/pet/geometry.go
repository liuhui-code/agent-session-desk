package pet

func clampPosition(x, y, width, height, left, top, right, bottom int) (int, int) {
	return max(left, min(x, right-width)), max(top, min(y, bottom-height))
}
