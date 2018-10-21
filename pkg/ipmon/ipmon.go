package ipmon

import (
	"context"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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
	reqCount.Inc()

	req, err := http.NewRequest(http.MethodGet, ServiceURL, nil)
	if err != nil {
		reqErrCount.Inc()
		return nil, errors.Wrap(err, "failed to create request")
	}

	client := http.Client{Timeout: reqTimeout}

	roundTripper := promhttp.InstrumentRoundTripperDuration(reqDurations, http.DefaultTransport)
	client.Transport = roundTripper

	res, err := client.Do(req)

	if res != nil {
		defer res.Body.Close()
	}

	if err != nil {
		reqErrCount.Inc()
		return nil, errors.Wrap(err, "failed to make request")
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		reqErrCount.Inc()
		return nil, errors.Wrap(err, "failed to parse response body")
	}

	ip := net.ParseIP(strings.TrimSpace(string(body)))
	if ip == nil {
		reqErrCount.Inc()
		return nil, errors.New("no IP address returned")
	}

	return ip, nil
}

// WatchExternalIP starts monitoring the external IP address for changes. It
// takes a channel and a period and returns changes in the external IP address
// via the channel. It returns a cancel function that can be used to terminate
// the routine. The Go channel must be serviced, or the routine will hang.
func WatchExternalIP(ch chan<- net.IP, p time.Duration) (context.CancelFunc, error) {
	ctx, cancel := context.WithCancel(context.Background())

	prevIP, err := GetExternalIP()
	if err != nil {
		return cancel, errors.Wrap(err, "unable to get current IP address")
	}

	go func() {
		for {
			t := time.NewTicker(p)
			select {
			case <-t.C:
				ip, err := GetExternalIP()
				if err != nil {
					// If we are unable to get the external IP, we should log
					// it and continue to retry. Temporary network issues are
					// common so retrying is sensible.
					log.Println(err)
					break
				}
				if ip != nil && prevIP.Equal(ip) == false {
					prevIP = ip
					ch <- ip
				}

			case <-ctx.Done():
				return
			}
		}
	}()

	return cancel, nil
}
