package remote_cli

import "fmt"

// Here we will prepare various callbacks for handling pagination output.
// So lets add most recent callbacks.
func (c *Cli) preparePagination() {
	// Cisco ' --More-- ':
	c.CliHandler.RegisterCallback(`(?msi:^ --More-- )`, func(){
			c.CliHandler.WriteRaw([]byte{' '})
		})

	// Juniper '--more--'
	c.CliHandler.RegisterCallback(`(?msi:^---\(more.*?\)---)`, func() {
			c.CliHandler.WriteRaw([]byte{' '})
		})

	if c.dlinkPagination {
		// DLink pagination
		c.CliHandler.RegisterCallback(`(?msi:CTRL\+C.+?a A[Ll][Ll]\s*)`, func() {
			fmt.Printf("MATCHED DLINK PAGING")
			c.CliHandler.WriteRaw([]byte{'a'})
		})

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

		// - make strings slice HERE
		// - after catching those - pagination - put LAST 5 LINES to this buffer
		// - etc, etc, etc
		// -- we need a way to get lower-level reader contents here, before return...
		//c.pagingBuf = make([]string, 0)
		c.CliHandler.RegisterCallback(`(?msi:CTRL\+C.+?[^\n]+Refresh([^\n]+)?\n$)`, c.continuousPager)
		// after first 'refresh ...', dlink doesnt show it more, but sends '0' character (32), so also try to match it
		//	c.CliHandler.RegisterCallback()

		// ^ This all does not work, because DLINK sends 'Refresh' string only once, and then re-sends all stuff except this 'refresh' string...
		// What we can do is to implement some shih in low-level CLIs, something like: enablePagerCatching()
	}
}

// Final version: this is dirty hack: we dont know how much pages are there, but we will just send 6 spaces and `quit`
func (c *Cli) continuousPager() {
	if c.paging {
		return
	}
	c.paging = true
	//fmt.Printf("MATCHED PAGING\n")
	//c.CliHandler.WriteRaw([]byte{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', 'q'})

	for i := 0; i < 6; i++ {
		c.CliHandler.WriteRaw([]byte{' '})
	}
	c.CliHandler.WriteRaw([]byte{'Q'})
}
