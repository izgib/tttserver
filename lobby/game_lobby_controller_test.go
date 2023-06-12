package lobby

import (
	"context"
	"fmt"
	"github.com/izgib/tttserver/internal/logger"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/izgib/tttserver/game"
)

func Test_CreateLobby(t *testing.T) {
	var wrongSettings = []struct {
		name     string
		settings game.GameSettings
		mark     game.PlayerMark
	}{
		{
			name:     "wrong row low bound",
			settings: game.GameSettings{game.GameRules[game.Rows].Start - 1, 3, 3},
			mark:     game.CrossMark,
		},
		{
			name:     "wrong row high bound",
			settings: game.GameSettings{game.GameRules[game.Rows].End + 1, 3, 3},
			mark:     game.CrossMark,
		},
		{
			name:     "wrong col low bound",
			settings: game.GameSettings{game.GameRules[game.Cols].Start - 1, 3, 3},
			mark:     game.CrossMark,
		},
		{
			name:     "wrong col high bound",
			settings: game.GameSettings{game.GameRules[game.Cols].End + 1, 3, 3},
			mark:     game.CrossMark,
		},
		{
			name:     "wrong win>row",
			settings: game.GameSettings{4, 5, 5},
			mark:     game.CrossMark,
		},
		{
			name:     "wrong win>col",
			settings: game.GameSettings{5, 4, 5},
			mark:     game.CrossMark,
		},
	}

	settings := game.GameSettings{Rows: 3, Cols: 3, Win: 3}
	CreatorMark := game.CrossMark

	u := NewGameLobbyController(PlainGameRecorder, &zerolog.Logger{})
	t.Run("success", func(t *testing.T) {
		lobby, err := u.CreateLobby(GameConfiguration{
			Settings: settings,
			Mark:     CreatorMark,
		})
		if !(reflect.DeepEqual(lobby.Settings(), settings) || lobby.CreatorMark() == CreatorMark) {
			t.Error("wrong lobby returned")
		}
		assert.NoError(t, err)
	})
	for _, test := range wrongSettings {
		t.Run(fmt.Sprintf("error: %s", test.name), func(t *testing.T) {
			_, err := u.CreateLobby(GameConfiguration{
				Settings: test.settings,
				Mark:     test.mark,
			})
			assert.Error(t, err)
		})
	}
}

func Test_GameLobbyController_JoinLobby(t *testing.T) {
	lobbySettings := game.GameSettings{3, 3, 3}
	createMark := game.CrossMark

	log := logger.CreateDebugLogger()
	controller := NewGameLobbyController(PlainGameRecorder, log)
	testLobby, err := controller.CreateLobby(GameConfiguration{
		Settings: lobbySettings,
		Mark:     createMark,
	})
	if err != nil {
		log.Err(err)
	}

	t.Run("lil", func(t *testing.T) {
		move := game.Move{
			I: 1,
			J: 1,
		}
		m := &game.Move{
			I: 1,
			J: 1,
		}
		if reflect.DeepEqual(m, move) {
			t.Fail()
		}
	})

	t.Run("found", func(t *testing.T) {
		lobby, err := controller.JoinLobby(testLobby.ID)
		if !reflect.DeepEqual(lobby, testLobby) {
			t.Error("wrong lobby returned")
		}
		assert.NoError(t, err)
	})
	var randomID = rand.Uint32()
	for randomID == testLobby.ID {
		randomID = rand.Uint32()
	}

	t.Run("error: not found", func(t *testing.T) {
		_, err := controller.JoinLobby(randomID)
		assert.Error(t, err)
	})

	testLogger := logger.CreateDebugLogger()
	t.Run("lobby cycle check", func(t *testing.T) {
		config := GameConfiguration{
			Settings: game.GameSettings{Rows: 3, Cols: 3, Win: 3},
			Mark:     game.CrossMark,
		}
		moves := []game.Move{{1, 1}, {0, 2}, {2, 2}, {0, 0}, {0, 1}, {2, 1}, {1, 0}, {1, 2}, {2, 0}}
		lobby, err := controller.CreateLobby(config)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		lobby.CreatorReadyChan() <- true
		testLogger.Debug().Msg("after creating")

		jLobby, jErr := controller.JoinLobby(lobby.ID)
		if jErr != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		jLobby.OpponentReadyChan() <- true
		testLogger.Debug().Msg("after joining")

		ctx, _ := context.WithTimeout(context.Background(), 200*time.Millisecond)
		group, _ := errgroup.WithContext(ctx)
		group.Go(func() error {
			if !(<-lobby.GameStartedChan()) {
				return fmt.Errorf("game do not start")
			}
			testLogger.Debug().Msg("creator started")
			return player(moves, lobby.CreatorMark(), lobby.GetGameController().CreatorChannels())
		})
		testLogger.Info().Bool("equal", lobby == jLobby).Msg("lobby check")
		group.Go(func() error {
			if !(<-jLobby.GameStartedChan()) {
				return fmt.Errorf("game do not start")
			}
			testLogger.Debug().Msg("opponent started")
			return player(moves, jLobby.OpponentMark(), lobby.GetGameController().OpponentChannels())
		})

		if err = group.Wait(); err != nil {
			t.Fatal(err)
		}

	})
}

func player(moves []game.Move, mark game.PlayerMark, chans *CommChannels) error {
	log := logger.CreateDebugLogger().With().Str("player", game.EnumNamesMoveChoice[game.MoveChoice(mark)]).Logger()
	log.Debug().Msg("test")
	turn := 0
loop:
	for {
		action := <-chans.Actions
		switch a := action.(type) {
		case Step:
			tl := log.Debug().Dict("move", zerolog.Dict().
				Uint16("i", moves[turn].I).
				Uint16("j", moves[turn].J))
			if a.Pos != nil {
				tl.Dict("move", zerolog.Dict().
					Uint16("i", moves[turn].I).
					Uint16("j", moves[turn].J))
				if !reflect.DeepEqual(*a.Pos, moves[turn]) {
					err := fmt.Errorf("expected move %v, but got %v", moves[turn], a.Pos)
					log.Error().Err(err).Msg("invalid move")
				}
			}
			tl.Interface("state", a.State).Msg("received")
			switch a.State.StateType {
			case game.Won:
				m := game.EnumNamesMoveChoice[game.MoveChoice(a.State.WinLine.Mark)]
				log.Debug().Str("Winner", m).Dict("win line", zerolog.Dict().
					Dict("start", zerolog.Dict().
						Uint16("i", a.State.WinLine.Start.I).
						Uint16("j", a.State.WinLine.Start.J)).
					Dict("end", zerolog.Dict().
						Uint16("i", a.State.WinLine.End.I).
						Uint16("j", a.State.WinLine.End.J)),
				).Msg("game won")
				break loop
			case game.Tie:
				log.Debug().Msg("tie")
				break loop
			case game.Running:
				turn++
				log.Debug().Msg("next turn")
			}
			chans.Reactions <- Status{Err: nil}
			log.Debug().Msg("reaction sent")
		case ReceiveMove:
			chans.Reactions <- ReceivedMove{
				Pos: &moves[turn],
			}
			log.Debug().Dict("move", zerolog.Dict().
				Uint16("i", moves[turn].I).
				Uint16("j", moves[turn].J)).Msg("sent")
		default:
			return fmt.Errorf("unexpected type: %v", a)
		}
	}
	return nil
}
