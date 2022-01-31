package lobby_test

import (
	"fmt"
	"github.com/izgib/tttserver/base"
	"github.com/izgib/tttserver/internal"
	"github.com/izgib/tttserver/lobby"
	"github.com/izgib/tttserver/recorder/text"
	"math/rand"
	"reflect"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"

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

	u := lobby.NewGameLobbyController(text.NewGameRecorder, &zerolog.Logger{})
	t.Run("success", func(t *testing.T) {
		lobby, err := u.CreateLobby(base.GameConfiguration{
			Settings: settings,
			Mark:     CreatorMark,
		})
		if !(reflect.DeepEqual(lobby.GetSettings(), settings) || lobby.GetCreatorMark() == CreatorMark) {
			t.Error("wrong lobby returned")
		}
		assert.NoError(t, err)
	})
	for _, test := range wrongSettings {
		t.Run(fmt.Sprintf("error: %s", test.name), func(t *testing.T) {
			_, err := u.CreateLobby(base.GameConfiguration{
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

	log := internal.CreateDebugLogger()
	controller := lobby.NewGameLobbyController(text.NewGameRecorder, log)
	testLobby, err := controller.CreateLobby(base.GameConfiguration{
		Settings: lobbySettings,
		Mark:     createMark,
	})
	if err != nil {
		log.Err(err)
	}

	t.Run("success", func(t *testing.T) {
		lobby, err := controller.JoinLobby(testLobby.GetID())
		if !reflect.DeepEqual(lobby, testLobby) {
			t.Error("wrong lobby returned")
		}
		assert.NoError(t, err)
	})
	var randomID = int16(rand.Int31() & 0xffff)
	for randomID == testLobby.GetID() {
		randomID = int16(rand.Int31() & 0xffff)
	}

	t.Run("error: not found", func(t *testing.T) {
		_, err := controller.JoinLobby(randomID)
		assert.Error(t, err)
	})
}
