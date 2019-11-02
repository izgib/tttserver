package rpc_service

import (
	"context"
	"fmt"
	"github.com/izgib/tttserver/game"
	"github.com/izgib/tttserver/game/interface/recorder/text"
	"github.com/stretchr/testify/assert"
	//"google.golang.org/grpc/status"
	"net"
	"sync"
	"testing"
	"time"

	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/test/bufconn"

	"github.com/izgib/tttserver/game/interface/rpc_service/i9e"
	"github.com/izgib/tttserver/game/mocks"
	"github.com/izgib/tttserver/game/models"
	"github.com/izgib/tttserver/game/usecase"
	"github.com/izgib/tttserver/internal"
)

const bufSize = 1024 * 16

var wg sync.WaitGroup

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
	var mockLobbies = []struct {
		ID       int16
		settings models.GameSettings
		mark     models.PlayerMark
	}{
		{10,
			models.GameSettings{Rows: 3, Cols: 3, Win: 3},
			models.CrossMark,
		},
		{
			ID:       100,
			settings: models.GameSettings{3, 3, 3},
			mark:     models.NoughtMark,
		},
	}

	t.Run("success: Win-X", func(t *testing.T) {
		lobby := mockLobbies[1]
		l := bufconn.Listen(bufSize)
		s := grpc.NewServer(serveOpts...)
		recorder := text.NewGameRecorder(game.GameConfiguration{
			ID:       lobby.ID,
			Settings: lobby.settings,
			Mark:     lobby.mark,
		})
		mockLobby := usecase.NewGameLobby(lobby.ID, lobby.settings, lobby.mark, recorder)

		mockGameLobbyUsecase := &mocks.GameLobbyUsecase{}
		mockGameLobbyUsecase.On("CreateLobby", mock.AnythingOfType("game.GameConfiguration")).Return(mockLobby, nil).Once()
		mockGameLobbyUsecase.On("JoinLobby", mock.AnythingOfType("int16")).Return(mockLobby, nil).Once()

		service := NewGameService(
			mockGameLobbyUsecase,
			internal.CreateDebugLogger(),
		)
		i9e.RegisterGameConfiguratorServer(s, service)

		go func() {
			if err := s.Serve(l); err != nil {
				log.Error().Err(err).Msg("failed to serve")
				t.Error(err)
			}
		}()

		ctx := context.Background()
		conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithDialer(
			func(s string, duration time.Duration) (conn net.Conn, e error) {
				return l.Dial()
			}),
			grpc.WithInsecure(),
			grpc.WithDefaultCallOptions(grpc.CallCustomCodec(flatbuffers.FlatbuffersCodec{})),
		)
		if err != nil {
			t.Fatalf("Failed to dial bufnet: %v", err)
		}
		defer conn.Close()

		moves := []models.Move{{1, 1}, {0, 2}, {2, 2}, {0, 0}, {0, 1}, {2, 1}, {1, 0}, {1, 2}, {2, 0}}

		creator := i9e.NewGameConfiguratorClient(conn)
		crLogger := internal.CreateDebugLogger().With().Str("player", models.EnumNamesMoveChoice[models.MoveChoice(lobby.mark)]).Logger()
		crContext := context.Background()
		crStream, err := creator.CreateGame(crContext)
		if err != nil {
			t.Error(err)
		}

		wg.Add(2)
		go func() {
			if err := gameCreatorInit(crStream, lobby.settings, lobby.mark, crLogger); err != nil {
				t.Error(err)
			}
			crLogger.Debug().Msg("game started")
			if err := gameCrInteractor(crStream, moves, lobby.mark, crLogger); err != nil {
				t.Error(err)
			}
			wg.Done()
		}()

		opponent := i9e.NewGameConfiguratorClient(conn)
		oppMark := (lobby.mark + 1) & 1
		oppLogger := internal.CreateDebugLogger().With().Str("player", models.EnumNamesMoveChoice[models.MoveChoice(oppMark)]).Logger()
		oppContext := context.Background()
		oppStream, err := opponent.JoinGame(oppContext)
		if err != nil {
			t.Error(err)
		}
		go func() {
			if err := gameOpponentInit(oppStream, lobby.ID, oppLogger); err != nil {
				t.Error(err)
			}
			oppLogger.Debug().Msg("game started")
			if err := gameOppInteractor(oppStream, moves, oppMark, oppLogger); err != nil {
				t.Error(err)
			}
			wg.Done()
		}()

		wg.Wait()
	})
	t.Run("failure: disconnect-X", func(t *testing.T) {
		wantInterrupt := fmt.Errorf("got interruption event: %s", i9e.EnumNamesGameEventType[i9e.GameEventTypeOppDisconnected])

		lobby := mockLobbies[0]
		l := bufconn.Listen(bufSize)
		s := grpc.NewServer(serveOpts...)
		recorder := text.NewGameRecorder(game.GameConfiguration{
			ID:       lobby.ID,
			Settings: lobby.settings,
			Mark:     lobby.mark,
		})
		mockLobby := usecase.NewGameLobby(lobby.ID, lobby.settings, lobby.mark, recorder)

		mockGameLobbyUsecase := &mocks.GameLobbyUsecase{}
		mockGameLobbyUsecase.On("CreateLobby", mock.AnythingOfType("game.GameConfiguration")).Return(mockLobby, nil).Once()
		mockGameLobbyUsecase.On("JoinLobby", mock.AnythingOfType("int16")).Return(mockLobby, nil).Once()

		service := NewGameService(
			mockGameLobbyUsecase,
			internal.CreateDebugLogger(),
		)
		i9e.RegisterGameConfiguratorServer(s, service)

		go func() {
			if err := s.Serve(l); err != nil {
				log.Error().Err(err).Msg("failed to serve")
				t.Error(err)
			}
		}()

		ctx := context.Background()
		conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithDialer(
			func(s string, duration time.Duration) (conn net.Conn, e error) {
				return l.Dial()
			}),
			grpc.WithInsecure(),
			grpc.WithDefaultCallOptions(grpc.CallCustomCodec(flatbuffers.FlatbuffersCodec{})),
		)
		if err != nil {
			t.Fatalf("Failed to dial bufnet: %v", err)
		}
		defer conn.Close()

		moves := []models.Move{{1, 1}, {0, 2}, {2, 2}, {0, 0}, {0, 1}, {2, 1}, {1, 0}, {1, 2}, {2, 0}}

		creator := i9e.NewGameConfiguratorClient(conn)
		crLogger := internal.CreateDebugLogger().With().Str("player", models.EnumNamesMoveChoice[models.MoveChoice(lobby.mark)]).Logger()
		crContext, crStreamCancel := context.WithCancel(context.Background())
		crStream, err := creator.CreateGame(crContext)
		if err != nil {
			t.Error(err)
		}

		wg.Add(2)
		go func() {
			if err := gameCreatorInit(crStream, lobby.settings, lobby.mark, crLogger); err != nil {
				t.Error(err)
			}
			crLogger.Debug().Msg("game started")
			if err = gameCrInteractor(crStream, moves[:1], lobby.mark, crLogger); err != nil {
				t.Error(err)
			}
			crStreamCancel()
			wg.Done()
		}()

		opponent := i9e.NewGameConfiguratorClient(conn)
		oppMark := (lobby.mark + 1) & 1
		oppLogger := internal.CreateDebugLogger().With().Str("player", models.EnumNamesMoveChoice[models.MoveChoice(oppMark)]).Logger()
		oppContext := context.Background()
		oppStream, err := opponent.JoinGame(oppContext)
		if err != nil {
			t.Error(err)
		}
		go func() {
			if err := gameOpponentInit(oppStream, lobby.ID, oppLogger); err != nil {
				t.Error(err)
			}
			oppLogger.Debug().Msg("game started")
			if err := gameOppInteractor(oppStream, moves, oppMark, oppLogger); err != nil {
				if !assert.Equal(t, err.Error(), wantInterrupt.Error()) {
					t.Error(err)
				}
			}
			wg.Done()
		}()
		wg.Wait()
	})
	t.Run("failure: disconnect-O", func(t *testing.T) {
		wantInterrupt := fmt.Errorf("got interruption event: %s", i9e.EnumNamesGameEventType[i9e.GameEventTypeOppDisconnected])
		lobby := mockLobbies[0]
		l := bufconn.Listen(bufSize)
		s := grpc.NewServer(serveOpts...)
		recorder := text.NewGameRecorder(game.GameConfiguration{
			ID:       lobby.ID,
			Settings: lobby.settings,
			Mark:     lobby.mark,
		})
		mockLobby := usecase.NewGameLobby(lobby.ID, lobby.settings, lobby.mark, recorder)

		mockGameLobbyUsecase := &mocks.GameLobbyUsecase{}
		mockGameLobbyUsecase.On("CreateLobby", mock.AnythingOfType("game.GameConfiguration")).Return(mockLobby, nil).Once()
		mockGameLobbyUsecase.On("JoinLobby", mock.AnythingOfType("int16")).Return(mockLobby, nil).Once()

		service := NewGameService(
			mockGameLobbyUsecase,
			internal.CreateDebugLogger(),
		)
		i9e.RegisterGameConfiguratorServer(s, service)

		go func() {
			if err := s.Serve(l); err != nil {
				log.Error().Err(err).Msg("failed to serve")
				t.Error(err)
			}
		}()

		ctx := context.Background()
		conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithDialer(
			func(s string, duration time.Duration) (conn net.Conn, e error) {
				return l.Dial()
			}),
			grpc.WithInsecure(),
			grpc.WithDefaultCallOptions(grpc.CallCustomCodec(flatbuffers.FlatbuffersCodec{})),
		)
		if err != nil {
			t.Fatalf("Failed to dial bufnet: %v", err)
		}
		defer conn.Close()

		moves := []models.Move{{1, 1}, {0, 2}, {2, 2}, {0, 0}, {0, 1}, {2, 1}, {1, 0}, {1, 2}, {2, 0}}

		creator := i9e.NewGameConfiguratorClient(conn)
		crLogger := internal.CreateDebugLogger().With().Str("player", models.EnumNamesMoveChoice[models.MoveChoice(lobby.mark)]).Logger()
		crContext := context.Background()
		crStream, err := creator.CreateGame(crContext)
		if err != nil {
			t.Error(err)
		}

		wg.Add(2)
		go func() {
			if err := gameCreatorInit(crStream, lobby.settings, lobby.mark, crLogger); err != nil {
				t.Error(err)
			}
			crLogger.Debug().Msg("game started")
			if err = gameCrInteractor(crStream, moves, lobby.mark, crLogger); err != nil {
				if !assert.Equal(t, err.Error(), wantInterrupt.Error()) {
					t.Error(err)
				}
			}
			wg.Done()
		}()

		opponent := i9e.NewGameConfiguratorClient(conn)
		oppMark := (lobby.mark + 1) & 1
		oppLogger := internal.CreateDebugLogger().With().Str("player", models.EnumNamesMoveChoice[models.MoveChoice(oppMark)]).Logger()
		oppContext, oppStreamCancel := context.WithCancel(context.Background())
		oppStream, err := opponent.JoinGame(oppContext)
		if err != nil {
			t.Error(err)
		}
		go func() {
			if err := gameOpponentInit(oppStream, lobby.ID, oppLogger); err != nil {
				t.Error(err)
			}
			oppLogger.Debug().Msg("game started")
			if err := gameOppInteractor(oppStream, moves[:2], oppMark, oppLogger); err != nil {
				if !assert.Equal(t, err.Error(), wantInterrupt.Error()) {
					t.Error(err)
				}
			}
			oppStreamCancel()
			wg.Done()
		}()
		wg.Wait()
	})
}

func gameCreatorInit(stream i9e.GameConfigurator_CreateGameClient, settings models.GameSettings, mark models.PlayerMark, logger zerolog.Logger) error {
	b := flatbuffers.NewBuilder(0)

	i9e.GameParamsStart(b)
	i9e.GameParamsAddRows(b, settings.Rows)
	i9e.GameParamsAddCols(b, settings.Cols)
	i9e.GameParamsAddWin(b, settings.Win)
	i9e.GameParamsAddMark(b, i9e.MarkType(mark))
	gcParams := i9e.GameParamsEnd(b)
	crWrapRequest(b, i9e.CreatorReqMsgGameParams, gcParams)

	if err := stream.Send(b); err != nil {
		return err
	}

	out, err := stream.Recv()
	if err != nil {
		return err
	}

	if out.RespType() == i9e.CreatorRespMsgGameId {
		gameID := crResponseUnwrapperID(out)
		logger.Debug().Int16("game", gameID.ID()).Msg("got game")
	}

	out, err = stream.Recv()
	if err != nil {
		return err
	}

	if out.RespType() != i9e.CreatorRespMsgGameEvent {
		resp := &flatbuffers.Table{}
		gameStart := &i9e.GameEvent{}
		if out.Resp(resp) {
			gameStart.Init(resp.Bytes, resp.Pos)
			if gameStart.Type() != i9e.GameEventTypeGameStarted {
				err := fmt.Errorf("expected Event Game Started, got %s", i9e.EnumNamesGameEventType[gameStart.Type()])
				logger.Error().Err(err)
				return err
			}
		}
	}

	return nil
}

func gameOpponentInit(stream i9e.GameConfigurator_JoinGameClient, ID int16, logger zerolog.Logger) error {
	b := flatbuffers.NewBuilder(0)

	i9e.GameIdStart(b)
	i9e.GameIdAddID(b, ID)
	oppRequestWrapper(b, i9e.OpponentReqMsgGameId, i9e.GameIdEnd(b))

	if err := stream.Send(b); err != nil {
		return err
	}

	out, err := stream.Recv()
	if err != nil {
		return err
	}

	if out.RespType() != i9e.OpponentRespMsgGameEvent {
		resp := &flatbuffers.Table{}
		gameStart := &i9e.GameEvent{}
		if out.Resp(resp) {
			gameStart.Init(resp.Bytes, resp.Pos)
			if gameStart.Type() != i9e.GameEventTypeGameStarted {
				err := fmt.Errorf("expected Event Game Started, got %s", i9e.EnumNamesGameEventType[gameStart.Type()])
				logger.Error().Err(err)
				return err
			}
		}
	}

	return nil

}

func sendMove(i interface{}, move models.Move) error {
	b := flatbuffers.NewBuilder(0)
	i9e.MoveStart(b)
	i9e.MoveAddRow(b, move.I)
	i9e.MoveAddCol(b, move.J)

	switch stream := i.(type) {
	case i9e.GameConfigurator_CreateGameClient:
		crWrapRequest(b, i9e.CreatorReqMsgMove, i9e.MoveEnd(b))
		return stream.Send(b)
	case i9e.GameConfigurator_JoinGameClient:
		oppWrapRequest(b, i9e.OpponentReqMsgMove, i9e.MoveEnd(b))
		return stream.Send(b)
	}
	return nil
}

func gameCrInteractor(stream i9e.GameConfigurator_CreateGameClient, moves []models.Move, mark models.PlayerMark, logger zerolog.Logger) error {
	mask := int(mark)
	turn := 0
	for _, move := range moves {
		var event *i9e.GameEvent
		if (turn & 1) == mask {
			if err := sendMove(stream, move); err != nil {
				return err
			}
			in, err := stream.Recv()
			if err != nil {
				return err
			}
			event = crResponseUnwrapperEvent(in)
		} else {
		moveLoopCr:
			for j := 0; j < 2; j++ {
				in, err := stream.Recv()
				if err != nil {
					return err
				}
				switch in.RespType() {
				case i9e.CreatorRespMsgMove:
					_ = crResponseUnwrapperMove(in)
				case i9e.CreatorRespMsgGameEvent:
					event = crResponseUnwrapperEvent(in)
					break moveLoopCr
				}
			}
		}
		switch eventType := event.Type(); eventType {
		case i9e.GameEventTypeOK:
			turn++
		case i9e.GameEventTypeTie, i9e.GameEventTypeWin:
			if event.Type() == i9e.GameEventTypeWin {
				winLine := event.FollowUp(nil)
				start := winLine.Start(nil)
				end := winLine.End(nil)
				logger.Debug().Str("Winner", models.EnumNamesMoveChoice[models.MoveChoice(winLine.Mark())]).Dict("win line", zerolog.Dict().
					Dict("start", zerolog.Dict().
						Int16("i", start.Row()).
						Int16("j", start.Col())).
					Dict("end", zerolog.Dict().
						Int16("i", end.Row()).
						Int16("j", end.Col())),
				).Msg("game ended")
			} else {
				logger.Debug().Str("result", "Tie").Msg("game ended")
			}
			return nil
		default:
			err := fmt.Errorf("got interruption event: %s", i9e.EnumNamesGameEventType[eventType])
			logger.Error().Str("type", i9e.EnumNamesGameEventType[eventType]).Msg("got interruption event")
			return err
		}
	}
	return nil
}

func gameOppInteractor(stream i9e.GameConfigurator_JoinGameClient, moves []models.Move, mark models.PlayerMark, logger zerolog.Logger) error {
	var in *i9e.OppResponse
	var err error
	var event *i9e.GameEvent
	mask := int(mark)
	turn := 0

	for _, move := range moves {
		if (turn & 1) == mask {
			if err := sendMove(stream, move); err != nil {
				return err
			}
			in, err := stream.Recv()
			if err != nil {
				return err
			}
			if in.RespType() == i9e.OpponentRespMsgGameEvent {
				event = oppResponseUnwrapperEvent(in)
			} else {
				panic("wtf")
			}
		} else {
		moveLoopOpp:
			for j := 0; j < 2; j++ {
				logger.Debug().Msg("attempt to receive move")
				in, err = stream.Recv()
				if err != nil {
					return err
				}
				switch in.RespType() {
				case i9e.OpponentRespMsgMove:
					_ = oppResponseUnwrapperMove(in)
				case i9e.OpponentRespMsgGameEvent:
					event = oppResponseUnwrapperEvent(in)
					break moveLoopOpp
				}
			}
		}
		switch eventType := event.Type(); eventType {
		case i9e.GameEventTypeOK:
			turn++
			logger.Debug().Msg("next turn")
		case i9e.GameEventTypeTie, i9e.GameEventTypeWin:
			if event.Type() == i9e.GameEventTypeWin {
				winLine := event.FollowUp(nil)
				start := winLine.Start(nil)
				end := winLine.End(nil)
				logger.Debug().Str("Winner", models.EnumNamesMoveChoice[models.MoveChoice(winLine.Mark())]).Dict("win line", zerolog.Dict().
					Dict("start", zerolog.Dict().
						Int16("i", start.Row()).
						Int16("j", start.Col())).
					Dict("end", zerolog.Dict().
						Int16("i", end.Row()).
						Int16("j", end.Col())),
				).Msg("game ended")
			} else {
				logger.Debug().Str("result", "Tie").Msg("game ended")
			}
			return nil
		default:
			err := fmt.Errorf("got interruption event: %s", i9e.EnumNamesGameEventType[eventType])
			logger.Error().Str("type", i9e.EnumNamesGameEventType[eventType]).Msg("got interruption event")
			return err
		}
	}
	return nil
}

func oppRequestWrapper(b *flatbuffers.Builder, reqType i9e.OpponentReqMsg, req flatbuffers.UOffsetT) {
	i9e.OppRequestStart(b)
	i9e.OppRequestAddReqType(b, reqType)
	i9e.OppRequestAddReq(b, req)
	b.Finish(i9e.OppRequestEnd(b))
}

func crResponseUnwrapperMove(resp *i9e.CrResponse) *i9e.Move {
	respT := &flatbuffers.Table{}
	move := &i9e.Move{}
	if resp.Resp(respT) {
		move.Init(respT.Bytes, respT.Pos)
	}
	return move
}

func crResponseUnwrapperEvent(resp *i9e.CrResponse) *i9e.GameEvent {
	respT := &flatbuffers.Table{}
	event := &i9e.GameEvent{}
	if resp.Resp(respT) {
		event.Init(respT.Bytes, respT.Pos)
	}
	return event
}

func crResponseUnwrapperID(resp *i9e.CrResponse) *i9e.GameId {
	respT := &flatbuffers.Table{}
	gameID := &i9e.GameId{}
	if resp.Resp(respT) {
		gameID.Init(respT.Bytes, respT.Pos)
	}
	return gameID
}

func oppResponseUnwrapperMove(resp *i9e.OppResponse) *i9e.Move {
	respT := &flatbuffers.Table{}
	move := &i9e.Move{}
	if resp.Resp(respT) {
		move.Init(respT.Bytes, respT.Pos)
	}
	return move
}

func oppResponseUnwrapperEvent(resp *i9e.OppResponse) *i9e.GameEvent {
	respT := &flatbuffers.Table{}
	event := &i9e.GameEvent{}
	if resp.Resp(respT) {
		event.Init(respT.Bytes, respT.Pos)
	}
	return event
}

func createRange(b *flatbuffers.Builder, start int16, end int16) flatbuffers.UOffsetT {
	i9e.RangeStart(b)
	i9e.RangeAddStart(b, start)
	i9e.RangeAddEnd(b, end)
	return i9e.RangeEnd(b)
}

func crWrapRequest(b *flatbuffers.Builder, respType i9e.CreatorReqMsg, resp flatbuffers.UOffsetT) {
	i9e.CrRequestStart(b)
	i9e.CrRequestAddReqType(b, respType)
	i9e.CrRequestAddReq(b, resp)
	b.Finish(i9e.CrRequestEnd(b))
}

func oppWrapRequest(b *flatbuffers.Builder, respType i9e.OpponentReqMsg, resp flatbuffers.UOffsetT) {
	i9e.OppRequestStart(b)
	i9e.OppRequestAddReqType(b, respType)
	i9e.OppRequestAddReq(b, resp)
	b.Finish(i9e.OppRequestEnd(b))
}
