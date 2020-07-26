## Telchat

A simple chat server in Go connected via telnet.

### Optional Features Supported.

1. Messages segmented to rooms. Only clients connected in the same room
 receive the messages!
2. An HTTP REST API to post messages
3. An HTTP REST API to query for all messages.
4. Ignore option where a client can choose to ignore
   (unsubscribe) from another client's messages !
5. Allow option where a client can choose to allow the ignored
   (unsubscribe) client's messages again!



### Limitation

1. No persistence Message.
2. If Sending to any client get timed out messages are dropped.
3. Shutting down entire server closes all the Connection, but makes no efforts to check if there are any connection that needs to be drained.
4. Terminal needs to be VT-100 compatible for all text display. Most modern terminal is VT-100 supported.

### How to run.
1. Install `go` 1.13 at least. 

2. Run `go run cmd/main.go` to start with default configuration file.
If you want to run with you own configuration file, pass the config file location as flag for run command.

`go run main.go -config /tmp/config.json
`

config file takes below options.

```json
{
  "log_file": "./telchat.log",
  "telnet_addr": ":3001",
  "http_addr": ":3002"
}
```
a. *log_file* - location of file where logs should be stored.

b. *telnet_addr* - telnet server address to start. "ip:port"

c. *http_addr* - http server address for rest api. "ip:port"

3. Once the Server has started you can start connection to chat server using telnet.

```shell script
telnet 127.0.0.1 3001
```

OUTPUT
```shell script
>> telnet 127.0.0.1 3001
Trying 127.0.0.1...
Connected to 127.0.0.1.
Escape character is '^]'.
Hi There! Welcome to TELCHAT! Please Enter Your Chatter Name: 
>>Ankur
Thanks for Joining!. You can type /h for help anytime. Quick guide.

 SERIAL         COMMAND         OPTION          ARGS            DESCRIPTION
 ------         -------         ------          ----            -----------
 1              /info                                           display username & current room
 2              /room           change          [name]          join to [name] room
 3              /client         ignore          [name]          ignore [name] client's messages
 4              /client         allow           [name]          allow [name] client's messages

Examples

 1      /info
 2      /room change myroom3
 3      /client ignore annoyignone
 4      /client allow annoyignore

Send your typed message to the current room by entering enter
Ankur: [default] 

```
Everyone joins the default room on connection. You can switch to new room following the guide. 
You can also ask for the guide anytime during telnet connection by sending `/h`.

Once connected follow the instruction on screen to start using it.
To test open another terminal, and start another telnet connection.

Happy Chatting!.


### Rest API Guide.

1. query for all messages.

Method: `GET`

ENDPOINT: `/messages`

2. post messages

Method: `POST`

ENDPOINT: `/post`

Content-Type: `application/json`

PostBody: 
```json
{
    "name": "Ankur",
    "room": "default",
    "msg": "Hi There from browser"
}
```

## Watch the demo video for working demo.
`demo.mp4`