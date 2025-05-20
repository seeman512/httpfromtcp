package main

import (
	"crypto/sha256"
	"fmt"
	"httpfromtcp/internal/headers"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"httpfromtcp/internal/server"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

const port = 8888 //42069

var template string = `<html>
  <head>
    <title>%s</title>
  </head>
  <body>
    <h1>%s</h1>
    <p>%s</p>
  </body>
</html>
`

func writeErr(w io.Writer, err error) {
	errorStr := err.Error()
	rw := response.NewWriter(w)
	inerr := rw.WriteStatusLine(response.SERVER_ERROR)
	if inerr != nil {
		fmt.Printf("Write status line error: %v\n", inerr)
		return
	}
	h := headers.NewHeaders()
	h.SetDefault(len(errorStr), nil)
	inerr = rw.WriteHeaders(h)
	if inerr != nil {
		fmt.Printf("Write headers error: %v\n", inerr)
		return
	}
	_, inerr = rw.WriteBody([]byte(errorStr))
	if inerr != nil {
		fmt.Printf("Write body error: %v\n", inerr)
		return
	}
}

func main() {

	handler := func(w io.Writer, r *request.Request) {
		rw := response.NewWriter(w)

		// chunked encoding
		prefix := "/httpbin/"
		if strings.HasPrefix(r.RequestLine.RequestTarget, prefix) {

			remote := fmt.Sprintf("https://httpbin.org/%s",
				strings.TrimPrefix(r.RequestLine.RequestTarget, prefix))
			fmt.Printf("REMOTE: %s\n", remote)
			res, err := http.Get(remote)
			if err != nil {
				writeErr(w, err)
				return
			}
			buf := make([]byte, 1024)

			rw.WriteStatusLine(response.OK)
			h := headers.NewHeaders()
			h.Set(map[string]string{
				"content-type":      "text/plain",
				"Transfer-Encoding": "chunked",
				"Trailer":           "X-Content-Sha256, X-Content-Length",
			})
			rw.WriteHeaders(h)

			cnt := 0
			body := []byte{}

			for {
				n, err := res.Body.Read(buf)
				cnt += n
				if err != nil {
					if err == io.EOF {
						rw.WriteChunkedBodyDone(true)
						break
					}
					fmt.Printf("Read error: %v\n", err)
					break
				}
				body = append(body, buf[:n]...)
				rw.WriteChunkedBody(buf[:n])
			}

			trailers := headers.NewHeaders()
			sum := sha256.Sum256(body)
			trailers.Set(map[string]string{
				"X-Content-Length": fmt.Sprintf("%d", cnt),
				"X-Content-Sha256": fmt.Sprintf("%x", sum),
			})
			rw.WriteTrailers(trailers)

			return
		}

		if strings.HasPrefix(r.RequestLine.RequestTarget, "/video") {
			body, err := os.ReadFile("./assets/vim.mp4")
			if err != nil {
				writeErr(w, err)
				return
			}

			rw.WriteStatusLine(response.OK)
			h := headers.NewHeaders()
			h.SetDefault(len(body), map[string]string{
				"content-type": "video/mp4",
			})
			rw.WriteHeaders(h)
			rw.WriteBody(body)
			return
		}

		statusCode := response.OK
		title := "200 OK"
		head := "Success!"
		msg := "Your request was an absolute banger."

		if r.RequestLine.RequestTarget == "/yourproblem" {
			statusCode = response.BAD_REQUEST
			title = "400 Bad Request"
			head = "Bad Request"
			msg = "Your request honestly kinda sucked."
		}

		if r.RequestLine.RequestTarget == "/myproblem" {
			statusCode = response.SERVER_ERROR
			title = "500 Internal Server Error"
			head = "Internal Server Error"
			msg = "Okay, you know what? This one is on me."
		}

		err := rw.WriteStatusLine(statusCode)
		if err != nil {
			fmt.Printf("Write status line error %v\n", err)
			return
		}

		bodyStr := fmt.Sprintf(template, title, head, msg)
		body := []byte(bodyStr)
		customHeaders := map[string]string{
			"content-type": "text/html",
		}

		h := headers.NewHeaders()
		h.SetDefault(len(body), customHeaders)
		err = rw.WriteHeaders(h)
		if err != nil {
			fmt.Printf("Write headers error %v\n", err)
			return
		}

		_, err = rw.WriteBody(body)
		if err != nil {
			fmt.Printf("Write body error %v\n", err)
			return
		}
	}
	server, err := server.Serv(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

