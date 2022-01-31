package rpc_service

import (
	"context"
	"fmt"
	"github.com/izgib/tttserver/base"

	"github.com/izgib/tttserver/game"
	"github.com/izgib/tttserver/lobby"
	"github.com/izgib/tttserver/recorder/text"
	"github.com/rs/zerolog/log"
	"github.com/sirkon/errors"
	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/status"
	"net"
	"reflect"
	"testing"
	"time"

	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/test/bufconn"

	"github.com/izgib/tttserver/internal"
	"github.com/izgib/tttserver/rpc_service/transport"
)

const bufSize = 1024 * 16

type gameLobbyID struct {
	id int16
}

func (g gameLobbyID) GetID() int16 {
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
	grpc.CustomCodec(flatbuffers.FlatbuffersCodec{}),
	grpc.KeepaliveEnforcementPolicy(keepalivePolicy),
	grpc.KeepaliveParams(keepaliveParams)}

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

	t.Run("success: Win-X", func(t *testing.T) {
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
		creator := transport.NewGameConfiguratorClient(conn)
		crLogger := internal.CreateDebugLogger().With().Str("player", game.EnumNamesMoveChoice[game.MoveChoice(testLobby.mark)]).Logger()
		crContext := context.Background()

		opponent := transport.NewGameConfiguratorClient(conn)
		oppMark := (testLobby.mark + 1) & 1
		oppLogger := internal.CreateDebugLogger().With().Str("player", game.EnumNamesMoveChoice[game.MoveChoice(oppMark)]).Logger()
		oppContext := context.Background()

		crStream, err := creator.CreateGame(crContext)
		if err != nil {
			t.Fatal(err)
		}

		oppStream, err := opponent.JoinGame(oppContext)
		if err != nil {
			t.Fatal(err)
		}

		var group errgroup.Group
		idChan := make(chan base.GameLobbyID)
		group.Go(func() error {
			id, crErr := gameCreatorInit(crStream, testLobby.settings, testLobby.mark, crLogger)
			if crErr != nil {
				return crErr
			}
			crLogger.Debug().Int16("game", id.GetID()).Msg("initialized")
			idChan <- id

			if crAwaitErr := gameCreatorStartAwait(crStream, crLogger); crAwaitErr != nil {
				return crAwaitErr
			}
			crLogger.Debug().Int16("game", id.GetID()).Msg("started")

			if gameErr := gameCrInteractor(crStream, moves, testLobby.mark, crLogger); gameErr != nil {
				return gameErr
			}
			return nil
		})

		group.Go(func() error {
			if err := gameOpponentInit(oppStream, (<-idChan).GetID(), oppLogger); err != nil {
				return err
			}
			if err := gameOppInteractor(oppStream, moves, oppMark, oppLogger); err != nil {
				return err
			}
			return nil
		})

		if err = group.Wait(); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("failure: disconnect-X", func(t *testing.T) {
		lobbyParams := testLobbies[0]
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
		wantInterruption := InterruptionCause_LEAVE

		creator := transport.NewGameConfiguratorClient(conn)
		crLogger := internal.CreateDebugLogger().With().Str("player", game.EnumNamesMoveChoice[game.MoveChoice(lobbyParams.mark)]).Logger()
		crContext, crStreamCancel := context.WithCancel(context.Background())

		opponent := transport.NewGameConfiguratorClient(conn)
		oppMark := (lobbyParams.mark + 1) & 1
		oppLogger := internal.CreateDebugLogger().With().Str("player", game.EnumNamesMoveChoice[game.MoveChoice(oppMark)]).Logger()
		oppContext := context.Background()

		crStream, err := creator.CreateGame(crContext)
		if err != nil {
			t.Error(err)
		}

		oppStream, err := opponent.JoinGame(oppContext)
		if err != nil {
			t.Error(err)
		}

		idChan := make(chan base.GameLobbyID)
		group.Go(func() error {
			id, intErr := gameCreatorInit(crStream, lobbyParams.settings, lobbyParams.mark, crLogger)
			if intErr != nil {
				return intErr
			}
			idChan <- id

			if crAwaitErr := gameCreatorStartAwait(crStream, crLogger); crAwaitErr != nil {
				return crAwaitErr
			}

			if intErr = gameCrInteractor(crStream, moves[:1], lobbyParams.mark, crLogger); intErr != nil {
				return intErr
			}
			crStreamCancel()
			return nil
		})

		group.Go(func() error {
			var intErr error
			if intErr = gameOpponentInit(oppStream, (<-idChan).GetID(), oppLogger); intErr != nil {
				return intErr
			}
			if intErr = gameOppInteractor(oppStream, moves, oppMark, oppLogger); intErr != nil {
				st := status.Convert(intErr)
				intInfo, ok := st.Details()[0].(*InterruptionInfo)
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
	t.Run("failure: disconnect-O", func(t *testing.T) {
		lobbyParams := testLobbies[0]
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
		wantInterruption := InterruptionCause_LEAVE

		creator := transport.NewGameConfiguratorClient(conn)
		crLogger := internal.CreateDebugLogger().With().Str("player", game.EnumNamesMoveChoice[game.MoveChoice(lobbyParams.mark)]).Logger()
		crContext := context.Background()

		opponent := transport.NewGameConfiguratorClient(conn)
		oppMark := (lobbyParams.mark + 1) & 1
		oppLogger := internal.CreateDebugLogger().With().Str("player", game.EnumNamesMoveChoice[game.MoveChoice(oppMark)]).Logger()
		oppContext, oppStreamCancel := context.WithCancel(context.Background())

		crStream, err := creator.CreateGame(crContext)
		if err != nil {
			t.Error(err)
		}

		oppStream, err := opponent.JoinGame(oppContext)
		if err != nil {
			t.Error(err)
		}

		idChan := make(chan base.GameLobbyID)
		group.Go(func() error {
			id, err := gameCreatorInit(crStream, lobbyParams.settings, lobbyParams.mark, crLogger)
			if err != nil {
				return err
			}
			idChan <- id

			if crAwaitErr := gameCreatorStartAwait(crStream, crLogger); crAwaitErr != nil {
				return crAwaitErr
			}
			if err = gameCrInteractor(crStream, moves, lobbyParams.mark, crLogger); err != nil {
				st := status.Convert(err)
				intInfo, ok := st.Details()[0].(*InterruptionInfo)
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
			if err := gameOpponentInit(oppStream, (<-idChan).GetID(), oppLogger); err != nil {
				return err
			}
			if err := gameOppInteractor(oppStream, moves[:2], oppMark, oppLogger); err != nil {
				st := status.Convert(err)
				intInfo, ok := st.Details()[0].(*InterruptionInfo)
				if !ok {
					t.Errorf("can not cast to interruptionInfo")
				}
				if !assert.Equal(t, wantInterruption, intInfo.Cause) {
					return err
				}
			}
			oppStreamCancel()
			return nil
		})

		if err = group.Wait(); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("check: cancel does not block game controller", func(t *testing.T) {
		lobbyParams := testLobbies[0]
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
		wantInterruption := InterruptionCause_LEAVE

		creator := transport.NewGameConfiguratorClient(conn)
		crLogger := internal.CreateDebugLogger().With().Str("player", game.EnumNamesMoveChoice[game.MoveChoice(lobbyParams.mark)]).Logger()
		crContext, crStreamCancel := context.WithCancel(context.Background())

		opponent := transport.NewGameConfiguratorClient(conn)
		oppMark := (lobbyParams.mark + 1) & 1
		oppLogger := internal.CreateDebugLogger().With().Str("player", game.EnumNamesMoveChoice[game.MoveChoice(oppMark)]).Logger()
		oppContext, oppStreamCancel := context.WithCancel(context.Background())

		crStream, err := creator.CreateGame(crContext)
		if err != nil {
			t.Fatal(err)
		}

		oppStream, err := opponent.JoinGame(oppContext)
		if err != nil {
			t.Fatal(err)
		}

		idChan := make(chan base.GameLobbyID)
		group.Go(func() error {
			id, err := gameCreatorInit(crStream, lobbyParams.settings, lobbyParams.mark, crLogger)
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
			if err := gameOpponentInit(oppStream, (<-idChan).GetID(), oppLogger); err != nil {
				return err
			}
			if err := gameOppInteractor(oppStream, moves, oppMark, oppLogger); err != nil {
				st := status.Convert(err)
				for _, detail := range st.Details() {
					intInfo, ok := detail.(*InterruptionInfo)
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

	logger := internal.CreateDebugLogger()
	controller := lobby.NewGameLobbyController(text.NewGameRecorder, logger)
	service := NewGameService(controller, logger)
	transport.RegisterGameConfiguratorServer(s, service)

	errc = make(chan error)

	ctx := context.Background()
	conn, err = grpc.DialContext(ctx, "bufnet", grpc.WithDialer(
		func(s string, duration time.Duration) (conn net.Conn, e error) {
			return l.Dial()
		}),
		grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(grpc.CallCustomCodec(flatbuffers.FlatbuffersCodec{})),
	)

	return conn, err, errc
}

func gameCreatorInit(
	stream transport.GameConfigurator_CreateGameClient,
	settings game.GameSettings,
	mark game.PlayerMark,
	logger zerolog.Logger,
) (lobbyID base.GameLobbyID, err error) {
	b := flatbuffers.NewBuilder(0)

	transport.GameParamsStart(b)
	transport.GameParamsAddRows(b, settings.Rows)
	transport.GameParamsAddCols(b, settings.Cols)
	transport.GameParamsAddWin(b, settings.Win)
	transport.GameParamsAddMark(b, transport.MarkType(mark))
	gcParams := transport.GameParamsEnd(b)
	crWrapRequest(b, transport.CreatorReqMsgGameParams, gcParams)

	if err := stream.Send(b); err != nil {
		return nil, err
	}

	out, err := stream.Recv()
	if err != nil {
		return nil, err
	}

	var gameId int16
	if out.RespType() == transport.CreatorRespMsgGameId {
		gameId = crResponseUnwrapperID(out).ID()
		return &gameLobbyID{id: gameId}, nil
	}

	out, err = stream.Recv()
	if err != nil {
		return nil, err
	}

	if out.RespType() != transport.CreatorRespMsgGameEvent {
		resp := &flatbuffers.Table{}
		gameStart := &transport.GameEvent{}
		if out.Resp(resp) {
			gameStart.Init(resp.Bytes, resp.Pos)
			if gameStart.Type() != transport.GameEventTypeGameStarted {
				err := fmt.Errorf("expected Event Game Started, got %s", transport.EnumNamesGameEventType[gameStart.Type()])
				logger.Error().Err(err)
				return nil, err
			}
		}
	}

	return &gameLobbyID{id: gameId}, nil
}

func gameCreatorStartAwait(stream transport.GameConfigurator_CreateGameClient, logger zerolog.Logger) error {
	out, err := stream.Recv()
	if err != nil {
		return err
	}

	if out.RespType() != transport.CreatorRespMsgGameEvent {
		resp := &flatbuffers.Table{}
		gameStart := &transport.GameEvent{}
		if out.Resp(resp) {
			gameStart.Init(resp.Bytes, resp.Pos)
			if gameStart.Type() != transport.GameEventTypeGameStarted {
				err := fmt.Errorf("expected Event Game Started, got %s", transport.EnumNamesGameEventType[gameStart.Type()])
				logger.Error().Err(err)
				return err
			}
		}
	}
	return nil
}

func gameOpponentInit(stream transport.GameConfigurator_JoinGameClient, ID int16, logger zerolog.Logger) error {
	b := flatbuffers.NewBuilder(0)

	transport.GameIdStart(b)
	transport.GameIdAddID(b, ID)
	oppRequestWrapper(b, transport.OpponentReqMsgGameId, transport.GameIdEnd(b))

	if err := stream.Send(b); err != nil {
		return err
	}

	out, err := stream.Recv()
	if err != nil {
		return err
	}

	if out.RespType() != transport.OpponentRespMsgGameEvent {
		resp := &flatbuffers.Table{}
		gameStart := &transport.GameEvent{}
		if out.Resp(resp) {
			gameStart.Init(resp.Bytes, resp.Pos)
			if gameStart.Type() != transport.GameEventTypeGameStarted {
				err := fmt.Errorf("expected Event Game Started, got %s", transport.EnumNamesGameEventType[gameStart.Type()])
				logger.Error().Err(err)
				return err
			}
		}
	}

	return nil

}

func sendMove(i interface{}, move game.Move) error {
	b := flatbuffers.NewBuilder(0)
	transport.MoveStart(b)
	transport.MoveAddRow(b, move.I)
	transport.MoveAddCol(b, move.J)

	switch stream := i.(type) {
	case transport.GameConfigurator_CreateGameClient:
		crWrapRequest(b, transport.CreatorReqMsgMove, transport.MoveEnd(b))
		return stream.Send(b)
	case transport.GameConfigurator_JoinGameClient:
		oppWrapRequest(b, transport.OpponentReqMsgMove, transport.MoveEnd(b))
		return stream.Send(b)
	}
	return nil
}

func gameCrInteractor(stream transport.GameConfigurator_CreateGameClient, moves []game.Move, mark game.PlayerMark, logger zerolog.Logger) error {
	mask := int(mark)
	turn := 0
	for _, move := range moves {
		playerLogger := logger.With().Int("turn", turn).Logger()
		playerLogger.Debug().Msg("turn started")
		var event *transport.GameEvent
		if (turn & 1) == mask {
			playerLogger.Debug().Msg("intend to send move")
			if err := sendMove(stream, move); err != nil {
				return err
			}
			playerLogger.Debug().Dict("move", zerolog.Dict().
				Int16("i", moves[turn].I).
				Int16("j", moves[turn].J)).Msg("sent")
			playerLogger.Debug().Msg("intent to receive event")
			in, err := stream.Recv()
			if err != nil {
				return err
			}
			event = crResponseUnwrapperEvent(in)
		} else {
			playerLogger.Debug().Msg("intent to receive response")
		moveLoopCr:
			for j := 0; j < 2; j++ {
				in, err := stream.Recv()
				if err != nil {
					return err
				}
				switch in.RespType() {
				case transport.CreatorRespMsgMove:
					resp := crResponseUnwrapperMove(in)
					receivedMove := game.Move{resp.Row(), resp.Col()}
					playerLogger.Debug().Msg("got move")
					if !reflect.DeepEqual(receivedMove, move) {
						return errors.Newf("received move have wrong value")
					}
					playerLogger.Debug().Dict("move", zerolog.Dict().
						Int16("i", moves[turn].I).
						Int16("j", moves[turn].J)).Msg("received")
				case transport.CreatorRespMsgGameEvent:
					playerLogger.Debug().Msg("received event")
					event = crResponseUnwrapperEvent(in)
					break moveLoopCr
				}
			}
		}
		switch eventType := event.Type(); eventType {
		case transport.GameEventTypeOK:
			turn++
			playerLogger.Debug().Msg("all OK going to next turn")
		case transport.GameEventTypeTie, transport.GameEventTypeWin:
			if event.Type() == transport.GameEventTypeWin {
				winLine := event.FollowUp(nil)
				start := winLine.Start(nil)
				end := winLine.End(nil)
				playerLogger.Debug().Str("Winner", game.EnumNamesMoveChoice[game.MoveChoice(winLine.Mark())]).Dict("win line", zerolog.Dict().
					Dict("start", zerolog.Dict().
						Int16("i", start.Row()).
						Int16("j", start.Col())).
					Dict("end", zerolog.Dict().
						Int16("i", end.Row()).
						Int16("j", end.Col())),
				).Msg("game ended")
			} else {
				playerLogger.Debug().Str("result", "Tie").Msg("game ended")
			}
			return nil
		default:
			err := fmt.Errorf("got interruption event: %s", transport.EnumNamesGameEventType[eventType])
			playerLogger.Error().Str("type", transport.EnumNamesGameEventType[eventType]).Msg("got interruption event")
			return err
		}
	}
	return nil
}

func gameOppInteractor(stream transport.GameConfigurator_JoinGameClient, moves []game.Move, mark game.PlayerMark, logger zerolog.Logger) error {
	var in *transport.OppResponse
	var err error
	var event *transport.GameEvent
	mask := int(mark)
	turn := 0

	for _, move := range moves {
		playerLogger := logger.With().Int("turn", turn).Logger()
		playerLogger.Debug().Msg("turn started")
		if (turn & 1) == mask {
			playerLogger.Debug().Msg("intend to send move")

			select {
			case <-stream.Context().Done():
				return stream.Context().Err()
			default:
				if err := sendMove(stream, move); err != nil {
					return err
				}
			}

			playerLogger.Debug().Dict("move", zerolog.Dict().
				Int16("i", moves[turn].I).
				Int16("j", moves[turn].J)).Msg("sent")
			playerLogger.Debug().Msg("intent to receive event")
			in, err := stream.Recv()
			if err != nil {
				return err
			}
			if in.RespType() == transport.OpponentRespMsgGameEvent {
				event = oppResponseUnwrapperEvent(in)
			} else {
				panic("wtf")
			}
		} else {
			playerLogger.Debug().Msg("intent to receive response")
		moveLoopOpp:
			for j := 0; j < 2; j++ {
				in, err = stream.Recv()
				if err != nil {
					return err
				}
				switch in.RespType() {
				case transport.OpponentRespMsgMove:
					resp := oppResponseUnwrapperMove(in)
					receivedMove := game.Move{resp.Row(), resp.Col()}
					if receivedMove != move {
						return errors.Newf("received move have wrong value")
					}
					playerLogger.Debug().Dict("move", zerolog.Dict().
						Int16("i", moves[turn].I).
						Int16("j", moves[turn].J)).Msg("received")
				case transport.OpponentRespMsgGameEvent:
					playerLogger.Debug().Msg("received event")
					event = oppResponseUnwrapperEvent(in)
					break moveLoopOpp
				}
			}
		}
		switch eventType := event.Type(); eventType {
		case transport.GameEventTypeOK:
			turn++
			playerLogger.Debug().Msg("all OK going to next turn")
		case transport.GameEventTypeTie, transport.GameEventTypeWin:
			if event.Type() == transport.GameEventTypeWin {
				winLine := event.FollowUp(nil)
				start := winLine.Start(nil)
				end := winLine.End(nil)
				playerLogger.Debug().Str("Winner", game.EnumNamesMoveChoice[game.MoveChoice(winLine.Mark())]).Dict("win line", zerolog.Dict().
					Dict("start", zerolog.Dict().
						Int16("i", start.Row()).
						Int16("j", start.Col())).
					Dict("end", zerolog.Dict().
						Int16("i", end.Row()).
						Int16("j", end.Col())),
				).Msg("game ended")
			} else {
				playerLogger.Debug().Str("result", "Tie").Msg("game ended")
			}
			return nil
		default:
			err := fmt.Errorf("got interruption event: %s", transport.EnumNamesGameEventType[eventType])
			playerLogger.Error().Str("type", transport.EnumNamesGameEventType[eventType]).Msg("got interruption event")
			return err
		}
	}
	return nil
}

func oppRequestWrapper(b *flatbuffers.Builder, reqType transport.OpponentReqMsg, req flatbuffers.UOffsetT) {
	transport.OppRequestStart(b)
	transport.OppRequestAddReqType(b, reqType)
	transport.OppRequestAddReq(b, req)
	b.Finish(transport.OppRequestEnd(b))
}

func crResponseUnwrapperMove(resp *transport.CrResponse) *transport.Move {
	respT := &flatbuffers.Table{}
	move := &transport.Move{}
	if resp.Resp(respT) {
		move.Init(respT.Bytes, respT.Pos)
	}
	return move
}

func crResponseUnwrapperEvent(resp *transport.CrResponse) *transport.GameEvent {
	respT := &flatbuffers.Table{}
	event := &transport.GameEvent{}
	if resp.Resp(respT) {
		event.Init(respT.Bytes, respT.Pos)
	}
	return event
}

func crResponseUnwrapperID(resp *transport.CrResponse) *transport.GameId {
	respT := &flatbuffers.Table{}
	gameID := &transport.GameId{}
	if resp.Resp(respT) {
		gameID.Init(respT.Bytes, respT.Pos)
	}
	return gameID
}

func oppResponseUnwrapperMove(resp *transport.OppResponse) *transport.Move {
	respT := &flatbuffers.Table{}
	move := &transport.Move{}
	if resp.Resp(respT) {
		move.Init(respT.Bytes, respT.Pos)
	}
	return move
}

func oppResponseUnwrapperEvent(resp *transport.OppResponse) *transport.GameEvent {
	respT := &flatbuffers.Table{}
	event := &transport.GameEvent{}
	if resp.Resp(respT) {
		event.Init(respT.Bytes, respT.Pos)
	}
	return event
}

func createRange(b *flatbuffers.Builder, start int16, end int16) flatbuffers.UOffsetT {
	transport.RangeStart(b)
	transport.RangeAddStart(b, start)
	transport.RangeAddEnd(b, end)
	return transport.RangeEnd(b)
}

func crWrapRequest(b *flatbuffers.Builder, respType transport.CreatorReqMsg, resp flatbuffers.UOffsetT) {
	transport.CrRequestStart(b)
	transport.CrRequestAddReqType(b, respType)
	transport.CrRequestAddReq(b, resp)
	b.Finish(transport.CrRequestEnd(b))
}

func oppWrapRequest(b *flatbuffers.Builder, respType transport.OpponentReqMsg, resp flatbuffers.UOffsetT) {
	transport.OppRequestStart(b)
	transport.OppRequestAddReqType(b, respType)
	transport.OppRequestAddReq(b, resp)
	b.Finish(transport.OppRequestEnd(b))
}
