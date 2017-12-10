package helpers

import "strings"

func IsVoxPopuli(author string) bool {
	extraVPNames := []string{
		"vox-populi",
		"rblogger",
		"vox.mens",
		"recenzent",
		"just-life",
		"vpodessa",
		"cyberanalytics",
		"poesie",
		"bizvoice",
		"fractal",
		"more-tsvetov",
		"ekomir",
		"iq4you",
		"digital-design",
		"cyber.events",
	}
	return strings.HasPrefix(author, "vp-") || Contains(extraVPNames, author)
}
