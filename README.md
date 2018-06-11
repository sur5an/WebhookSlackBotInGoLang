# WebhookSlackBotInGoLang
## What it does
This code tries to dequeue message from given rabbbitmq config and sends it thru SlackBot websocket

At the start the program it connects to the slack with the given bot token and loads all the channels that it is part of
Then it also start a go routine to dequeue messages from rabbitmq and sends the data thru channel
When msg is received in the other end of the channel the data is passed to slack websocket.

The message in the rabbitmq is expected to be in certain json format - which will help to find the intend slack channel and the person who is sending the data
Most of the other fields are not used as of now. 

## How to build and run
do make clean and then make 
edit config/slackbot.yam - add the slack token and rabbitmq details
run the binary SlackBot

## Whats next
add Unit test
expose an api to send the message to given channel or to person
add functionality to do interative bot
