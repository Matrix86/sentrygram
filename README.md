# SentryGram

SentryGram is a Telegram bot that allows you to execute commands on your machine/server or receive messages on events. It can load your custom JS plugins (it uses the JS interpreter https://github.com/robertkrimen/otto/) to exec commands received from a chat.
Only authorized users can use it and receive messages from it.

*It is still in its initial state so many functionalities should be added.*

## Installation
Make sure you have a GO >= 1.10 environment installed.

    $ go get github.com/Matrix86/sentrygram
    $ cd $GOPATH/src/github.com/Matrix86/sentrygram
    $ make
    $ make install

## Configuration file
The server needs a configuration file to work. On the configuration file you can specify the Telegram APIKEY, enable the users that can use the bot and so on.
The configuration file is a simple JSON file like below
```json
{
    "logs" : "",
    "tgm_api": "TELEGRAM APIKEY",
    "plugins_path": "plugins/",
    "users": ["userName"],
    "rpc_enabled": true,
    "rpc_port": 8082
}
```

* **logs**         : specifies the path of the output log (if empty it will use the stdout).
* **tgm_api**      : specifies the Telegram APIKEY (see https://core.telegram.org/bots#3-how-do-i-create-a-bot).
* **plugins_path** : specifies the directory path that contains the JS plugins.
* **users**        : array with all the usernames of users that can interact with the bot.
* **rpc_enabled**  : if true the server enables the rpc server and you can use the client to interact with the bot from your shell.
* **rpc_port**     : rpc port (default 8082).


## Client
If you need to send a message to an authorized user you could use the client program. For example, if a cronjob starts and you want to be alerted of this, you could call this exe to send you a message on the chat that you had previously opened with the bot.

    $ sentrygram_client -u "Username" -t "Text of the message"

## Plugins

The bot integrates a plugin system based on Javascript. It is possible to load one or more scripts with different functions in it.  
Each functions (first letter capitalized) defined in the JS scripts will enable a command on the Telegram bot. The functions exported
as commands will receive as argument the message read by the bot on private chat, group, etc.
Events like "message received on a group" or "message received on a private chat" call precise functions if they are defined on JS side.
The event functions are:

Event | Description
--- | ---
OnUserAddOnGroup | one or more users had been added to the group
OnGroupMessage | a message arrived on a group
OnChanMessage | a message arrived on channel
OnPrivateChatMessage | a message arrived on private chat
OnSuperGroupMessage | a message arrived on supergroup

All the events, as the command will receive an object with the following fields:

```
type BotMessage struct {
	From         string
	Content      string
	ChatName     string
	IsGroup      bool
	IsPrivate    bool
	IsSuperGroup bool
	IsChannel    bool
	IsCommand    bool
}
```

The commands are parsed and executed after the events. If a callback event returns false, the command parsing phase will be skipped.  
Only users defined in the configuration can execute commands.
On JS side there are some functions defined:

Function | Description
--- | ---
log(text string) | write a log line
readFile(path string, binarymode bool) | read a file on the system and return its content as binary or string
cpuUsage() | return an array with the usage of each cpu
getProcesses() | return an array of struct { Name string, Cpu  float64, Pid  int32, Mem  float64 } with all the processes
newBarGraph(title string, values []float64, labels []string) | return the path of an image of a bar graph
exec(cmd string) | exec a command on the system and return the result
sendMessage(to string, text string) | send a message to a user or group
sendImage(to string, path string) | send an image to a user or group
sendFile(to string, path string) | send a file to a user or group
sendAudio(to string, path string) | send an audio to a user or group
sendVideo(to string, path string) | send a video to a user or group
addBotAdmin(name string) | add a bot's administrator
getBotAdmins() | return a list with all the current defined administrators
kickUser(username string, group string) | kick a user from a group (or chan)
leaveGroup(group string) | the bot leaves a group
getChatMembersCount(group string) | returns the number of users in a group
