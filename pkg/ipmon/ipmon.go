package ipmon

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// ServiceURL is the URL of the remote service that will return the external
// IP address. Any service can be used as long as the response body contains
// only a valid IP address.
var ServiceURL = "https://billglover-golang.appspot.com/ip"

// RequestTimeout is the timeout used when making a request to the external
// service.
var reqTimeout = 1 * time.Second

// GetExternalIP returns the external IP address or an error. The external IP
// is identified by making a call to an externally hosted service. This
// requires access to the internet.
func GetExternalIP() (net.IP, error) {

	req, err := http.NewRequest(http.MethodGet, ServiceURL, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}

	client := http.Client{Timeout: reqTimeout}
	res, err := client.Do(req)

	if res != nil {
		defer res.Body.Close()
	}

	if err != nil {
		return nil, errors.Wrap(err, "failed to make request")
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse response body")
	}

	ip := net.ParseIP(strings.TrimSpace(string(body)))
	if ip == nil {
		return nil, errors.New("no IP address returned")
	}

	return ip, nil
}

// WatchExternalIP starts monitoring the external IP address for changes. It
// takes a channel and uses it to return changes in the external IP address.
// It returns a cancel function that can be used to terminate the routine.
// The Go channel must be serviced, or the routine will hang.
func WatchExternalIP(ch chan<- net.IP) (context.CancelFunc, error) {
	ctx, cancel := context.WithCancel(context.Background())

	prevIP, err := GetExternalIP()
	if err != nil {
		return cancel, errors.Wrap(err, "unable to get current IP address")
	}

	go func() {
		for {
			t := time.NewTicker(10 * time.Second)
			select {
			case <-t.C:
				ip, err := GetExternalIP()
				if err != nil {
					log.Println(err)
				}
				if ip != nil && prevIP.Equal(ip) == false {
					prevIP = ip
					ch <- ip
				}

			case <-ctx.Done():
				fmt.Println("WatchExternalIP cancelled")
				return
			}
		}
	}()

	return cancel, nil
}
