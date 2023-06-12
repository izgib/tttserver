package rpc_service

import (
	"context"
	"fmt"
	"github.com/izgib/tttserver/game"
	"github.com/izgib/tttserver/internal/logger"
	"github.com/izgib/tttserver/lobby"
	"github.com/rs/zerolog/log"
	"github.com/sirkon/errors"
	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/reflect/protoreflect"
	"io"
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/test/bufconn"

	"github.com/izgib/tttserver/rpc_service/transport"
)

const bufSize = 1024 * 16

type gameLobbyID struct {
	id uint32
}

func (g gameLobbyID) ID() uint32 {
	return g.id
}

type lobbyParams struct {
	settings game.GameSettings
	mark     game.PlayerMark
}

var keepalivePolicy = keepalive.EnforcementPolicy{MinTime: 10 * time.Second}
var keepaliveParams = keepalive.ServerParameters{
	MaxConnectionIdle: 60 * time.Second,
	Time:              15 * time.Second,
}

var serveOpts = []grpc.ServerOption{
	//grpc.CustomCodec(flatbuffers.FlatbuffersCodec{}),
	grpc.KeepaliveEnforcementPolicy(keepalivePolicy),
	grpc.KeepaliveParams(keepaliveParams)}

type PlayerParams struct {
	turn  int
	moves []game.Move
	mark  game.PlayerMark
}

func TestGameService(t *testing.T) {
	var testLobbies = []lobbyParams{
		{
			game.GameSettings{Rows: 3, Cols: 3, Win: 3},
			game.CrossMark,
		},
		{
			settings: game.GameSettings{Rows: 3, Cols: 3, Win: 3},
			mark:     game.NoughtMark,
		},
	}
	timeout := 500 * time.Millisecond
	//timeout := 300 * time.Second
	crLeaveRequest := transport.CreateRequest{
		Payload: &transport.CreateRequest_Action{Action: transport.ClientAction_CLIENT_ACTION_LEAVE},
	}
	oppLeaveRequest := transport.JoinRequest{
		Payload: &transport.JoinRequest_Action{Action: transport.ClientAction_CLIENT_ACTION_LEAVE},
	}

	t.Run("success: Tie", func(t *testing.T) {
		quit := make(chan struct{})
		testLobby := testLobbies[1]
		conn, err, errc := startupServer(quit)
		if err != nil {
			t.Fatal(err)
		}
		go func() {
			t.Fatal(<-errc)
		}()

		defer func() {
			close(quit)
		}()

		moves := []game.Move{{1, 1}, {0, 2}, {2, 2}, {0, 0}, {0, 1}, {2, 1}, {1, 0}, {1, 2}, {2, 0}}
		wantRes := transport.GameStatus_GAME_STATUS_TIE

		creator := transport.NewGameConfiguratorClient(conn)
		crLogger := logger.CreateDebugLogger().With().Str("player", game.EnumNamesMoveChoice[game.MoveChoice(testLobby.mark)]).Logger()
		crContext, _ := context.WithTimeout(context.Background(), timeout)

		opponent := transport.NewGameConfiguratorClient(conn)
		oppMark := (testLobby.mark + 1) & 1
		oppLogger := logger.CreateDebugLogger().With().Str("player", game.EnumNamesMoveChoice[game.MoveChoice(oppMark)]).Logger()
		oppContext, _ := context.WithTimeout(context.Background(), timeout)

		crStream, err := creator.CreateGame(crContext)
		if err != nil {
			t.Fatal(err)
		}

		oppStream, err := opponent.JoinGame(oppContext)
		if err != nil {
			t.Fatal(err)
		}

		ctx, _ := context.WithTimeout(context.Background(), timeout)
		group, _ := errgroup.WithContext(ctx)
		idChan := make(chan lobby.GameLobbyID)
		group.Go(func() error {
			id, crErr := gameCreatorInit(crStream, testLobby.settings, testLobby.mark, crLogger)
			if crErr != nil {
				return crErr
			}
			idChan <- id

			if crAwaitErr := gameCreatorStartAwait(crStream, crLogger); crAwaitErr != nil {
				return crAwaitErr
			}

			crLogger.Debug().Msg("started")
			gameRes, gameErr := gameCrInteractor(crStream, PlayerParams{moves: moves, mark: testLobby.mark}, crLogger)
			if gameErr != nil {
				return gameErr
			}
			assert.Equal(t, wantRes, gameRes)
			return nil
		})

		group.Go(func() error {
			oppLogger.Debug().Msg("before connection")
			if err := gameOpponentInit(oppStream, (<-idChan).ID(), oppLogger); err != nil {
				return err
			}
			oppLogger.Debug().Msg("started")
			gameRes, gameErr := gameOppInteractor(oppStream, PlayerParams{moves: moves, mark: oppMark}, oppLogger)
			if gameErr != nil {
				return gameErr
			}
			assert.Equal(t, wantRes, gameRes)
			return nil
		})

		if err = group.Wait(); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("success: Win-X(Creator GiveUp)", func(t *testing.T) {
		quit := make(chan struct{})
		testLobby := testLobbies[1]
		conn, err, errc := startupServer(quit)
		if err != nil {
			t.Fatal(err)
		}
		go func() {
			t.Fatal(<-errc)
		}()

		defer func() {
			close(quit)
		}()
		moves := []game.Move{{1, 1}, {1, 2}, {2, 2}, {0, 0}}
		oppMark := (testLobby.mark + 1) & 1
		wantRes := transport.WinLine{Mark: enumMarkType[oppMark], Start: nil, End: nil}
		creator := transport.NewGameConfiguratorClient(conn)
		crLogger := logger.CreateDebugLogger().With().Str("player", game.EnumNamesMoveChoice[game.MoveChoice(testLobby.mark)]).Logger()
		crContext, _ := context.WithTimeout(context.Background(), timeout)

		opponent := transport.NewGameConfiguratorClient(conn)
		oppLogger := logger.CreateDebugLogger().With().Str("player", game.EnumNamesMoveChoice[game.MoveChoice(oppMark)]).Logger()
		oppContext, _ := context.WithTimeout(context.Background(), timeout)

		crStream, err := creator.CreateGame(crContext)
		if err != nil {
			t.Fatal(err)
		}

		oppStream, err := opponent.JoinGame(oppContext)
		if err != nil {
			t.Fatal(err)
		}

		ctx, _ := context.WithTimeout(context.Background(), timeout)
		group, _ := errgroup.WithContext(ctx)
		idChan := make(chan lobby.GameLobbyID)
		group.Go(func() error {
			id, crErr := gameCreatorInit(crStream, testLobby.settings, testLobby.mark, crLogger)
			if crErr != nil {
				return crErr
			}
			idChan <- id

			if crAwaitErr := gameCreatorStartAwait(crStream, crLogger); crAwaitErr != nil {
				return crAwaitErr
			}

			crLogger.Debug().Msg("started")
			//res, gameErr := gameCrInteractor(crStream, append(moves, game.Move{I: 0, J: 0}), testLobby.mark, crLogger)
			_, gameErr := gameCrInteractor(crStream, PlayerParams{moves: moves[:2], mark: testLobby.mark}, crLogger)
			if gameErr != nil {
				return gameErr
			}
			crGiveUpRequest := transport.CreateRequest{
				Payload: &transport.CreateRequest_Action{Action: transport.ClientAction_CLIENT_ACTION_GIVE_UP},
			}
			if gameErr = crStream.Send(&crGiveUpRequest); gameErr != nil {
				return gameErr
			}
			res, gameErr := gameCrInteractor(crStream, PlayerParams{turn: 2, moves: moves[2:], mark: testLobby.mark}, crLogger)
			if gameErr != nil {
				return gameErr
			}
			gameRes := res.(*transport.WinLine)
			assert.Equal(t, wantRes.Mark, gameRes.Mark)
			assert.Equal(t, wantRes.Start, gameRes.Start)
			assert.Equal(t, wantRes.End, gameRes.End)
			crLogger.Debug().Msg("finished execution")
			return nil
		})

		group.Go(func() error {
			oppLogger.Debug().Msg("before connection")
			if err := gameOpponentInit(oppStream, (<-idChan).ID(), oppLogger); err != nil {
				return err
			}
			oppLogger.Debug().Msg("started")
			res, gameErr := gameOppInteractor(oppStream, PlayerParams{moves: moves, mark: oppMark}, oppLogger)
			if gameErr != nil {
				return gameErr
			}
			gameRes := res.(*transport.WinLine)
			assert.Equal(t, wantRes.Mark, gameRes.Mark)
			assert.Equal(t, wantRes.Start, gameRes.Start)
			assert.Equal(t, wantRes.End, gameRes.End)
			oppLogger.Debug().Msg("finished execution")
			return nil
		})

		if err = group.Wait(); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("success: Tie -> Restart -> Leave", func(t *testing.T) {
		quit := make(chan struct{})
		testLobby := testLobbies[1]
		conn, oppErr, errc := startupServer(quit)
		if oppErr != nil {
			t.Fatal(oppErr)
		}
		go func() {
			t.Fatal(<-errc)
		}()

		defer func() {
			close(quit)
		}()

		tieMoves := []game.Move{{1, 1}, {0, 2}, {2, 2}, {0, 0}, {0, 1}, {2, 1}, {1, 0}, {1, 2}, {2, 0}}
		partGameMoves := tieMoves[:5]
		wantRes := transport.GameStatus_GAME_STATUS_TIE

		creator := transport.NewGameConfiguratorClient(conn)
		crLogger := logger.CreateDebugLogger().With().Str("player", game.EnumNamesMoveChoice[game.MoveChoice(testLobby.mark)]).Logger()
		crContext, _ := context.WithTimeout(context.Background(), timeout)

		opponent := transport.NewGameConfiguratorClient(conn)
		oppMark := (testLobby.mark + 1) & 1
		oppLogger := logger.CreateDebugLogger().With().Str("player", game.EnumNamesMoveChoice[game.MoveChoice(oppMark)]).Logger()
		oppContext, _ := context.WithTimeout(context.Background(), timeout)

		crStream, oppErr := creator.CreateGame(crContext)
		if oppErr != nil {
			t.Fatal(oppErr)
		}

		oppStream, oppErr := opponent.JoinGame(oppContext)
		if oppErr != nil {
			t.Fatal(oppErr)
		}

		ctx, _ := context.WithTimeout(context.Background(), timeout)
		group, _ := errgroup.WithContext(ctx)
		idChan := make(chan lobby.GameLobbyID)
		group.Go(func() error {
			id, crErr := gameCreatorInit(crStream, testLobby.settings, testLobby.mark, crLogger)
			if crErr != nil {
				return crErr
			}
			idChan <- id

			if crAwaitErr := gameCreatorStartAwait(crStream, crLogger); crAwaitErr != nil {
				return crAwaitErr
			}

			crLogger.Debug().Msg("started")
			gameRes, crErr := gameCrInteractor(crStream, PlayerParams{moves: tieMoves, mark: testLobby.mark}, crLogger)
			if crErr != nil {
				return crErr
			}
			assert.Equal(t, wantRes, gameRes)
			crLogger.Debug().Msg("finished execution after first game")

			_, crErr = gameCrInteractor(crStream, PlayerParams{moves: partGameMoves, mark: oppMark}, crLogger)
			if crErr != nil {
				return crErr
			}
			return crStream.Send(&crLeaveRequest)
		})

		group.Go(func() error {
			oppLogger.Debug().Msg("before connection")
			if oppErr := gameOpponentInit(oppStream, (<-idChan).ID(), oppLogger); oppErr != nil {
				return oppErr
			}
			oppLogger.Debug().Msg("started")
			gameRes, oppErr := gameOppInteractor(oppStream, PlayerParams{moves: tieMoves, mark: oppMark}, oppLogger)
			if oppErr != nil {
				return oppErr
			}
			assert.Equal(t, wantRes, gameRes)
			oppLogger.Debug().Msg("finished execution after first game")

			_, oppErr = gameOppInteractor(oppStream, PlayerParams{moves: partGameMoves, mark: testLobby.mark}, oppLogger)
			if oppErr != nil {
				return oppErr
			}

			return nil
		})

		if oppErr = group.Wait(); oppErr != nil {
			t.Fatal(oppErr)
		}
	})

	t.Run("failure: leave-Creator(X)", func(t *testing.T) {
		testLobby := testLobbies[0]
		quit := make(chan struct{})
		conn, err, errc := startupServer(quit)
		if err != nil {
			t.Error(err)
		}

		go func() {
			t.Fatal(<-errc)
		}()

		defer func() {
			close(quit)
		}()

		var group errgroup.Group
		moves := []game.Move{{1, 1}, {0, 2}, {2, 2}, {0, 0}, {0, 1}, {2, 1}, {1, 0}, {1, 2}, {2, 0}}
		wantInterruption := transport.StopCause_STOP_CAUSE_LEAVE

		creator := transport.NewGameConfiguratorClient(conn)
		crLogger := logger.CreateDebugLogger().With().Str("player", game.EnumNamesMoveChoice[game.MoveChoice(testLobby.mark)]).Logger()
		crContext, _ := context.WithTimeout(context.Background(), timeout)

		opponent := transport.NewGameConfiguratorClient(conn)
		oppMark := (testLobby.mark + 1) & 1
		oppLogger := logger.CreateDebugLogger().With().Str("player", game.EnumNamesMoveChoice[game.MoveChoice(oppMark)]).Logger()
		oppContext, _ := context.WithTimeout(context.Background(), timeout)

		crStream, err := creator.CreateGame(crContext)
		if err != nil {
			t.Error(err)
		}

		oppStream, err := opponent.JoinGame(oppContext)
		if err != nil {
			t.Error(err)
		}

		idChan := make(chan lobby.GameLobbyID)
		group.Go(func() error {
			id, intErr := gameCreatorInit(crStream, testLobby.settings, testLobby.mark, crLogger)
			if intErr != nil {
				return intErr
			}
			idChan <- id

			if crAwaitErr := gameCreatorStartAwait(crStream, crLogger); crAwaitErr != nil {
				return crAwaitErr
			}

			if _, intErr = gameCrInteractor(crStream, PlayerParams{moves: moves[:1], mark: testLobby.mark}, crLogger); intErr != nil {
				return intErr
			}
			return crStream.Send(&crLeaveRequest)
		})

		group.Go(func() error {
			var intErr error
			if intErr = gameOpponentInit(oppStream, (<-idChan).ID(), oppLogger); intErr != nil {
				return intErr
			}

			_, intErr = gameOppInteractor(oppStream, PlayerParams{moves: moves, mark: oppMark}, oppLogger)
			if intErr != nil {
				st := status.Convert(intErr)
				if len(st.Details()) == 0 {
					return intErr
				}
				intInfo, ok := st.Details()[0].(*transport.Interruption)
				if !ok {
					t.Errorf("can not cast to interruptionInfo")
				}
				if !assert.Equal(t, wantInterruption, intInfo.Cause) {
					return intErr
				}
			}
			return nil
		})

		if err = group.Wait(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("failure: disconnect-Creator(X)", func(t *testing.T) {
		testLobby := testLobbies[0]
		quit := make(chan struct{})
		conn, err, errc := startupServer(quit)
		if err != nil {
			t.Error(err)
		}

		go func() {
			t.Fatal(<-errc)
		}()

		defer func() {
			close(quit)
		}()

		var group errgroup.Group
		moves := []game.Move{{1, 1}, {0, 2}, {2, 2}, {0, 0}, {0, 1}, {2, 1}, {1, 0}, {1, 2}, {2, 0}}
		wantInterruption := transport.StopCause_STOP_CAUSE_LEAVE

		creator := transport.NewGameConfiguratorClient(conn)
		crLogger := logger.CreateDebugLogger().With().Str("player", game.EnumNamesMoveChoice[game.MoveChoice(testLobby.mark)]).Logger()
		crContext, _ := context.WithTimeout(context.Background(), timeout)

		opponent := transport.NewGameConfiguratorClient(conn)
		oppMark := (testLobby.mark + 1) & 1
		oppLogger := logger.CreateDebugLogger().With().Str("player", game.EnumNamesMoveChoice[game.MoveChoice(oppMark)]).Logger()
		oppContext, _ := context.WithTimeout(context.Background(), timeout)

		crStream, err := creator.CreateGame(crContext)
		if err != nil {
			t.Error(err)
		}

		oppStream, err := opponent.JoinGame(oppContext)
		if err != nil {
			t.Error(err)
		}

		idChan := make(chan lobby.GameLobbyID)
		group.Go(func() error {
			id, intErr := gameCreatorInit(crStream, testLobby.settings, testLobby.mark, crLogger)
			if intErr != nil {
				return intErr
			}
			idChan <- id

			if crAwaitErr := gameCreatorStartAwait(crStream, crLogger); crAwaitErr != nil {
				return crAwaitErr
			}

			if _, intErr = gameCrInteractor(crStream, PlayerParams{moves: moves[:2], mark: testLobby.mark}, crLogger); intErr != nil {
				return intErr
			}
			crStream.Send(&crLeaveRequest)
			return crStream.CloseSend()
		})

		group.Go(func() error {
			var intErr error
			if intErr = gameOpponentInit(oppStream, (<-idChan).ID(), oppLogger); intErr != nil {
				return intErr
			}
			_, intErr = gameOppInteractor(oppStream, PlayerParams{moves: moves, mark: oppMark}, oppLogger)
			if intErr != nil {
				st := status.Convert(intErr)
				intInfo, ok := st.Details()[0].(*transport.Interruption)
				if !ok {
					t.Errorf("can not cast to interruptionInfo")
				}
				if !assert.Equal(t, wantInterruption, intInfo.Cause) {
					return intErr
				}
			}
			return nil
		})

		if err = group.Wait(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("failure: disconnect-Opponent(O)", func(t *testing.T) {
		testLobby := testLobbies[0]
		quit := make(chan struct{})
		conn, err, errc := startupServer(quit)
		if err != nil {
			t.Error(err)
		}

		go func() {
			t.Fatal(<-errc)
		}()

		defer func() {
			close(quit)
		}()

		var group errgroup.Group
		moves := []game.Move{{1, 1}, {0, 2}, {2, 2}, {0, 0}, {0, 1}, {2, 1}, {1, 0}, {1, 2}, {2, 0}}
		wantInterruption := transport.StopCause_STOP_CAUSE_LEAVE

		creator := transport.NewGameConfiguratorClient(conn)
		crLogger := logger.CreateDebugLogger().With().Str("player", game.EnumNamesMoveChoice[game.MoveChoice(testLobby.mark)]).Logger()
		crContext, _ := context.WithTimeout(context.Background(), timeout)

		opponent := transport.NewGameConfiguratorClient(conn)
		oppMark := (testLobby.mark + 1) & 1
		oppLogger := logger.CreateDebugLogger().With().Str("player", game.EnumNamesMoveChoice[game.MoveChoice(oppMark)]).Logger()
		oppContext, _ := context.WithTimeout(context.Background(), timeout)

		crStream, err := creator.CreateGame(crContext)
		if err != nil {
			t.Error(err)
		}

		oppStream, err := opponent.JoinGame(oppContext)
		if err != nil {
			t.Error(err)
		}

		idChan := make(chan lobby.GameLobbyID)
		group.Go(func() error {
			id, err := gameCreatorInit(crStream, testLobby.settings, testLobby.mark, crLogger)
			if err != nil {
				return err
			}
			idChan <- id

			if crAwaitErr := gameCreatorStartAwait(crStream, crLogger); crAwaitErr != nil {
				return crAwaitErr
			}
			if _, err = gameCrInteractor(crStream, PlayerParams{moves: moves[:2], mark: testLobby.mark}, crLogger); err != nil {
				st := status.Convert(err)
				intInfo, ok := st.Details()[0].(*transport.Interruption)
				if !ok {
					t.Errorf("can not cast to interruptionInfo")
				}
				if !assert.Equal(t, wantInterruption, intInfo.Cause) {
					return err
				}
			}
			return nil
		})

		group.Go(func() error {
			if err := gameOpponentInit(oppStream, (<-idChan).ID(), oppLogger); err != nil {
				return err
			}
			_, err = gameOppInteractor(oppStream, PlayerParams{moves: moves[:2], mark: oppMark}, oppLogger)
			if err != nil {
				return err
			}
			return oppStream.Send(&oppLeaveRequest)
			//oppStreamCancel()
		})

		if err = group.Wait(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("check: cancel does not block game controller", func(t *testing.T) {
		testLobby := testLobbies[0]
		quit := make(chan struct{})
		conn, err, errc := startupServer(quit)
		if err != nil {
			t.Fatal(err)
		}

		go func() {
			t.Fatal(<-errc)
		}()

		defer func() {
			close(quit)
		}()

		var group errgroup.Group
		moves := []game.Move{{1, 1}, {0, 2}, {2, 2}, {0, 0}, {0, 1}, {2, 1}, {1, 0}, {1, 2}, {2, 0}}
		wantInterruption := transport.StopCause_STOP_CAUSE_LEAVE

		creator := transport.NewGameConfiguratorClient(conn)
		crLogger := logger.CreateDebugLogger().With().Str("player", game.EnumNamesMoveChoice[game.MoveChoice(testLobby.mark)]).Logger()
		crContext, crStreamCancel := context.WithTimeout(context.Background(), timeout)

		opponent := transport.NewGameConfiguratorClient(conn)
		oppMark := (testLobby.mark + 1) & 1
		oppLogger := logger.CreateDebugLogger().With().Str("player", game.EnumNamesMoveChoice[game.MoveChoice(oppMark)]).Logger()
		oppContext, oppStreamCancel := context.WithTimeout(context.Background(), timeout)

		crStream, err := creator.CreateGame(crContext)
		if err != nil {
			t.Fatal(err)
		}

		oppStream, err := opponent.JoinGame(oppContext)
		if err != nil {
			t.Fatal(err)
		}

		idChan := make(chan lobby.GameLobbyID)
		group.Go(func() error {
			id, err := gameCreatorInit(crStream, testLobby.settings, testLobby.mark, crLogger)
			if err != nil {
				return err
			}
			idChan <- id
			if err = gameCreatorStartAwait(crStream, crLogger); err != nil {
				return err
			}

			crStreamCancel()
			return nil
		})

		group.Go(func() error {
			if err := gameOpponentInit(oppStream, (<-idChan).ID(), oppLogger); err != nil {
				return err
			}
			_, err = gameOppInteractor(oppStream, PlayerParams{moves: moves, mark: oppMark}, oppLogger)
			if err != nil {
				st := status.Convert(err)
				for _, detail := range st.Details() {
					intInfo, ok := detail.(*transport.Interruption)
					if !ok {
						t.Errorf("can not cast to interruptionInfo")
					}
					if !assert.Equal(t, wantInterruption, intInfo.Cause) {
						return err
					}
				}
			}
			oppStreamCancel()
			return nil
		})

		if err = group.Wait(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("check: interruption instantly close game", func(t *testing.T) {
		testLobby := testLobbies[0]
		quit := make(chan struct{})
		conn, err, errc := startupServer(quit)
		if err != nil {
			t.Fatal(err)
		}

		go func() {
			t.Fatal(<-errc)
		}()

		defer func() {
			close(quit)
		}()

		var group errgroup.Group
		moves := []game.Move{{1, 1}, {0, 2}}
		wantInterruption := transport.StopCause_STOP_CAUSE_LEAVE

		creator := transport.NewGameConfiguratorClient(conn)
		crLogger := logger.CreateDebugLogger().With().Str("player", game.EnumNamesMoveChoice[game.MoveChoice(testLobby.mark)]).Logger()
		crContext, _ := context.WithTimeout(context.Background(), timeout)

		opponent := transport.NewGameConfiguratorClient(conn)
		oppMark := (testLobby.mark + 1) & 1
		oppLogger := logger.CreateDebugLogger().With().Str("player", game.EnumNamesMoveChoice[game.MoveChoice(oppMark)]).Logger()
		oppContext, _ := context.WithTimeout(context.Background(), timeout)
		//oppContext, _ := context.WithTimeout(context.Background(), timeout)

		crStream, err := creator.CreateGame(crContext)
		if err != nil {
			t.Fatal(err)
		}

		oppStream, err := opponent.JoinGame(oppContext)
		if err != nil {
			t.Fatal(err)
		}

		isInterupted := make(chan bool)
		idChan := make(chan lobby.GameLobbyID)
		group.Go(func() error {
			id, intErr := gameCreatorInit(crStream, testLobby.settings, testLobby.mark, crLogger)
			if intErr != nil {
				return intErr
			}
			idChan <- id
			if intErr = gameCreatorStartAwait(crStream, crLogger); intErr != nil {
				return intErr
			}

			<-isInterupted
			if _, gameErr := gameCrInteractor(crStream, PlayerParams{moves: moves[:2], mark: testLobby.mark}, crLogger); gameErr != nil {
				return gameErr
			}
			select {
			case <-crStream.Context().Done():
				return crStream.Context().Err()
			default:
				t.Fatal("interruption from other player does not interrupt game")
			}

			return nil
		})

		group.Go(func() error {
			if err := gameOpponentInit(oppStream, (<-idChan).ID(), oppLogger); err != nil {
				return err
			}
			err = oppStream.Send(&oppLeaveRequest)
			close(isInterupted)
			return err
		})

		if err = group.Wait(); err != nil {
			st := status.Convert(err)
			intInfo, ok := st.Details()[0].(*transport.Interruption)
			if !ok {
				t.Errorf("can not cast to interruptionInfo")
			}
			if !assert.Equal(t, wantInterruption, intInfo.Cause) {
				t.Fatal(err)
			}
		}
	})

}

func startupServer(quit <-chan struct{}) (conn *grpc.ClientConn, err error, errc chan error) {
	l := bufconn.Listen(bufSize)
	s := grpc.NewServer(serveOpts...)
	go func() {
		go func() {
			if err := s.Serve(l); err != nil {
				log.Error().Err(err).Msg("failed to serve")
				errc <- err
			}
		}()
		<-quit
		conn.Close()
		return
	}()

	servLogger := logger.CreateDebugLogger()
	controller := lobby.NewGameLobbyController(lobby.PlainGameRecorder, servLogger)
	service := NewGameService(controller, servLogger)
	transport.RegisterGameConfiguratorServer(s, service)

	errc = make(chan error)

	ctx := context.Background()
	conn, err = grpc.DialContext(ctx, "bufnet", grpc.WithDialer(
		func(s string, duration time.Duration) (conn net.Conn, e error) {
			return l.Dial()
		}),
		grpc.WithInsecure(),
		//grpc.WithDefaultCallOptions(grpc.CallCustomCodec(flatbuffers.FlatbuffersCodec{})),
	)

	return conn, err, errc
}

func gameCreatorInit(
	stream transport.GameConfigurator_CreateGameClient,
	settings game.GameSettings,
	mark game.PlayerMark,
	logger zerolog.Logger,
) (lobbyID lobby.GameLobbyID, err error) {
	params := transport.CreateRequest{
		Payload: &transport.CreateRequest_Params{Params: &transport.GameParams{
			Rows: uint32(settings.Rows),
			Cols: uint32(settings.Cols),
			Win:  uint32(settings.Win),
			Mark: enumMarkType[mark],
		}},
	}

	if err := stream.Send(&params); err != nil {
		return nil, err
	}

	out, err := stream.Recv()
	if err != nil {
		return nil, err
	}

	id, ok := out.Payload.(*transport.CreateResponse_GameId)
	if !ok {
		return nil, fmt.Errorf("expected %s, but got %s",
			protoreflect.ValueOf(transport.CreateResponse_GameId{}).String(),
			out.ProtoReflect().Descriptor().Name())
	}
	return &gameLobbyID{id: id.GameId}, nil
}

func gameCreatorStartAwait(stream transport.GameConfigurator_CreateGameClient, logger zerolog.Logger) error {
	out, err := stream.Recv()
	if err != nil {
		return err
	}

	started, ok := out.Payload.(*transport.CreateResponse_Status)
	if !ok {
		return fmt.Errorf("expected %s, but got %s",
			protoreflect.ValueOf(transport.CreateResponse_Status{}).String(),
			out.ProtoReflect().Descriptor().Name())
	}
	if started.Status != transport.GameStatus_GAME_STATUS_GAME_STARTED {
		err = fmt.Errorf(
			"expected Event Game Started, got %s",
			transport.GameStatus.Type(transport.GameStatus_GAME_STATUS_GAME_STARTED).Descriptor().Name(),
		)
		logger.Error().Err(err)
	}
	return err
}

func gameOpponentInit(stream transport.GameConfigurator_JoinGameClient, ID uint32, logger zerolog.Logger) error {
	gameId := transport.JoinRequest{
		Payload: &transport.JoinRequest_GameId{GameId: ID},
	}

	if err := stream.Send(&gameId); err != nil {
		return err
	}

	out, err := stream.Recv()
	if err != nil {
		logger.Error().Err(err).Msg("receive status")
		return err
	}
	gameStatus, ok := out.Payload.(*transport.JoinResponse_Status)
	if !ok {
		return fmt.Errorf("expected %s, but got %s",
			protoreflect.ValueOf(transport.JoinResponse_Status{}).String(),
			out.ProtoReflect().Descriptor().Name())
	}

	if gameStatus.Status != transport.GameStatus_GAME_STATUS_GAME_STARTED {
		err = fmt.Errorf(
			"expected Event Game Started, got %s",
			transport.GameStatus.Type(transport.GameStatus_GAME_STATUS_GAME_STARTED).Descriptor().Name(),
		)
		logger.Error().Err(err)
	}
	return err
}

func gameCrInteractor(stream transport.GameConfigurator_CreateGameClient, params PlayerParams, logger zerolog.Logger) (interface{}, error) {
	mask := int(params.mark)
	turn := params.turn
	var sending bool
	l := logger.With().Int("moves", len(params.moves)).Logger()
	l.Debug().Msg("crInteractor started")
	for _, move := range params.moves {
		sending = (turn & 1) == mask
		playerLogger := logger.With().Int("turn", turn).Logger()
		playerLogger.Debug().Msg("turn started")

		if sending {
			playerLogger.Debug().Msg("intend to send move")
			m := transport.CreateRequest{
				Payload: &transport.CreateRequest_Move{Move: &transport.Move{Row: uint32(move.I), Col: uint32(move.J)}},
			}
			if err := stream.Send(&m); err != nil && err != io.EOF {
				return nil, err
			}
			playerLogger.Debug().Dict("move", zerolog.Dict().
				Uint16("i", move.I).
				Uint16("j", move.J)).Msg("sent")
		}
		playerLogger.Debug().Msg("intent to receive msg")

		in, err := stream.Recv()
		if err != nil {
			playerLogger.Error().Err(err).Msg("received error")
			return nil, err
		}
		if in == nil {
			return nil, fmt.Errorf("WTF")
		}
		if !sending && in.Move != nil {
			receivedMove := game.Move{I: uint16(in.Move.Row), J: uint16(in.Move.Col)}
			playerLogger.Debug().Dict("move", zerolog.Dict().
				Uint16("i", move.I).
				Uint16("j", move.J)).Msg("received")
			if !reflect.DeepEqual(receivedMove, move) {
				return nil, errors.Newf("received game step, move have wrong value")
			}
		}

		switch st := in.Payload.(type) {
		case *transport.CreateResponse_Status:
			s := st.Status
			switch s {
			case transport.GameStatus_GAME_STATUS_OK:
				turn++
				playerLogger.Debug().Msg("all OK going to next turn")
			case transport.GameStatus_GAME_STATUS_TIE:
				playerLogger.Debug().Str("result", "Tie").Msg("game ended")
				return s, nil
			default:
				err = fmt.Errorf(
					"unexpected Game Status, got %s",
					transport.GameStatus.Type(s).Descriptor().Name(),
				)
				return nil, err
			}
		case *transport.CreateResponse_WinLine:
			m := game.EnumNamesMoveChoice[game.MoveChoice(enumPlayerMark[st.WinLine.Mark])]
			start := st.WinLine.Start
			end := st.WinLine.End
			if start == nil && end == nil {
				playerLogger.Debug().Str("Winner", m).Msg("opponent gave-up")
			} else {
				playerLogger.Debug().Str("Winner", m).Dict("win line", zerolog.Dict().
					Dict("start", zerolog.Dict().
						Uint32("i", start.Row).
						Uint32("j", start.Col)).
					Dict("end", zerolog.Dict().
						Uint32("i", end.Row).
						Uint32("j", end.Col)),
				).Msg("game ended")
			}
			return st.WinLine, nil
		default:
			return nil, fmt.Errorf("unexpected payload, got %s",
				in.ProtoReflect().Descriptor().Name(),
			)
		}
	}
	return transport.GameStatus_GAME_STATUS_OK, nil
}

func gameOppInteractor(stream transport.GameConfigurator_JoinGameClient, params PlayerParams, logger zerolog.Logger) (interface{}, error) {
	mask := int(params.mark)
	turn := params.turn
	var sending bool
	l := logger.With().Int("moves", len(params.moves)).Logger()
	l.Debug().Msg("opp interactor started")
	for _, move := range params.moves {
		sending = (turn & 1) == mask
		playerLogger := logger.With().Int("turn", turn).Logger()
		playerLogger.Debug().Msg("turn started")

		if sending {
			playerLogger.Debug().Msg("intend to send move")
			m := transport.JoinRequest{
				Payload: &transport.JoinRequest_Move{Move: &transport.Move{Row: uint32(move.I), Col: uint32(move.J)}},
			}
			if err := stream.Send(&m); err != nil && err != io.EOF {
				return nil, err
			}
			playerLogger.Debug().Dict("move", zerolog.Dict().
				Uint16("i", move.I).
				Uint16("j", move.J)).Msg("sent")
		}
		playerLogger.Debug().Msg("intent to receive msg")

		in, err := stream.Recv()
		if err != nil {
			return nil, err
		}
		if !sending && in.Move != nil {
			receivedMove := game.Move{uint16(in.Move.Row), uint16(in.Move.Col)}
			playerLogger.Debug().Dict("move", zerolog.Dict().
				Uint16("i", move.I).
				Uint16("j", move.J)).Msg("received")
			if !reflect.DeepEqual(receivedMove, move) {
				return nil, errors.Newf("received game step, move have wrong value")
			}
		}

		switch st := in.Payload.(type) {
		case *transport.JoinResponse_Status:
			s := st.Status
			switch s {
			case transport.GameStatus_GAME_STATUS_OK:
				turn++
				playerLogger.Debug().Msg("all OK going to next turn")
			case transport.GameStatus_GAME_STATUS_TIE:
				playerLogger.Debug().Str("result", "Tie").Msg("game ended")
				return s, nil
			default:
				err = fmt.Errorf(
					"unexpected Game Status, got %s",
					transport.GameStatus.Type(s).Descriptor().Name(),
				)
				return nil, err
			}
		case *transport.JoinResponse_WinLine:
			m := game.EnumNamesMoveChoice[game.MoveChoice(enumPlayerMark[st.WinLine.Mark])]
			start := st.WinLine.Start
			end := st.WinLine.End
			if start == nil && end == nil {
				playerLogger.Debug().Str("Winner", m).Msg("opponent gave-up")
			} else {
				playerLogger.Debug().Str("Winner", m).Dict("win line", zerolog.Dict().
					Dict("start", zerolog.Dict().
						Uint32("i", start.Row).
						Uint32("j", start.Col)).
					Dict("end", zerolog.Dict().
						Uint32("i", end.Row).
						Uint32("j", end.Col)),
				).Msg("game ended")
			}
			return st.WinLine, nil
		default:
			return nil, fmt.Errorf("unexpected payload, got %s",
				in.ProtoReflect().Descriptor().Name(),
			)
		}
	}
	return transport.GameStatus_GAME_STATUS_OK, nil
}
