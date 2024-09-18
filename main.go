package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/zeppelinmc/zeppelin/protocol/net"
	"github.com/zeppelinmc/zeppelin/protocol/net/io/encoding"
	"github.com/zeppelinmc/zeppelin/protocol/text"
	"github.com/zeppelinmc/zeppelin/server"
	"github.com/zeppelinmc/zeppelin/server/command"
	"github.com/zeppelinmc/zeppelin/util/log"
)

type Log struct {
	State, Packet int32
	Sender        bool //T:Client,F:Server
}

var logs = map[Log]struct{}{}

var dumper io.WriteCloser
var file *os.File

var ZeppelinPluginExport = server.Plugin{
	Identifier: "packetlog",
	Unload: func(p *server.Plugin) {
		dumper.Close()
		file.Close()
	},
	OnLoad: func(p *server.Plugin) {
		srv := p.Server()

		var err error
		file, err = os.OpenFile(p.Dir()+"/packets.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0755)
		if err != nil {
			log.Errorlnf("Error loading packetlog plugin: %v", err)
			return
		}

		srv.CommandManager.Register(packetlogCommand)
		log.Infoln("Packetlog plugin enabled")

		dumper = hex.Dumper(file)

		net.PacketWriteInterceptor = func(c *net.Conn, pk *bytes.Buffer, headerSize int32) (stop bool) {
			id, data, err := encoding.VarInt(pk.Bytes()[headerSize:])
			if err != nil {
				return
			}

			state := c.State()
			if _, ok := logs[Log{state, id, false}]; !ok {
				return
			}

			file.WriteString(fmt.Sprintf("%s Packet 0x%02x (clientbound):\n", time.Now(), id))

			dumper.Write(data)
			file.WriteString("\n\n\n")
			return
		}
	},
}

var packetlogCommand = command.Command{
	Namespace: "packetlog",
	Node: command.NewLiteral("packetlog",
		command.NewStringArgument("protocol", command.StringSingleWord,
			command.NewStringArgument("sender", command.StringSingleWord, command.NewIntegerArgument("id", nil, nil))),
	),
	Callback: func(ccc command.CommandCallContext) {
		if d := len(ccc.Arguments); d != 3 {
			ccc.Reply(text.Unmarshalf('&', "&cExpected 3 arguments, got %d", d))
			return
		}
		state, ok := states[ccc.Arguments[0]]
		if !ok {
			ccc.Reply(text.Unmarshal("&cInvalid protocol; expected: handshake|status|login|configuration|play", '&'))
			return
		}
		sender, ok := senders[ccc.Arguments[1]]
		if !ok {
			ccc.Reply(text.Unmarshal("&cInvalid sender; expected: server|client", '&'))
			return
		}
		packet, err := strconv.ParseInt(ccc.Arguments[2], 10, 32)
		if err != nil {
			ccc.Reply(text.Unmarshal("&cReceived invalid number for packet id", '&'))
			return
		}
		logs[Log{state, int32(packet), sender}] = struct{}{}
		ccc.Reply(text.Sprintf("Started logging packet 0x%02x", packet))
	},
}

var states = map[string]int32{
	"handshake":     0,
	"status":        1,
	"login":         2,
	"configuration": 3,
	"play":          4,
}
var senders = map[string]bool{
	"server": false,
	"client": true,
}
