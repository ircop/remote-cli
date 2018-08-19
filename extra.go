package remote_cli

func (c *Cli) handleExtraChars(output string) string {

	bytes := []byte(output)
	// find backspaces
	for i := range bytes {
		// remove THIS and PREVIOUS character
		// decrement i by 2 (2 chars removed)
		if bytes[i] == 8 && i > 1 {
			bytes = append(bytes[:i-1], bytes[i+1:]...)
			i -= 2
		}
	}

	return string(bytes)
}
