package email

import (
	"errors"
	"fmt"
	"net"
	"net/smtp"
	"regexp"
	"strings"
	"time"
)

const forceDisconnectAfter = time.Second * 3

// init
var (
	ErrBadFormat        = errors.New("invalid format")
	ErrUnresolvableHost = errors.New("unresolvable host")

	emailRegexp = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
)

// ValidateFormat check format
func ValidateFormat(email string) error {
	if !emailRegexp.MatchString(email) {
		return ErrBadFormat
	}
	return nil
}

// ValidateHost check host
func ValidateHost(email string) error {
	_, host := split(email)
	mx, err := net.LookupMX(host)
	if err != nil {
		return ErrUnresolvableHost
	}

	client, err := DialTimeout(fmt.Sprintf("%s:%d", mx[0].Host, 25), forceDisconnectAfter)
	if err != nil {
		return err
	}
	defer client.Close()

	err = client.Mail("check@mail.com")
	if err != nil {
		return err
	}
	err = client.Rcpt(email)
	if err != nil {
		return err
	}
	return nil
}

// DialTimeout returns a new Client connected to an SMTP server at addr.
// The addr must include a port, as in "mail.example.com:smtp".
func DialTimeout(addr string, timeout time.Duration) (*smtp.Client, error) {
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return nil, err
	}

	t := time.AfterFunc(timeout, func() { conn.Close() })
	defer t.Stop()

	host, _, _ := net.SplitHostPort(addr)
	return smtp.NewClient(conn, host)
}

func split(email string) (account, host string) {
	i := strings.LastIndexByte(email, '@')
	account = email[:i]
	host = email[i+1:]
	return
}
