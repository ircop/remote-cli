# remote-cli

An golang implementation of telnet/ssh cli client (underlying protocol clients implementations are [ircop/sshclient](https://github.com/ircop/sshclient) and [ircop/tclient](https://github.com/ircop/tclient)

`remote-cli` use them as interface implementations.

You can create callbacks based on output parsing patterns. If current output matches given pattern, callback will be called.

You can just Write() to remote (will be added '\n' or '\r\n'), WriteRaw() (nothing will be added).

You can set global timeout, which will break all operations (i.e. if you pattern does not match output and all stucks). Default is 3 * network timeout.

You can ReadUntil(pattern string) - read incoming data until it matched your pattern or until timeout reached.

You can call Cmd(string) - it will send your data to remote and read output until it matched default pattern.


**Note:** this tool also strips out all escape sequences (like colors), because they often makes impossible to parse output with scripts.

# Some examples


## Telnet D-Link switch. "show switch" has too long output and there is pagination, but we will omit this.

```
import (
	"github.com/ircop/remote-cli"
	"fmt"
)

func main() {
  cli := remote_cli.New(remote_cli.CliTypeTelnet, "1.1.1.1", 23, "script", "xxxxxx", `(?msi:[\$%#>]$)`, 3)
  
  err := cli.Connect()
	if err != nil {
		panic(err)
	}
  
  err := cli.Connect()
	if err != nil {
		panic(err)
	}
	
	out, err := cli.Cmd("show switch")
	if err != nil {
		panic(err)
	}
	fmt.Printf(out)
}
```

Output:

![Output](https://i.imgur.com/pUC97PI.png)


## Ssh to Cisco switch and execute command with large output. Register our own pattern for omitting pagination.

```
  cli := remote_cli.New(remote_cli.CliTypeSsh, "20.20.20.20", 22, "script", "xxxxx", `(?msi:^\S+?>)`, 3)

  err := cli.Connect()
	if err != nil {
		panic(err)
	}
  
  cli.RegisterCallback(`(?msi:^ --More-- )`, func(){ cli.WriteRaw([]byte{' '}) })

  out, err := cli.Cmd("show int status")
	if err != nil {
		fmt.Printf(out)
		panic(err)
	}
	fmt.Printf(out)
```

Output:

![Output](https://i.imgur.com/EB4nozo.png)


## Ssh to linux host and execute some command

```
	cli := remote_cli.New(remote_cli.CliTypeSsh, "10.10.10.10", 22, "login", "password", `(?msi:(~ \$|#)\s+$)`, 3)

	err := cli.Connect()
	if err != nil {
		panic(err)
	}
	
	out, err := cli.Cmd("ls /var -l")
	if err != nil {
		fmt.Printf(out)
		panic(err)
	}
	fmt.Printf("%s\n", out)
```

Output:

![Output](https://i.imgur.com/ypW108c.png)

