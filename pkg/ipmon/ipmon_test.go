package ipmon

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetExternalIP(t *testing.T) {
	want := net.ParseIP("127.0.0.1")

	server := httptest.NewServer(http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		fmt.Fprintln(resp, "127.0.0.1")
	}))
	defer server.Close()
	ServiceURL = server.URL

	got, err := GetExternalIP()
	if err != nil {
		t.Errorf("unexpected error returned: %s", err)
	}

	if want.Equal(got) == false {
		t.Errorf("want: %s, got %s", want, got)
	}
}

func TestGetExternalIPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		fmt.Fprintln(resp, "unexpected response from remote service")
	}))
	defer server.Close()
	ServiceURL = server.URL

	ip, err := GetExternalIP()
	if err == nil {
		t.Error("expected an error to be returned, got none")
	}
	t.Log("ip:", ip)
}
