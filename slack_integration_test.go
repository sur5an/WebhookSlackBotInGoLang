package main

import (
	"testing"
	"encoding/json"
	"golang.org/x/net/websocket"
	"fmt"
	"net/http"
	"bytes"
	"io"
)

func getTestSlackClient() (sc slackClient)  {
	ch := `[
               {"id": "1234", "name":"channelName", "is_member": true},
               {"id": "2345", "name":"myName", "is_member": false},
               {"id": "3456", "name":"SlackBot", "is_member": true},
               {"id": "4567", "name":"sur5an", "is_member": false}
           ]`
	cl := channelList{}
	json.Unmarshal([]byte(ch), &cl)
	sc.webSocket = nil
	sc.memberChannels = cl
	sc.myID = "1234"
	return
}

func TestFind(t *testing.T) {
	ch := `[
               {"id": "1234", "name":"channelName", "is_member": true},
               {"id": "2345", "name":"myName", "is_member": false},
               {"id": "3456", "name":"SlackBot", "is_member": true},
               {"id": "4567", "name":"sur5an", "is_member": false}
           ]`
	cl := channelList{}
	json.Unmarshal([]byte(ch), &cl)

	inputExpectedOutput := []struct {
		input          string
		expectedOutput string
	}{
		{"channelName", "1234"},
		{"myNam", "myNam"},
		{"sur5an", "4567"},
		{"SlackBot123", "SlackBot123"},
		{"^sur5an$", "^sur5an$"},
	}
	for _, element := range inputExpectedOutput {
		if cl.find(element.input) != element.expectedOutput {
			t.Errorf("Contains failed for %s unexpected output %s came", element.input, element.expectedOutput)
		}
	}
}

func TestReceiveMessage(t *testing.T) {
	oldWebSocketReceive := webSocketReceive
	defer func() {webSocketReceive = oldWebSocketReceive}()

	webSocketReceive = func(ws *websocket.Conn, v interface{}) (err error) {
		ch := `{"id": "1234", "type": "message", "channel": "channelName", "text": "sample message"}`
		m := v.(*Message)
		json.Unmarshal([]byte(ch), m)
		return nil
	}

	ch := `[
               {"id": "1234", "name":"channelName", "is_member": true},
               {"id": "2345", "name":"myName", "is_member": false},
               {"id": "3456", "name":"SlackBot", "is_member": true},
               {"id": "4567", "name":"sur5an", "is_member": false}
           ]`
	cl := channelList{}
	json.Unmarshal([]byte(ch), &cl)
	sc := getTestSlackClient()

	message, err := sc.receiveMessage()
	if err != nil || message != "sample message" {
		t.Errorf("unit test failed while getting message from slack mock")
	}
}

func TestSendMessage(t *testing.T) {
	oldWebSocketSend := webSocketSend
	defer func() { webSocketSend = oldWebSocketSend }()

	webSocketSend = func(ws *websocket.Conn, v interface{}) (err error) {
		var m Message
		data, payloadType, err := websocket.JSON.Marshal(v)
		websocket.JSON.Unmarshal(data,payloadType, &m)
		if m.Text != "test message" || m.Type != "message" || m.Channel != "1234" {
			err := fmt.Errorf("unit test for sendMessage fails")
			return err
		}
		return nil
	}

	sc := getTestSlackClient()

	err := sc.sendMessage("test message", "channelName")
	if err != nil {
		t.Errorf("unit test failed %s", err.Error())
	}

	err = sc.sendMessage("fail me", "unknownChannel")
	if err == nil {
		t.Errorf("unit test failed expected error message but found nothing")
	}
}
type nopCloser struct {
	io.Reader
}



func (nopCloser) Close() error { return nil }

func TestConnect(t *testing.T) {
	OldGetUrl := GetURL
	OldJsonUnMarshal := JsonUnMarshal
	OldWSDial := WSDial
	defer func(){GetURL = OldGetUrl;JsonUnMarshal=OldJsonUnMarshal;WSDial=OldWSDial}()
	sc := getTestSlackClient()

	GetURLPass := func(url string) (resp *http.Response, err error) {
		resp = &http.Response{}
		resp.StatusCode = 200
		resp.Body = nopCloser{bytes.NewBufferString("sample")}
		return
	}
	GetURL = GetURLPass

	WSDial = func(url_, protocol, origin string) (ws *websocket.Conn, err error) {
		return
	}

	JsonUnMarshal = func(data []byte, v interface{}) error {
		ch := `{"Ok": true, "Self": {"Id": "1234"}, "Channels": [
               {"id": "1234", "name":"channelName", "is_member": true},
               {"id": "2345", "name":"myName", "is_member": false},
               {"id": "3456", "name":"SlackBot", "is_member": true},
               {"id": "4567", "name":"sur5an", "is_member": false}
               ], "Groups": [
               {"id": "xyz", "name":"channel", "members": ["a", "b", "xyz"]},
               {"id": "wer", "name":"yourName", "members": ["a", "1"]},
               {"id": "qwer", "name":"BOT", "members": ["qwer", "1234", "sadasd"]},
               {"id": "45asd", "name":"suan", "members": ["asdas"]}
               ]}`
		m := v.(*responseRtmStart)
		json.Unmarshal([]byte(ch), m)
		return nil
	}

	sc.connect("1234")

	failOnError = func(err error, msg string) {
		if err == nil && msg != "error while dialing to webscoket" {
			t.Errorf("Failed in UT while testing negative case")
		}
	}

	GetURL = func(url string) (resp *http.Response, err error) {
		resp = &http.Response{}
		resp.StatusCode = 500
		resp.Body = nopCloser{bytes.NewBufferString("sample")}
		return
	}
	sc.connect("1234")


	GetURL = func(url string) (resp *http.Response, err error) {
		resp = &http.Response{}
		resp.Body = nopCloser{bytes.NewBufferString("sample")}
		err = fmt.Errorf("simulating error")
		return
	}
	sc.connect("1234")

	GetURL = GetURLPass

	JsonUnMarshal = func(data []byte, v interface{}) error {
		err := fmt.Errorf("simulating error")
		return err
	}

	sc.connect("1234")

	JsonUnMarshal = func(data []byte, v interface{}) error {
		ch := `{"Ok": false, "Self": {"Id": "1234"}, "Channels": [
               {"id": "1234", "name":"channelName", "is_member": true},
               {"id": "2345", "name":"myName", "is_member": false},
               {"id": "3456", "name":"SlackBot", "is_member": true},
               {"id": "4567", "name":"sur5an", "is_member": false}
               ], "Groups": [
               {"id": "xyz", "name":"channel", "members": ["a", "b", "xyz"]},
               {"id": "wer", "name":"yourName", "members": ["a", "1"]},
               {"id": "qwer", "name":"BOT", "members": ["qwer", "1234", "sadasd"]},
               {"id": "45asd", "name":"suan", "members": ["asdas"]}
               ]}`
		m := v.(*responseRtmStart)
		json.Unmarshal([]byte(ch), m)
		return nil
	}

	sc.connect("1234")

	ReadHttpBody = func(r io.Reader) ([]byte, error) {
		err := fmt.Errorf("simulating error")
		return nil, err
	}

	sc.connect("1234")
}
