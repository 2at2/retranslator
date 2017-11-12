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
	targetUrl  *url.URL
	forwardUri bool
	cl         *http.Client
	log        wd.Watchdog
}

func NewDeliver(targetUrl string, forwardUri bool) (*Deliver, error) {
	uri, err := url.Parse(targetUrl)

	if err != nil {
		return nil, err
	}

	return &Deliver{
		targetUrl:  uri,
		forwardUri: forwardUri,
		cl:         &http.Client{},
		log:        wd.NewLogger("deliver"),
	}, nil
}

func (d Deliver) BuildTargetUrl(uri string) *url.URL {

	if d.forwardUri { // return target url with injected uri
		targetUrl := *d.targetUrl
		uriParts := strings.SplitN(uri, "?", 2)
		targetUrl.Path = uriParts[0]
		if len(uriParts) > 1 {
			targetUrl.RawQuery = uriParts[1]
		} else {
			targetUrl.RawQuery = ""
		}

		return &targetUrl
	} else { // return unchanged target url
		urlCopy := *d.targetUrl
		// return copy to protect origin url from changes
		return &urlCopy
	}
}

func (d Deliver) Send(reqPacket retranslator.RequestPacket) (*retranslator.ResponsePacket, error) {

	targetUrl := d.BuildTargetUrl(reqPacket.RequestUri)
	d.log.Info("Request local url " + targetUrl.String())

	req := &http.Request{
		Method: reqPacket.Method,
		URL:    targetUrl,
		Body:   ioutil.NopCloser(bytes.NewReader(reqPacket.Body)),
		Header: make(http.Header),
	}

	if reqPacket.Headers != nil {
		for key, values := range reqPacket.Headers {
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

	return &retranslator.ResponsePacket{
		StatusCode: statusCode,
		Status:     status,
		Header:     header,
		Body:       body,
	}, nil
}
