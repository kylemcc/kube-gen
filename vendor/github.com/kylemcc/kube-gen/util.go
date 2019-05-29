package kubegen

func containsString(sl []string, v string) bool {
	for _, s := range sl {
		if v == s {
			return true
		}
	}
	return false
}
