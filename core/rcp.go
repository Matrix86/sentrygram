package core

import (
	"errors"
	"fmt"
	"net"
	"net/rpc"
	"strconv"

	"github.com/Matrix86/sentrygram/bot"
	"github.com/evilsocket/islazy/log"
)

type RpcResponse struct {
	Message string
}

type RpcRequest struct {
	Text     string
	UserName string
}

type Handler struct {
	bot *bot.Telegram
}

func (h *Handler) SendMessage(req RpcRequest, res *RpcResponse) (err error) {
	if req.UserName == "" {
		err = errors.New("A username should be specified")
		return
	}

	log.Info("Received SendMessage RPC [ %s : '%s']", req.UserName, req.Text)

	err = h.bot.SendMessage(req.UserName, req.Text)
	if err != nil {
		res.Message = fmt.Sprintf("Error: %s", err)
		log.Error(err.Error())
	} else {
		res.Message = "Sended"
	}
	return
}

func NewRpcHandler(port int, bot *bot.Telegram) {
	rpc.Register(&Handler{bot: bot})
	listener, _ := net.Listen("tcp", ":"+strconv.Itoa(port))
	defer listener.Close()

	rpc.Accept(listener)
}
