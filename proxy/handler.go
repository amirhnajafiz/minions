package proxy

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/amirhnajafiz/xerox/internal/metric"
	"go.uber.org/zap"
)

type ReqHandFunc func(w http.ResponseWriter, r *http.Request)

func HandleRequest(originServerURL *url.URL, logger *zap.Logger, mtc metric.Metrics) ReqHandFunc {
	// handle request method will return a proxy handler by forwarding our client
	return func(rw http.ResponseWriter, req *http.Request) {
		logger.Info("[reverse proxy server] received request", zap.Time("time", time.Now()))

		mtc.TotalRequests.Add(1)

		// set the parameters to forward our client to the main server
		req.Host = originServerURL.Host
		req.URL.Host = originServerURL.Host
		req.URL.Scheme = originServerURL.Scheme
		req.RequestURI = ""

		// supporting only http and https
		if req.URL.Scheme != "http" && req.URL.Scheme != "https" {
			msg := "unsupported protocol schema " + req.URL.Scheme

			http.Error(rw, msg, http.StatusBadRequest)
			logger.Error("unsupported protocol schema", zap.String("url", req.URL.Scheme))

			mtc.FailedRequests.Add(1)

			return
		}

		// deleting the hop to hop headers
		DeleteHopHeaders(req.Header)

		// appending host to x forward header in proxy server
		if clientIP, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
			AppendHostToXForwardHeader(req.Header, clientIP)
		}

		// send a request to the origin server
		originServerResponse, err := http.DefaultClient.Do(req)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)

			_, _ = fmt.Fprint(rw, err)

			mtc.FailedRequests.Add(1)

			return
		}

		// deleting the hop to hop headers
		DeleteHopHeaders(originServerResponse.Header)
		// adding the response headers from origin server
		CopyHeader(rw.Header(), originServerResponse.Header)

		mtc.SuccessfulRequests.Add(1)

		// return response to the client
		rw.WriteHeader(http.StatusOK)
		_, _ = io.Copy(rw, originServerResponse.Body)
	}
}
