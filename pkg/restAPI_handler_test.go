package pkg

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestRestAPIHandler_ServeHTTP(t *testing.T) {
	t.Parallel()
	file, err := ioutil.TempFile("", "telchat.*.log")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file.Name())
	readfile, err := os.OpenFile(file.Name(), os.O_RDONLY, 0644)
	if err != nil {
		t.Fatal(err)
	}
	rh := newRestAPIHandler(newMessageIO(file, readfile), newChatDataStore(ioutil.Discard))

	tcPostM := []struct {
		name    string
		req     *http.Request
		expCode int
	}{
		{
			name: "invalid body",
			req: httptest.NewRequest(http.MethodPost, "/post",
				ioutil.NopCloser(bytes.NewBuffer([]byte("")))),
			expCode: 400,
		},
		{
			name: "valid body",
			req: httptest.NewRequest(http.MethodPost, "/post",
				ioutil.NopCloser(bytes.NewBuffer(validReq))),
			expCode: 201,
		},
	}

	for _, tc := range tcPostM {
		t.Run(tc.name, func(t *testing.T) {
			rsp := httptest.NewRecorder()
			rh.postMessageHandler(rsp, tc.req)
			if rsp.Code != tc.expCode {
				t.Errorf("expected response code %d got %d", tc.expCode, rsp.Code)
			}
		})
	}

	// query messages
	req := httptest.NewRequest(http.MethodGet, "/messages", nil)
	rsp := httptest.NewRecorder()
	rh.messageHandler(rsp, req)
	if rsp.Code != 200 {
		t.Errorf("expected response code %d got %d", 200, rsp.Code)
	}
}

var validReq = []byte(`{
    "name": "Ankur",
    "room": "new",
    "msg": "Hi There from browser"
}`)
