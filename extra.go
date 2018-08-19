package remote_cli

func (c *Cli) handleExtraChars(output string) string {

	bytes := []byte(output)
	newBytes := []byte{}
	// find backspaces, remove them and prev.chars
	for i := range bytes {
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
