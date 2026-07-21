package submitter

import (
	"backend/auth"
	"backend/config"
	"bufio"
	"fmt"
	"net"
	"shared/golog"
)

var flagManager *FlagManager
var authManager *auth.AuthManager

var tcpLog = golog.New("TCP")

func Start(cfg *config.Config, fm *FlagManager, am *auth.AuthManager) {
	flagManager = fm
	authManager = am

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.SubmitterPort))
	if err != nil {
		panic(err)
	}

	tcpLog.Info("Listening on %v", listener.Addr())

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				continue
			}
			go handle(conn)
		}
	}()
}

func handle(conn net.Conn) {
	defer conn.Close()

	tcpLog.Info("Connection from %s", conn.RemoteAddr())

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	token, submitter, err := ReadAuth(reader)
	if err != nil {
		tcpLog.Error("Auth read error: " + err.Error())
		return
	}

	if authManager.GetSession(token) == "" {
		tcpLog.Error("Invalid token from " + conn.RemoteAddr().String())
		WriteAuthResponse(writer, false, "token expired or invalid, run 'marte login'")
		writer.Flush()
		return
	}

	WriteAuthResponse(writer, true, "")
	writer.Flush()

	data, err := ReadSubmitData(reader)
	if err != nil {
		tcpLog.Error("Read error: " + err.Error())
		return
	}

	tcpLog.Info("Received %d flags from %s", len(data.Flags), submitter)

	for _, fd := range data.Flags {
		f := Flag{
			Value:     fd.Value,
			TeamId:    fd.TargetTeam,
			Submitter: submitter,
			Service:   fd.Service,
		}
		flagManager.SubmitFlag(f)
	}
}
