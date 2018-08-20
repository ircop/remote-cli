package remote_cli

// Here we will prepare various callbacks for handling pagination output.
// So lets add most recent callbacks.
func (c *Cli) preparePagination() {
	// Cisco ' --More-- ':
	c.CliHandler.RegisterCallback(`(?msi:^ --More--)`, func(){
			c.CliHandler.WriteRaw([]byte{' '})
		})

	// Juniper '--more--'
	c.CliHandler.RegisterCallback(`(?msi:^---\(more.*?\)---)`, func() {
			c.CliHandler.WriteRaw([]byte{' '})
		})
}

// Final version: this is dirty hack: we dont know how much pages are there, but we will just send 6 spaces and `quit`
func (c *Cli) continuousPager() {
	if c.paging {
		return
	}
	c.paging = true

	for i := 0; i < 6; i++ {
		c.CliHandler.WriteRaw([]byte{' '})
	}
	c.CliHandler.WriteRaw([]byte{'Q'})
}

// Enable dlink pagination: it consumes much resources...
func (c *Cli) DlinkPagination() {
	c.CliHandler.RegisterCallback(`(?msi:CTRL\+C.+?a A[Ll][Ll]\s*)`, func() {
		c.CliHandler.WriteRaw([]byte{'a'})
	})
	c.CliHandler.RegisterCallback(`(?msi:CTRL\+C.+?[^\n]+Refresh([^\n]+)?\n$)`, c.continuousPager)
}
