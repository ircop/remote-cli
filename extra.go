package remote_cli

func (c *Cli) handleExtraChars(output string) string {

	bytes := []byte(output)
	newBytes := []byte{}

	for i := range bytes {
		// remove '\r''s
		if bytes[i] == 13 {
			continue
		}
		// find backspaces, remove them and prev.chars
		if bytes[i] == 8 {
			if i == 1 {
				continue
			}
			newBytes = newBytes[:len(newBytes)-1]
			continue
		}
		newBytes = append(newBytes, bytes[i])
	}

	return string(newBytes)
}
