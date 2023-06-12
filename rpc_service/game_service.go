package rpc_service

import (
	"github.com/izgib/tttserver/game"
	gLobby "github.com/izgib/tttserver/lobby"
	"github.com/izgib/tttserver/rpc_service/transport"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
)

type GameService struct {
	lobbiesController gLobby.GameLobbyController
	logger            *zerolog.Logger
	transport.UnimplementedGameConfiguratorServer
}

var enumPlayerMark = map[transport.MarkType]game.PlayerMark{
	transport.MarkType_MARK_TYPE_CROSS:  game.CrossMark,
	transport.MarkType_MARK_TYPE_NOUGHT: game.NoughtMark,
}

var enumMoveChoise = map[transport.MarkType]game.MoveChoice{
	transport.MarkType_MARK_TYPE_CROSS:       game.Cross,
	transport.MarkType_MARK_TYPE_NOUGHT:      game.Nought,
	transport.MarkType_MARK_TYPE_UNSPECIFIED: game.Empty,
}

var enumMarkType = map[game.PlayerMark]transport.MarkType{
	game.CrossMark:  transport.MarkType_MARK_TYPE_CROSS,
	game.NoughtMark: transport.MarkType_MARK_TYPE_NOUGHT,
}

// GetListOfGames returns list of Games
func (s *GameService) GetListOfGames(params *transport.GameFilter, stream transport.GameConfigurator_GetListOfGamesServer) error {
	rowsStart := params.GetRows().Start
	rowsEnd := params.GetRows().End

	colsStart := params.GetCols().Start
	colsEnd := params.GetCols().End

	winStart := params.GetWin().Start
	winEnd := params.GetWin().End

	s.logger.Debug().Dict("filter", zerolog.Dict().
		Dict("row", zerolog.Dict().
			Uint32("start", rowsStart).
			Uint32("end", rowsEnd)).
		Dict("col", zerolog.Dict().
			Uint32("start", colsStart).
			Uint32("end", colsEnd)).
		Dict("win", zerolog.Dict().
										Uint32("start", winStart).
										Uint32("end", winEnd)).
		Str("mark", transport.MarkType_name[int32(params.Mark)]), //transport.EnumNamesMarkTypeFilter[params.Mark()]),
	).Msg("Request List of Games")

	filter := gLobby.GameFilter{
		Rows: gLobby.Filter{Start: uint16(rowsStart), End: uint16(rowsEnd)},
		Cols: gLobby.Filter{Start: uint16(colsStart), End: uint16(colsEnd)},
		Win:  gLobby.Filter{Start: uint16(winStart), End: uint16(winEnd)},
		Mark: enumMoveChoise[params.Mark],
	}

	games, err := s.lobbiesController.ListLobbies(filter)
	if err != nil {
		return err
	}

	i := 0
	for _, v := range games {
		s.logger.Debug().Uint32("game", v.ID).Msgf("%d:", i)
		params := transport.GameParams{
			Rows: uint32(v.Settings.Rows),
			Cols: uint32(v.Settings.Cols),
			Win:  uint32(v.Settings.Win),
			Mark: enumMarkType[v.Mark],
		}

		item := transport.ListItem{
			Id:     v.ID,
			Params: &params,
		}

		if err = stream.Send(&item); err != nil {
			return err
		}
		i++
	}

	return nil
}

func NewGameService(gameController gLobby.GameLobbyController, logger *zerolog.Logger) *GameService {
	s := &GameService{
		lobbiesController: gameController,
		logger:            logger,
	}
	return s
}

func (s *GameService) CreateGame(stream transport.GameConfigurator_CreateGameServer) error {
	in, err := stream.Recv()
	if err != nil {
		return err
	}

	var lobby *gLobby.GameLobby

	params, ok := in.Payload.(*transport.CreateRequest_Params)
	if !ok {
		err = status.Error(codes.InvalidArgument, "Expected to get GameParams")
		s.logger.Err(err)
		return err
	}
	rows := uint16(params.Params.Rows)
	cols := uint16(params.Params.Cols)
	win := uint16(params.Params.Win)

	s.logger.Debug().Uint16("rows", rows).
		Uint16("cols", cols).
		Uint16("win", win).
		Str("mark", transport.MarkType_name[int32(params.Params.Mark)]).
		Msg("Request of creating game lobby")

	gameConfig := gLobby.GameConfiguration{
		Settings: game.GameSettings{
			Rows: rows,
			Cols: cols,
			Win:  win,
		},
		Mark: enumPlayerMark[params.Params.Mark],
	}

	lobby, err = s.lobbiesController.CreateLobby(gameConfig)
	if err != nil {
		s.logger.Error().Err(err).Msg("got error")
		statusError := status.Error(codes.OutOfRange, err.Error())
		return statusError
	}

	lobby.CreatorReadyChan() <- true

	gameId := transport.CreateResponse{
		Payload: &transport.CreateResponse_GameId{
			GameId: lobby.ID,
		},
		Move: nil,
	}

	if err = stream.Send(&gameId); err != nil {
		lobby.CreatorReadyChan() <- false
		s.logger.Debug().Err(err)
		return err
	}

	gameLogger := game.CreateGameDebugLogger(lobby.ID).With().
		Str("player", game.EnumNamesMoveChoice[game.MoveChoice(lobby.CreatorMark())]).Logger()

	gameLogger.Debug().Msg("initialized")

	var Chans *gLobby.CommChannels
	select {
	case gs := <-lobby.GameStartedChan():
		if gs {
			Chans = lobby.GetGameController().CreatorChannels()
			started := transport.CreateResponse{
				Payload: &transport.CreateResponse_Status{Status: transport.GameStatus_GAME_STATUS_GAME_STARTED},
				Move:    nil,
			}

			if err := stream.Send(&started); err != nil {
				gameLogger.Error().Err(err).Msg("disconnected after creation")
				Chans.Err = err
				close(Chans.Done)
				return err
			}
			gameLogger.Debug().Msg("event started have sent")
		} else {
			return nil
		}
	case <-stream.Context().Done():
		gameLogger.Error().Err(stream.Context().Err()).Msg("canceled")
		lobby.CreatorReadyChan() <- false
		return stream.Context().Err()
	}
	gameLogger.Debug().Msg("started")

	for {
		select {
		case action := <-Chans.Actions:
			var resp = &transport.CreateResponse{}
			switch a := action.(type) {
			case gLobby.Step:
				if a.Pos != nil {
					resp.Move = moveToTransport(a.Pos)
				}
				switch a.State.StateType {
				case game.Won:
					var start *transport.Move = nil
					var end *transport.Move = nil
					if a.State.WinLine != nil {
						start = moveToTransport(a.State.WinLine.Start)
						end = moveToTransport(a.State.WinLine.End)
					}
					resp.Payload = &transport.CreateResponse_WinLine{WinLine: &transport.WinLine{
						Mark:  enumMarkType[a.State.WinLine.Mark],
						Start: start,
						End:   end,
					}}
				case game.Tie:
					resp.Payload = &transport.CreateResponse_Status{Status: transport.GameStatus_GAME_STATUS_TIE}
				case game.Running:
					resp.Payload = &transport.CreateResponse_Status{Status: transport.GameStatus_GAME_STATUS_OK}
				}

				err = stream.Send(resp)
				Chans.Reactions <- gLobby.Status{Err: err}
				if err != nil {
					gameLogger.Debug().Dict("move", zerolog.Dict().
						Uint16("i", a.Pos.I).
						Uint16("j", a.Pos.J),
					).Err(err).Msg("can not send move")
					return err
				}
				if a.State.StateType != game.Running {
					gameLogger.Info().Msg("ended")
				}
			case gLobby.Interruption:
				return interrupt(gameLogger, a.Cause)
			case gLobby.ReceiveMove:
				go func() {
					in, err = stream.Recv()
					if err != nil {
						gameLogger.Error().Err(err).Msg("can not receive move")
						Chans.Reactions <- gLobby.ReceivedMove{Err: err}
						return
					}

					switch payload := in.Payload.(type) {
					case *transport.CreateRequest_Move:
						Chans.Reactions <- gLobby.ReceivedMove{Pos: &game.Move{
							I: uint16(payload.Move.Row), J: uint16(payload.Move.Col),
						}}
					case *transport.CreateRequest_Action:
						var cause gLobby.InterruptionCause
						switch payload.Action {
						case transport.ClientAction_CLIENT_ACTION_GIVE_UP:
							cause = gLobby.GiveUp
						case transport.ClientAction_CLIENT_ACTION_LEAVE:
							cause = gLobby.Leave
						}
						Chans.Reactions <- gLobby.Interruption{Cause: cause}
						return
					default:
						err = status.Error(codes.InvalidArgument, "unexpected type")
					}
				}()
			}
		case <-stream.Context().Done():
			ctxErr := status.FromContextError(stream.Context().Err()).Err()
			gameLogger.Error().Err(ctxErr).Msg("connection canceled")
			Chans.Err = ctxErr
			close(Chans.Done)
			return stream.Context().Err()
		}
	}
}

func (s *GameService) JoinGame(stream transport.GameConfigurator_JoinGameServer) error {
	var Chans *gLobby.CommChannels

	in, err := stream.Recv()
	if err != nil {
		if err == io.EOF {
			return nil
		}
		return err
	}

	var lobby *gLobby.GameLobby

	id, ok := in.Payload.(*transport.JoinRequest_GameId)
	if !ok {
		err = status.Error(codes.InvalidArgument, "Expected to get GameID")
		s.logger.Error().Err(err).Msg("wrong type")
		return err
	}

	s.logger.Debug().Uint32("game", id.GameId).Msg("opponent request to join")
	lobby, err = s.lobbiesController.JoinLobby(id.GameId)
	if err != nil {
		statusErr := status.Error(codes.NotFound, err.Error())
		s.logger.Info().Err(err).Msg("not found")
		return statusErr
	}

	gameLogger := game.CreateGameDebugLogger(lobby.ID).With().Str("player", game.EnumNamesMoveChoice[game.MoveChoice(lobby.OpponentMark())]).Logger()
	lobby.OpponentReadyChan() <- true

	if <-lobby.GameStartedChan() {
		started := transport.JoinResponse{
			Payload: &transport.JoinResponse_Status{Status: transport.GameStatus_GAME_STATUS_GAME_STARTED},
			Move:    nil,
		}

		if err := stream.Send(&started); err != nil {
			gameLogger.Error().Err(err).Msg("disconnected after creation")
			Chans.Err = err
			close(Chans.Done)
			s.logger.Err(err).Uint32("game", lobby.ID).Msg("can not send game start event")
			return err
		}
		gameLogger.Debug().Msg("event started have sent")
	} else {
		return nil
	}

	gameLogger.Debug().Msg("started")
	Chans = lobby.GetGameController().OpponentChannels()

	for {
		select {
		case action := <-Chans.Actions:
			var resp = &transport.JoinResponse{}
			switch a := action.(type) {
			case gLobby.Step:
				resp.Move = moveToTransport(a.Pos)
				switch a.State.StateType {
				case game.Won:
					resp.Payload = &transport.JoinResponse_WinLine{WinLine: &transport.WinLine{
						Mark:  enumMarkType[a.State.WinLine.Mark],
						Start: moveToTransport(a.State.WinLine.Start),
						End:   moveToTransport(a.State.WinLine.End),
					}}
				case game.Tie:
					resp.Payload = &transport.JoinResponse_Status{Status: transport.GameStatus_GAME_STATUS_TIE}
				case game.Running:
					resp.Payload = &transport.JoinResponse_Status{Status: transport.GameStatus_GAME_STATUS_OK}
				}

				err = stream.Send(resp)
				Chans.Reactions <- gLobby.Status{Err: err}
				if err != nil {
					gameLogger.Debug().Dict("move", zerolog.Dict().
						Uint16("i", a.Pos.I).
						Uint16("j", a.Pos.J),
					).Err(err).Msg("can not send move")
					return err
				}
				if a.State.StateType != game.Running {
					gameLogger.Info().Msg("ended")
					//return nil
				}
			case gLobby.Interruption:
				return interrupt(gameLogger, a.Cause)
			case gLobby.ReceiveMove:
				go func() {
					in, err = stream.Recv()
					if err != nil {
						gameLogger.Error().Err(err).Msg("can not receive move")
						Chans.Reactions <- gLobby.ReceivedMove{Err: err}
						return
					}

					switch payload := in.Payload.(type) {
					case *transport.JoinRequest_Move:
						Chans.Reactions <- gLobby.ReceivedMove{Pos: &game.Move{
							I: uint16(payload.Move.Row), J: uint16(payload.Move.Col),
						}}
					case *transport.JoinRequest_Action:
						var cause gLobby.InterruptionCause
						switch payload.Action {
						case transport.ClientAction_CLIENT_ACTION_GIVE_UP:
							cause = gLobby.GiveUp
						case transport.ClientAction_CLIENT_ACTION_LEAVE:
							cause = gLobby.Leave
						}
						Chans.Reactions <- gLobby.Interruption{Cause: cause}
						return
					default:
						err = status.Error(codes.InvalidArgument, "unexpected type")
					}
				}()
			}
		case <-stream.Context().Done():
			ctxErr := status.FromContextError(stream.Context().Err()).Err()
			gameLogger.Error().Err(ctxErr).Msg("connection canceled")
			Chans.Err = ctxErr
			close(Chans.Done)
			return stream.Context().Err()
		}
	}
}

func moveToTransport(move *game.Move) *transport.Move {
	if move == nil {
		return nil
	}
	return &transport.Move{
		Row: uint32(move.I),
		Col: uint32(move.J),
	}
}

func interruptionToStatus(interruptCause gLobby.InterruptionCause) (*status.Status, error) {
	var msg string
	var cause transport.StopCause
	switch interruptCause {
	case gLobby.Internal:
		msg = "unknown error"
		cause = transport.StopCause_STOP_CAUSE_INTERNAL
	case gLobby.InvalidMove:
		msg = "invalid move"
		cause = transport.StopCause_STOP_CAUSE_INVALID_MOVE
	case gLobby.Disconnect:
		cause = transport.StopCause_STOP_CAUSE_DISCONNECT
		msg = "opponent disconnected"
	case gLobby.Leave:
		msg = "opponent leave the game"
		cause = transport.StopCause_STOP_CAUSE_LEAVE
	}

	st := status.New(codes.Internal, msg)
	detSt, err := st.WithDetails(&transport.Interruption{Cause: cause})
	if err != nil {
		m := "can not attach details"
		return nil, status.Error(codes.Internal, m)
	}
	return detSt, nil
}

func interrupt(gameLogger zerolog.Logger, interruptCause gLobby.InterruptionCause) error {
	st, err := interruptionToStatus(interruptCause)
	if err != nil {
		gameLogger.Error().Err(err).Msg("can not process interruption")
	}
	interruptLogger := gameLogger.With().Str("cause", st.Message()).Logger()
	(&interruptLogger).Info().Msg("interrupted")
	return st.Err()
}
