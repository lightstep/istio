package lightstep

import (
	"fmt"
	"github.com/hashicorp/go-multierror"
	"net"
	"regexp"
	"strconv"
	"strings"
)

const (
	dns1123LabelMaxLength int    = 63
	dns1123LabelFmt       string = "[a-zA-Z0-9]([-a-z-A-Z0-9]*[a-zA-Z0-9])?"
)

var dns1123LabelRegexp = regexp.MustCompile("^" + dns1123LabelFmt + "$")

// ValidateServerAddress validates that the server address provided is valid
func ValidateServerAddress(addr string) error {
	if addr == "" {
		return fmt.Errorf("address cannot be empty")
	}
	return validateNetworkAddress(addr)
}

// ValidateServerAddress validates that the access token provided is valid
func ValidateAccessToken(token string) error {
	if token == "" {
		return fmt.Errorf("access token cannot be empty")
	}
	return nil
}

// ValidateSocketAddress validates that the socket address provided is valid
func ValidateSocketAddress(addr string) error {
	if addr == "" {
		return fmt.Errorf("address cannot be empty")
	}
	return validateNetworkAddress(addr)
}

// validateNetworkAddress checks that a network address is well-formed
// This function was copied (and renamed) from the istio project.
// https://github.com/istio/istio/blob/master/pilot/pkg/model/validation.go
func validateNetworkAddress(hostAddr string) error {
	host, p, err := net.SplitHostPort(hostAddr)
	if err != nil {
		return fmt.Errorf("unable to split %q: %v", hostAddr, err)
	}
	port, err := strconv.Atoi(p)
	if err != nil {
		return fmt.Errorf("port (%s) is not a number: %v", p, err)
	}
	if err = validatePort(port); err != nil {
		return err
	}
	if err = validateFQDN(host); err != nil {
		ip := net.ParseIP(host)
		if ip == nil {
			return fmt.Errorf("%q is not a valid hostname or an IP address", host)
		}
	}

	return nil
}

// validateFQDN checks a fully-qualified domain name
func validateFQDN(fqdn string) error {
	return AppendErrors(checkDNS1123Preconditions(fqdn), validateDNS1123Labels(fqdn))
}

// encapsulates DNS 1123 checks common to both wildcarded hosts and FQDNs
func checkDNS1123Preconditions(name string) error {
	if len(name) > 255 {
		return fmt.Errorf("domain name %q too long (max 255)", name)
	}
	if len(name) == 0 {
		return fmt.Errorf("empty domain name not allowed")
	}
	return nil
}
func validateDNS1123Labels(domain string) error {
	for _, label := range strings.Split(domain, ".") {
		if !isDNS1123Label(label) {
			return fmt.Errorf("domain name %q invalid (label %q invalid)", domain, label)
		}
	}
	return nil
}

// isDNS1123Label tests for a string that conforms to the definition of a label in
// DNS (RFC 1123).
func isDNS1123Label(value string) bool {
	return len(value) <= dns1123LabelMaxLength && dns1123LabelRegexp.MatchString(value)
}

// validatePort checks that the network port is in range
func validatePort(port int) error {
	if 1 <= port && port <= 65535 {
		return nil
	}
	return fmt.Errorf("port number %d must be in the range 1..65535", port)
}

// wrapper around multierror.Append that enforces the invariant that if all input errors are nil, the output
// error is nil (allowing validation without branching).
func AppendErrors(err error, errs ...error) error {
	appendError := func(err, err2 error) error {
		if err == nil {
			return err2
		} else if err2 == nil {
			return err
		}
		return multierror.Append(err, err2)
	}

	for _, err2 := range errs {
		err = appendError(err, err2)
	}
	return err
}
