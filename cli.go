package remote_cli

import (
	"github.com/ircop/tclient"
	"github.com/ircop/sshclient"
	"regexp"
	"github.com/pkg/errors"
	"fmt"
)

const (
	CliTypeTelnet = 1
	CliTypeSsh = 2
)

type CliInterface interface {
	Connect() error
	Close()
	GetLogin() string
	SetLogin(string)
	SetLoginPrompt(string)
	SetPasswordPrompt(string)
	SetPassword(string)
	SetPrompt(string)
	DisablePagination()
	Cmd(string) (string,error)
	ReadUntil(string) (string,error)
	RegisterCallback(string, func()) error
	RegisterErrorPattern(string,string) error
	GlobalTimeout(int)
	Write([]byte) error
	WriteRaw([]byte) error
	DlinkPagination()
}

// Cli is interface wrapper around telnet/ssh remote-cli subtypes.
type CliDummy interface {
	// ReadUntil reads data until given pattern matches
	ReadUntil(waitfor string) (string,error)
	// SetPrompt allows you to change prompt without re-creating ssh client
	SetPrompt(prompt string)
	// SetLoginPrompt
	SetLoginPrompt(prompt string)
	// SetLogin for cli instance
	SetLogin(login string)
	// You may need to change password for enable (because prompt is same as login)
	SetPassword(pw string)
	// Cmd is the same as ReadUntil, but pattern is default prompt, defined earlier
	Cmd(cmd string) (string, error)
	// RegisterCallback allows you to register some regex with callback, which will be called on regex match.
	// Useful for pagination on various devices.
	RegisterCallback(pattern string, callback func()) error
	// GlobalTimeout allows you to set global timeout value manually. Default is
	GlobalTimeout(t int)
	// Write sends your data to remote with adding \n or \r\n.
	Write(bytes []byte) error
	// WriteRaw sends your data to remote without adding \n or \r\n
	WriteRaw(bytes []byte) error
	// Open connection
	Open(host string, port int) error
	// Close closes the connection
	Close()
	// GetBuffer returns current buffer from reader as a string
	GetBuffer() string
}

type errorPattern struct {
	Re				*regexp.Regexp
	Description		string
}

// Cli struct
type Cli struct {
	ctype			int
	ip				string
	port			int
	login			string
	password		string
	prompt			string
	timeout			int
	globalTimeout	int

	pagination		bool
	paging			bool
	pagingBuf		[]string
	errorPatterns	[]errorPattern
	cache			map[string]string

	// CliHandler is downstream implementation for telnet/ssh communications.
	// All commands listed in CliDummy interface should be called via this CliHandler
	CliHandler		CliDummy
	connected		bool
}

// New creates new instance of CLI basing on CLI type given.
func New(cliType int, ip string, port int, login string, password string, prompt string, timeout int) *Cli {
	c := Cli{
		ip:ip,
		port:port,
		login:login,
		password:password,
		prompt:prompt,
		timeout:timeout,
		globalTimeout:timeout * 4,
		ctype:cliType,
		pagination:true,
		errorPatterns:make([]errorPattern, 0),
		cache:make(map[string]string),
	}

	if c.prompt == "" {
		c.prompt = `(?msi:[\$%#>]$)`
	}

	if c.port == 0 {
		if c.ctype == CliTypeSsh {
			c.port = 22
		}
		if c.ctype == CliTypeTelnet {
			c.port = 23
		}
	}
	if c.timeout == 0 {
		c.timeout = 3
	}

	if cliType == CliTypeTelnet {
		c.CliHandler = tclient.New(timeout, login, password, c.prompt)
	}
	if cliType == CliTypeSsh {
		c.CliHandler = sshclient.New(timeout, login, password, c.prompt)
	}

	return &c
}

// GetLogin
func (c *Cli) GetLogin() string {
	return c.login
}
// SetLogin
func (c *Cli) SetLogin(login string) {
	c.login = login
	c.CliHandler.SetLogin(login)
}

// SetLoginPrompt - change default login prompt
func (c *Cli) SetLoginPrompt(prompt string) {
	c.CliHandler.SetLoginPrompt(prompt)
	//c.RegisterCallback(prompt, func() {
	//	c.Write([]byte(c.login))
	//})
}

// SetPasswordPrompt - change default password prompt
func (c *Cli) SetPasswordPrompt(prompt string) {
	c.RegisterCallback(prompt, func() {
		c.Write([]byte(c.password))
	})
}

// Connect method calls dial methods from underlying cli implementations (telnet/ssh).
// Also here we registering pagination callbacks if it was not disabled earlier with DisablePagination.
func (c *Cli) Connect() error {
	err := c.CliHandler.Open(c.ip, c.port) ; if err != nil {
		return err
	} else {
		c.connected = true
		return nil
	}
}

// Close closes the connection if it's still opened
func (c *Cli) Close() {
	if c.connected {
		c.CliHandler.Close()
		c.connected = false
	}
}

// You may need to change password for enable (because prompt is same as login)
func (c *Cli) SetPassword(pw string) {
	c.CliHandler.SetPassword(pw)
}

// SetPrompt allows you to change prompt after initialization
func (c *Cli) SetPrompt(prompt string) {
	c.CliHandler.SetPrompt(prompt)
}

// DisablePagination disables pagination regexp checks, which are enabled by default (see preparePagination).
// With pagination checks (as well as with any regex callbacks) readers are trying to match every new recieved data
// with one or multiple regexps, which may consume additional resources.
// DisablePagination should be called before Connect() to take effect.
func (c *Cli) DisablePagination() {
	c.pagination = false
}


// These are just wrappers around underlying cli methods to avoid double class call hierarchy (i.e. we can
// call`cli.Cmd()` instead of `cli.CliHandler.Cmd()`). But are you still able to call underlying methods directly.
// Cmd sends given data and returns resulting output and/or error
func (c *Cli) Cmd(cmd string) (string, error) {

	if cached, ok := c.cache[cmd]; ok {
		return cached, nil
	}

	// set this every cmd call, not once
	c.GlobalTimeout(c.globalTimeout)

	c.paging = false
	result, err := c.CliHandler.Cmd(cmd)
	if err != nil {
		return result, err
	}

	result = c.handleExtraChars(result)

	for _, pattern := range c.errorPatterns {
		if pattern.Re.Match([]byte(result)) {
			return result, fmt.Errorf("Error: %s\n\n%s\n", pattern.Description, result)
		}
	}

	c.cache[cmd] = result

	return result, nil
}

// ReadUntil reads data until given pattern matched and returns result output and/or error.
func (c *Cli) ReadUntil(waitfor string) (string, error) {
	return c.CliHandler.ReadUntil(waitfor)
}


// RegisterCallback allows you to register some regex with callback, which will be called on regex match.
// Useful for pagination on various devices.
func (c *Cli) RegisterCallback(pattern string, callback func()) error {
	return c.CliHandler.RegisterCallback(pattern, callback)
}

// RegisterErrorPattern allows you to register regex pattern, which indicates error in final output.
// For example, something like '% Bad IP address or host name% Unknown command or computer name, or unable to find computer address' on Cisco
// Or 'Available commands' on DLink
// Patterns used only in Cmd()
func (c *Cli) RegisterErrorPattern(pattern string, description string) error {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return errors.Wrap(err, "RegisterErrorPattern: cannot compile pattern")
	}

	c.errorPatterns = append(c.errorPatterns, errorPattern{Re:re,Description:description})

	return nil
}

// GlobalTimeout allows you to set global timeout value manually. Default is
func (c *Cli) GlobalTimeout(t int) {
	c.CliHandler.GlobalTimeout(t)
}

// Write sends your data to remote with adding \n or \r\n.
func (c *Cli) Write(bytes []byte) error {
	return c.CliHandler.Write(bytes)
}
// WriteRaw sends your data to remote without adding \n or \r\n
func (c *Cli) WriteRaw(bytes []byte) error {
	return c.CliHandler.WriteRaw(bytes)
}
