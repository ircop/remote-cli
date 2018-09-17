package remote_cli

func (c *Cli) handleExtraChars(output string) string {

	bytes := []byte(output)
	newBytes := []byte{}

	for i := range bytes {
		// remove '\r''s
		if bytes[i] == 13 {
			continue
		}

		// find ^@ (\0 char) ; remove whole string before up to prev. \n
		//fmt.Printf("%d | %s\n", bytes[i], string(bytes[i]))
		if bytes[i] == 0 {
			//fmt.Printf("-- GOT CHAR ZERO --\n")
			if i == 1 {
				continue
			}
			upToLineBreak:
			for j := i-1; j > 0; j-- {
				//fmt.Printf("-- PREV CHAR = %d (%s)", bytes[j], string(bytes[j]))
				if bytes[j] != '\n' && bytes[j] != 0 {
					//remove 'j' byte
					newBytes = newBytes[:len(newBytes)-1]
				} else {
					break upToLineBreak
				}
			}
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
