package main

import (
	"flag"
	"fmt"
	"github.com/izgib/tttserver/game/interface/recorder/db"
	_ "github.com/jackc/pgx/v4/stdlib"
	"net"
	"time"

	"github.com/google/flatbuffers/go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	"github.com/izgib/tttserver/game/interface/rpc_service"
	"github.com/izgib/tttserver/game/interface/rpc_service/i9e"
	"github.com/izgib/tttserver/game/usecase"
	"github.com/izgib/tttserver/internal"
)

func main() {
	var host string
	var port int

	flag.StringVar(&host, "host", "0.0.0.0", "host to serve upon")
	flag.IntVar(&port, "port", 8080, "port to serve upon")
	flag.Parse()

	keepalivePolicy := keepalive.EnforcementPolicy{MinTime: 10 * time.Second}
	keepaliveParams := keepalive.ServerParameters{
		MaxConnectionIdle: 60 * time.Second,
		Time:              15 * time.Second,
	}

	log := internal.CreateDebugLogger()

	addr := fmt.Sprintf("%s:%d", host, port)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to start listening")
	}

	defer l.Close()

	s := grpc.NewServer(
		grpc.CustomCodec(flatbuffers.FlatbuffersCodec{}),
		grpc.KeepaliveEnforcementPolicy(keepalivePolicy),
		grpc.KeepaliveParams(keepaliveParams),
	)

	service := rpc_service.NewGameService(
		usecase.NewGameLobbyUsecase(
			db.NewGameRecorder,
			log,
		),
		log,
	)
	i9e.RegisterGameConfiguratorServer(s, service)

	log.Info().Msg("GameService started")

	if err = s.Serve(l); err != nil {
		log.Error().Err(err).Msg("failed to serve")
		return
	}
}
