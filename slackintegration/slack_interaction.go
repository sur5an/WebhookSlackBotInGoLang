package slackintegration

import (
	"encoding/json"
	"fmt"
	"golang.org/x/net/websocket"
	"io/ioutil"
	"log"
	"net/http"
	"sync/atomic"
	"WebhookSlackBotInGoLang/utils"
)

type SlackClient struct {
	webSocket      *websocket.Conn
	memberChannels channelList
	myID           string
}

type responseRtmStart struct {
	Ok       bool         `json:"ok"`
	Error    string       `json:"error"`
	Url      string       `json:"url"`
	Self     responseSelf `json:"self"`
	Channels channelList  `json:"channels"`
	Groups   []groupList  `json:"groups"`
}

type responseSelf struct {
	Id string `json:"id"`
}

type groupList struct {
	Id      string      `json:"id"`
	Name    string      `json:"name"`
	Members utils.StringArray `json:"members"`
}

type channelDetails struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	IsMember bool   `json:"is_member"`
}

type channelList []channelDetails

type Message struct {
	Id      uint64 `json:"id"`
	Type    string `json:"type"`
	Channel string `json:"channel"`
	Text    string `json:"text"`
}

var counter uint64
var webSocketReceive = websocket.JSON.Receive
var webSocketSend = websocket.JSON.Send
var GetURL = http.Get
var ReadHttpBody = ioutil.ReadAll
var JsonUnMarshal = json.Unmarshal
var WSDial = websocket.Dial

func (channels channelList) find(channelName string) (channelToPublish string) {
	channelToPublish = channelName
	for _, channel := range channels {
		if channel.Name == channelName || channel.Id == channelName {
			channelToPublish = channel.Id
			break
		}
	}
	return
}

func (client SlackClient) ReceiveMessage() (message string, err error) {
	println("waiting for message")
	var m Message
	err = webSocketReceive(client.webSocket, &m)
	utils.FailOnError(err, "unable to receive message through web socket")
	message = m.Text
	println(m.Id, m.Type, m.Channel, m.Type)
	return
}

func (client SlackClient) SendMessage(messageToChannel string, channel string) (err error) {
	var message Message
	message.Id = atomic.AddUint64(&counter, 1)
	message.Channel = client.memberChannels.find(channel)
	message.Type = "message"
	message.Text = messageToChannel
	err = webSocketSend(client.webSocket, message)
	return err
}

func (client *SlackClient) Connect(token string) {
	log.Print("connecting to slack")
	url := fmt.Sprintf("https://slack.com/api/rtm.start?token=%s", token)

	resp, err := GetURL(url)
	if err != nil {
		utils.FailOnError(err, "error while connecting to slack")
	}
	if resp.StatusCode != 200 {
		err = fmt.Errorf("SlackAPI Failed with return code %d", resp.StatusCode)
		utils.FailOnError(err, "slack api returned non 200 result")
	}

	log.Print("getting details about bot")
	body, err := ReadHttpBody(resp.Body)
	resp.Body.Close()
	if err != nil {
		utils.FailOnError(err, "error on getting details about bot")
	}
	var responseObject responseRtmStart
	log.Print("parsing the details")
	err = JsonUnMarshal(body, &responseObject)
	if err != nil {
		utils.FailOnError(err, "error while parsing slack details")
	}
	if !responseObject.Ok {
		err = fmt.Errorf("SlackError: %s", responseObject.Error)
		utils.FailOnError(err, "error slack api return non ok for details")
	}
	client.myID = responseObject.Self.Id

	log.Print("getting the channel details")
	for _, element := range responseObject.Channels {
		if element.IsMember {
			client.memberChannels = append(client.memberChannels, element)
		}
	}

	for _, element := range responseObject.Groups {
		if element.Members.Contains(client.myID) {
			var group channelDetails
			group.Id = element.Id
			group.Name = element.Name
			group.IsMember = true
			client.memberChannels = append(client.memberChannels, group)
		}
	}
	client.webSocket, err = WSDial(responseObject.Url, "", "https://api.slack.com/")
	utils.FailOnError(err, "error while dialing to webscoket")
	return
}
