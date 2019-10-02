// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import game "github.com/izgib/tttserver/game"
import mock "github.com/stretchr/testify/mock"

// GameLobbyUsecase is an autogenerated mock type for the GameLobbyUsecase type
type GameLobbyUsecase struct {
	mock.Mock
}

// CreateLobby provides a mock function with given fields: config
func (_m *GameLobbyUsecase) CreateLobby(config game.GameConfiguration) (game.GameLobby, error) {
	ret := _m.Called(config)

	var r0 game.GameLobby
	if rf, ok := ret.Get(0).(func(game.GameConfiguration) game.GameLobby); ok {
		r0 = rf(config)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(game.GameLobby)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(game.GameConfiguration) error); ok {
		r1 = rf(config)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeleteLobby provides a mock function with given fields: lobby
func (_m *GameLobbyUsecase) DeleteLobby(lobby game.GameLobby) {
	_m.Called(lobby)
}

// JoinLobby provides a mock function with given fields: ID
func (_m *GameLobbyUsecase) JoinLobby(ID int16) (game.GameLobby, error) {
	ret := _m.Called(ID)

	var r0 game.GameLobby
	if rf, ok := ret.Get(0).(func(int16) game.GameLobby); ok {
		r0 = rf(ID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(game.GameLobby)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(int16) error); ok {
		r1 = rf(ID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListLobbies provides a mock function with given fields: filter
func (_m *GameLobbyUsecase) ListLobbies(filter game.GameFilter) ([]game.GameConfiguration, error) {
	ret := _m.Called(filter)

	var r0 []game.GameConfiguration
	if rf, ok := ret.Get(0).(func(game.GameFilter) []game.GameConfiguration); ok {
		r0 = rf(filter)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]game.GameConfiguration)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(game.GameFilter) error); ok {
		r1 = rf(filter)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
