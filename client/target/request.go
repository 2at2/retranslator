package target

import (
	"bytes"
	"github.com/2at2/retranslator"
	"github.com/mono83/slf/wd"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type Deliver struct {
	target *url.URL
	cl     *http.Client
	log    wd.Watchdog
}

func NewDeliver(target string) (*Deliver, error) {
	uri, err := url.Parse(target)

	if err != nil {
		return nil, err
	}

	return &Deliver{
		target: uri,
		cl:     &http.Client{},
		log:    wd.NewLogger("deliver"),
	}, nil
}

func (d Deliver) Send(reqBody []byte, reqHeader map[string][]string) (*retranslator.Packet, error) {
	req := &http.Request{
		Method: "POST",
		URL:    d.target,
		Body:   ioutil.NopCloser(bytes.NewReader(reqBody)),
		Header: make(http.Header),
	}

	if reqHeader != nil {
		for key, values := range reqHeader {
			req.Header.Add(key, strings.Join(values, "\n"))
		}
	}

	response, err := d.cl.Do(req)

	var statusCode int
	var status string
	var header map[string][]string
	var body []byte

	if err != nil {
		d.log.Error("Failed target request - :err", wd.ErrParam(err))

		statusCode = http.StatusInternalServerError
		status = "Internal error"
	} else {
		body, err = ioutil.ReadAll(response.Body)

		if err != nil {
			d.log.Error("Unable to read body - :err", wd.ErrParam(err))

			statusCode = http.StatusInternalServerError
			status = "Internal error"
		} else {
			response.Body.Close()

			statusCode = response.StatusCode
			status = response.Status
			header = response.Header
		}
	}

	return &retranslator.Packet{
		StatusCode: statusCode,
		Status:     status,
		Header:     header,
		Body:       body,
	}, nil
}
