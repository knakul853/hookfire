package cli

// nearest returns the closest candidate to target by Levenshtein distance, and
// whether it is close enough to be worth suggesting (within half the target
// length, rounded up, plus one).
func nearest(target string, candidates []string) (string, bool) {
	best := ""
	bestDist := -1
	for _, c := range candidates {
		d := levenshtein(target, c)
		if bestDist == -1 || d < bestDist {
			bestDist, best = d, c
		}
	}
	if best == "" || bestDist > len(target)/2+1 {
		return "", false
	}
	return best, true
}

func levenshtein(a, b string) int {
	ar, br := []rune(a), []rune(b)
	prev := make([]int, len(br)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(ar); i++ {
		cur := make([]int, len(br)+1)
		cur[0] = i
		for j := 1; j <= len(br); j++ {
			cost := 1
			if ar[i-1] == br[j-1] {
				cost = 0
			}
			cur[j] = min3(cur[j-1]+1, prev[j]+1, prev[j-1]+cost)
		}
		prev = cur
	}
	return prev[len(br)]
}

func min3(a, b, c int) int {
	m := a
	if b < m {
		m = b
	}
	if c < m {
		m = c
	}
	return m
}
