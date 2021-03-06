package httpmd5

import (
	"crypto/md5"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func makeTestHTTPServer() *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Test request parameters
		// Send response to be tested
		reqURL := req.URL.String()
		if reqURL == "/" || strings.HasPrefix(reqURL, "/path") {
			rw.WriteHeader(http.StatusOK)
			_, _ = rw.Write([]byte(reqURL))
			return
		}

		rw.WriteHeader(http.StatusNotFound)
		_, _ = rw.Write([]byte("not found"))
	}))

	return server
}

func makeMD5(s string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(s)))
}

func TestOneURL(t *testing.T) {
	// Close the server when test finishes
	httpServer := makeTestHTTPServer()
	defer httpServer.Close()

	httpMD5 := NewHTTPMD5(httpServer.Client(), time.Second)
	testURL := httpServer.URL + "/"
	md5s := httpMD5.GetMD5(1, []string{testURL})

	if len(md5s) != 1 {
		t.Fatal("one answer is expected")
	}

	if md5s[0].URL != testURL {
		t.Errorf("url is not correct: %v", md5s[0].URL)
	}

	if md5s[0].Err != nil || md5s[0].MD5 == nil {
		t.Error("error is not expected")
	}

	if *md5s[0].MD5 != makeMD5("/") {
		t.Error("md5 is not correct")
	}
}

func TestWrongURL(t *testing.T) {
	// Close the server when test finishes
	httpServer := makeTestHTTPServer()
	defer httpServer.Close()

	httpMD5 := NewHTTPMD5(httpServer.Client(), time.Second)
	testURL := "http://www.not-existing-host-anywhere"
	md5s := httpMD5.GetMD5(1, []string{testURL})

	if len(md5s) != 1 {
		t.Fatal("one answer is expected")
	}

	if md5s[0].URL != testURL {
		t.Errorf("url is not correct: %v", md5s[0].URL)
	}

	if md5s[0].Err == nil || md5s[0].MD5 != nil {
		t.Errorf("error is expected")
	}
}

func TestManyURLs(t *testing.T) {
	// Close the server when test finishes
	httpServer := makeTestHTTPServer()
	defer httpServer.Close()

	httpMD5 := NewHTTPMD5(httpServer.Client(), time.Second)

	testURLs := []string{}

	const numReqests = 1000

	for i := 0; i < numReqests; i++ {
		testURLs = append(testURLs, fmt.Sprintf("%s%s%d", httpServer.URL, "/path", i))
	}

	md5s := httpMD5.GetMD5(10, testURLs)

	if len(md5s) != numReqests {
		t.Fatalf("wrong number of responses: %v", len(md5s))
	}

	for _, urlMD5 := range md5s {
		if urlMD5.Err != nil || urlMD5.MD5 == nil {
			t.Fatal("error is not expected")
		}

		body := strings.TrimPrefix(urlMD5.URL, httpServer.URL)

		bodyMD5 := makeMD5(body)

		if *urlMD5.MD5 != bodyMD5 {
			t.Fatalf("md5 is not correct: %v", urlMD5)
		}
	}
}
