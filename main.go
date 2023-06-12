package main

import (
	"flag"
	"fmt"
	"github.com/izgib/tttserver/internal/logger"
	"github.com/izgib/tttserver/lobby"
	"github.com/izgib/tttserver/recorder"
	_ "github.com/jackc/pgx/v4/stdlib"
	//"github.com/rs/zerolog/log"
	"net"
	"os"
	"time"

	//"github.com/google/flatbuffers/go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	"github.com/izgib/tttserver/rpc_service"
	"github.com/izgib/tttserver/rpc_service/transport"
)

func main() {
	var host string
	var port int
	var http2Debug bool

	flag.StringVar(&host, "host", "0.0.0.0", "host to serve upon")
	flag.IntVar(&port, "port", 8080, "port to serve upon")
	flag.BoolVar(&http2Debug, "http2debug", false, "print http2 debug info")
	flag.Parse()
	log := logger.CreateDebugLogger()
	if http2Debug {
		log.Info().Msg("http/2 debug enabled")
		//os.Setenv("GRPC_GO_LOG_SEVERITY_LEVEL", "info")
		//os.Setenv("GRPC_GO_LOG_VERBOSITY_LEVEL", "2")
		os.Setenv("GRPC_TRACE", "all")
		os.Setenv("GRPC_VERBOSITY", "DEBUG")
		os.Setenv("GODEBUG", "http2debug=2")
	}

	keepalivePolicy := keepalive.EnforcementPolicy{MinTime: 10 * time.Second}
	keepaliveParams := keepalive.ServerParameters{
		MaxConnectionIdle: 60 * time.Second,
		Time:              15 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to start listening")
	}

	defer l.Close()

	s := grpc.NewServer(
		//grpc.CustomCodec(flatbuffers.FlatbuffersCodec{}),
		grpc.KeepaliveEnforcementPolicy(keepalivePolicy),
		grpc.KeepaliveParams(keepaliveParams),
	)

	service := rpc_service.NewGameService(
		lobby.NewGameLobbyController(
			recorder.NewGameRecorder,
			log,
		),
		log,
	)
	transport.RegisterGameConfiguratorServer(s, service)

	log.Info().Msg("GameService started")

	if err = s.Serve(l); err != nil {
		log.Error().Err(err).Msg("failed to serve")
		return
	}
}
