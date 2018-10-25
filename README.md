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
The bot integrates a plugin system using Javascript file as plugins. On JS side your script will receive the request and the response to modify. You can manipulate the received request as you want and to do this you can call several predefined Go functions like: log, readFile, cpuUsage, getProcesses, newBarGraph.
For an example of how use the plugin system you can see the `example.js` file on the directory plugins of this project.
