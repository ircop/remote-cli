package remote_cli

// Here we will prepare various callbacks for handling pagination output.
// So lets add most recent callbacks.
func (c *Cli) preparePagination() {
	// Cisco ' --More-- ':
	c.CliHandler.RegisterCallback(`(?msi:^ --More-- )`, func(){ c.CliHandler.WriteRaw([]byte{' '}) })

	// Juniper '--more--'
	c.CliHandler.RegisterCallback(`(?msi:^---\(more.*?\)---)`, func() { c.CliHandler.WriteRaw([]byte{' '}) })

	// DLink pagination
	c.CliHandler.RegisterCallback(`(?msi:CTRL\+C.+?a A[Ll][Ll]\s*)`, func() { c.CliHandler.WriteRaw([]byte{'a'}) })

	// DLink CONTINUOUS pagination. It dont releases output before we will press some key(ESC, q...)
	// What can we do with this?... Probably the only way is:
	//
	// 1) match `CTRL+C ESC q Quit SPACE n Next Page p Previous Page r Refresh` string
	// 2) set some boolean variable, like 'paging', to true
	// 3) in sub-handlers, make some function like 'writeToBuffer(buffer)'. When in paging state, write output also to this buffer.
	//		It may be pointer to bytes.buffer, passed from this uplink.
	// 4) make some []string here in uplink. Every time when we are catching pagination:
	//		- if there is < 2 elements in []string, just add current buffer into it and flush buffer
	//		- check last 2 elements in this []string. If they are just similar to current, send 'q'
	//
	// TODO: think about it. Maybe just drop such 'paginations' and collect data by snmp?
}
