package format

func ShortenNftId(id string) string {
	if len(id) <= 4 {
		return id
	}
	return id[0:4] + "..." + id[len(id)-4:]
}

func NftIdsToString(ids []string, separator string, shorten bool) string {
	var result string
	for _, id := range ids {
		if result != "" {
			result += separator
		}

		if shorten {
			result += ShortenNftId(id)
		} else {
			result += id
		}
	}
	return result
}
