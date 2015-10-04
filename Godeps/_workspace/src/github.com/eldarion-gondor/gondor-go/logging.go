package gondor

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
)

var reqWriteExcludeHeaderDump = map[string]bool{
	"Host":              true, // not in Header map anyway
	"Content-Length":    true,
	"Transfer-Encoding": true,
	"Trailer":           true,
}

func drainBody(b io.ReadCloser) (r1, r2 io.ReadCloser, err error) {
	var buf bytes.Buffer
	if _, err = buf.ReadFrom(b); err != nil {
		return nil, nil, err
	}
	if err = b.Close(); err != nil {
		return nil, nil, err
	}
	return ioutil.NopCloser(&buf), ioutil.NopCloser(bytes.NewReader(buf.Bytes())), nil
}

func valueOrDefault(value, def string) string {
	if value != "" {
		return value
	}
	return def
}

func (c *Client) logRequest(req *http.Request) error {
	if c.logHTTP {
		var err error
		body := true
		save := req.Body
		if !body || req.Body == nil {
			req.Body = nil
		} else {
			save, req.Body, err = drainBody(req.Body)
			if err != nil {
				return err
			}
		}
		fmt.Fprintln(os.Stderr, "----------- request start -----------")
		fmt.Fprintf(
			os.Stderr,
			"%s %s HTTP/%d.%d\r\n",
			valueOrDefault(req.Method, "GET"),
			req.URL.RequestURI(),
			req.ProtoMajor,
			req.ProtoMinor,
		)
		host := req.Host
		if host == "" && req.URL != nil {
			host = req.URL.Host
		}
		if host != "" {
			fmt.Fprintf(os.Stderr, "Host: %s\r\n", host)
		}
		chunked := len(req.TransferEncoding) > 0 && req.TransferEncoding[0] == "chunked"
		if len(req.TransferEncoding) > 0 {
			fmt.Fprintf(os.Stderr, "Transfer-Encoding: %s\r\n", strings.Join(req.TransferEncoding, ","))
		}
		if req.Close {
			fmt.Fprintf(os.Stderr, "Connection: close\r\n")
		}
		err = req.Header.WriteSubset(os.Stderr, reqWriteExcludeHeaderDump)
		if err != nil {
			return err
		}
		io.WriteString(os.Stderr, "\r\n")
		fmt.Fprintln(os.Stderr, "----------- body start -----------")
		if req.Body != nil {
			var dest io.Writer = os.Stderr
			if chunked {
				dest = httputil.NewChunkedWriter(dest)
			}
			_, err = io.Copy(dest, req.Body)
			if chunked {
				dest.(io.Closer).Close()
				io.WriteString(os.Stderr, "\r\n")
			}
		}
		fmt.Fprintln(os.Stderr, "----------- body end -----------")
		req.Body = save
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "----------- request end -----------")
	}
	return nil
}

var errNoBody = errors.New("sentinel error value")

type failureToReadBody struct{}

func (failureToReadBody) Read([]byte) (int, error) { return 0, errNoBody }
func (failureToReadBody) Close() error             { return nil }

var emptyBody = ioutil.NopCloser(strings.NewReader(""))

func (c *Client) logResponse(resp *http.Response) error {
	if c.logHTTP {
		var err error
		save := resp.Body
		savecl := resp.ContentLength
		body := true
		if !body {
			resp.Body = failureToReadBody{}
		} else if resp.Body == nil {
			resp.Body = emptyBody
		} else {
			save, resp.Body, err = drainBody(resp.Body)
			if err != nil {
				return err
			}
		}
		fmt.Println("----------- response start -----------")
		err = resp.Write(os.Stderr)
		if err == errNoBody {
			err = nil
		}
		resp.Body = save
		resp.ContentLength = savecl
		if err != nil {
			return err
		}
		fmt.Println("----------- response end -----------")
	}
	return nil
}
