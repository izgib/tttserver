package usecase

import (
	"fmt"
	"github.com/izgib/tttserver/game/interface/recorder/text"
	"reflect"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"

	"github.com/izgib/tttserver/game"
	"github.com/izgib/tttserver/game/models"
)

func min(a, b int16) int16 {
	if a <= b {
		return a
	} else {
		return b
	}
}

func Test_CreateLobby(t *testing.T) {
	var wrongSettings = []struct {
		name     string
		settings models.GameSettings
		mark     models.PlayerMark
	}{
		{
			name:     "wrong row low bound",
			settings: models.GameSettings{models.GameRules[models.Rows].Start - 1, 3, 3},
			mark:     models.CrossMark,
		},
		{
			name:     "wrong row high bound",
			settings: models.GameSettings{models.GameRules[models.Rows].End + 1, 3, 3},
			mark:     models.CrossMark,
		},
		{
			name:     "wrong col low bound",
			settings: models.GameSettings{models.GameRules[models.Cols].Start - 1, 3, 3},
			mark:     models.CrossMark,
		},
		{
			name:     "wrong col high bound",
			settings: models.GameSettings{models.GameRules[models.Cols].End + 1, 3, 3},
			mark:     models.CrossMark,
		},
		{
			name:     "wrong win>row",
			settings: models.GameSettings{4, 5, 5},
			mark:     models.CrossMark,
		},
		{
			name:     "wrong win>col",
			settings: models.GameSettings{5, 4, 5},
			mark:     models.CrossMark,
		},
	}

	settings := models.GameSettings{Rows: 3, Cols: 3, Win: 3}
	CreatorMark := models.CrossMark

	u := NewGameLobbyUsecase(text.NewGameRecorder, &zerolog.Logger{})
	t.Run("success", func(t *testing.T) {
		lobby, err := u.CreateLobby(game.GameConfiguration{
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
			_, err := u.CreateLobby(game.GameConfiguration{
				Settings: test.settings,
				Mark:     test.mark,
			})
			assert.Error(t, err)
		})
	}
}

func Test_gameLobbyUsecase_JoinLobby(t *testing.T) {
	mockLobby := &gameLobby{
		ID:          10,
		Settings:    models.GameSettings{3, 3, 3},
		CreatorMark: models.CrossMark,
	}
	recorder := text.NewGameRecorder(game.GameConfiguration{
		ID:       mockLobby.ID,
		Settings: mockLobby.Settings,
		Mark:     mockLobby.CreatorMark,
	})
	mockLobby.recorder = recorder
	lobbies := map[int16]game.GameLobby{mockLobby.ID: mockLobby}
	u := gameLobbyUsecase{
		recordCreator: text.NewGameRecorder,
		lobbies:       lobbies,
		logger:        &zerolog.Logger{},
	}
	t.Run("succes", func(t *testing.T) {
		lobby, err := u.JoinLobby(10)
		if !reflect.DeepEqual(lobby, mockLobby) {
			t.Error("wrong lobby returned")
		}
		assert.NoError(t, err)
	})

	t.Run("error: not found", func(t *testing.T) {
		/*		errString := fmt.Sprintf("GameConfiguration %d was not found", 11)
				repoErr := fmt.Errorf(errString)*/
		_, err := u.JoinLobby(11)
		assert.Error(t, err)
	})
}
