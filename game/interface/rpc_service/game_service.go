package rpc_service

import (
	"github.com/google/flatbuffers/go"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"

	"github.com/izgib/tttserver/game"
	"github.com/izgib/tttserver/game/interface/rpc_service/i9e"
	"github.com/izgib/tttserver/game/models"
)

//
type GameService struct {
	lobbiesController game.GameLobbyUsecase
	logger            *zerolog.Logger
}

var EnumPlayerMark = map[i9e.MarkType]models.PlayerMark{
	i9e.MarkTypeCross:  models.CrossMark,
	i9e.MarkTypeNought: models.NoughtMark,
}

// GetListOfGames returns list of Games
func (s *GameService) GetListOfGames(params *i9e.GameFilter, stream i9e.GameConfigurator_GetListOfGamesServer) error {
	b := flatbuffers.NewBuilder(0)

	filterRange := &i9e.Range{}
	params.Rows(filterRange)
	rowsStart := filterRange.Start()
	rowsEnd := filterRange.End()

	filterRange = &i9e.Range{}
	params.Cols(filterRange)
	colsStart := filterRange.Start()
	colsEnd := filterRange.End()

	filterRange = &i9e.Range{}
	params.Win(filterRange)
	winStart := filterRange.Start()
	winEnd := filterRange.End()

	s.logger.Debug().Dict("filter", zerolog.Dict().
		Dict("row", zerolog.Dict().
			Int16("start", rowsStart).
			Int16("end", rowsEnd)).
		Dict("col", zerolog.Dict().
			Int16("start", colsStart).
			Int16("end", colsEnd)).
		Dict("win", zerolog.Dict().
			Int16("start", winStart).
			Int16("end", winEnd)).
		Str("mark", i9e.EnumNamesMarkTypeFilter[params.Mark()]),
	).Msg("Request List of Games")

	filter := game.GameFilter{
		Rows: game.Filter{rowsStart, rowsEnd},
		Cols: game.Filter{colsStart, colsEnd},
		Win:  game.Filter{winStart, winEnd},
		Mark: models.MoveChoice(params.Mark()),
	}

	games, err := s.lobbiesController.ListLobbies(filter)
	if err != nil {
		return err
	}

	i := 0
	for _, v := range games {
		s.logger.Debug().Int16("game", v.ID).Msgf("%d:", i)
		i9e.GameParamsStart(b)
		i9e.GameParamsAddRows(b, v.Settings.Rows)
		i9e.GameParamsAddCols(b, v.Settings.Cols)
		i9e.GameParamsAddWin(b, v.Settings.Win)
		i9e.GameParamsAddMark(b, i9e.MarkType(v.Mark))
		gcParams := i9e.GameParamsEnd(b)
		i9e.ListItemStart(b)
		i9e.ListItemAddID(b, v.ID)
		i9e.ListItemAddParams(b, gcParams)
		b.Finish(i9e.ListItemEnd(b))
		i++
		if err = stream.Send(b); err != nil {
			return err
		}
	}

	return nil
}

func NewGameService(gameUsecase game.GameLobbyUsecase, logger *zerolog.Logger) *GameService {
	s := &GameService{
		lobbiesController: gameUsecase,
		logger:            logger,
	}
	return s
}

func (s *GameService) CreateGame(stream i9e.GameConfigurator_CreateGameServer) error {
	in, err := stream.Recv()
	if err != nil {
		return err
	}

	var lobby game.GameLobby

	b := flatbuffers.NewBuilder(0)
	req := &flatbuffers.Table{}
	params := &i9e.GameParams{}
	if in.Req(req) {
		if in.ReqType() == i9e.CreatorReqMsgGameParams {
			params.Init(req.Bytes, req.Pos)

			s.logger.Debug().Int16("rows", params.Rows()).Int16("cols", params.Cols()).
				Int16("win", params.Win()).Str("mark", i9e.EnumNamesMarkType[params.Mark()]).
				Msg("Request creating of lobbiesController")

			gameConfig := game.GameConfiguration{
				Settings: models.GameSettings{
					Rows: params.Rows(),
					Cols: params.Cols(),
					Win:  params.Win(),
				},
				Mark: EnumPlayerMark[params.Mark()],
			}

			lobby, err = s.lobbiesController.CreateLobby(gameConfig)
			if err != nil {
				s.logger.Error().Err(err).Msg("got error")
				statusError := status.Error(codes.OutOfRange, err.Error())
				return statusError
			}

			go lobby.OnStart()
			lobby.CreatorReadyChan() <- true

			i9e.GameIdStart(b)
			i9e.GameIdAddID(b, lobby.GetID())
			CrWrapResponse(b, i9e.CreatorRespMsgGameId, i9e.GameIdEnd(b))

			if err := stream.Send(b); err != nil {
				lobby.CreatorReadyChan() <- false
				s.logger.Debug().Err(err)
				return err
			}
		}
	} else {
		err = status.Error(codes.InvalidArgument, "Expected to get GameParams")
		s.logger.Err(err)
		return err
	}

	gameLogger := models.CreateGameDebugLogger(lobby.GetID()).With().Str("player", models.EnumNamesMoveChoice[models.MoveChoice(lobby.GetCreatorMark())]).Logger()
	gameLogger.Debug().Msg("initialized")

	select {
	case gs := <-lobby.GameStartedChan():
		if gs {
			i9e.GameEventStart(b)
			i9e.GameEventAddType(b, i9e.GameEventTypeGameStarted)
			CrWrapResponse(b, i9e.CreatorRespMsgGameEvent, i9e.GameEventEnd(b))

			if err := stream.Send(b); err != nil {
				return err
			}
			gameLogger.Debug().Msg("event started have sent")
			// TODO add error handling after game creation
		} else {
			return nil
		}
	case <-stream.Context().Done():
		lobby.CreatorReadyChan() <- false
		return stream.Context().Err()
	}
	gameLogger.Debug().Msg("started")

	Chans := lobby.GetGameController().GetCreatorComChannels()

L:
	for {
		switch <-Chans.TypeChan {
		case game.MoveCh:
			if <-Chans.MoveRequesterChan {
				in, err := stream.Recv()
				if err != nil {
					gameLogger.Err(err).Msg("can not receive move")
					Chans.ErrChan <- err
					return err
				}
				req := &flatbuffers.Table{}
				if in.Req(req) {
					switch in.ReqType() {
					case i9e.CreatorReqMsgMove:
						move := &i9e.Move{}
						move.Init(req.Bytes, req.Pos)
						Chans.CommChan <- models.Move{I: move.Row(), J: move.Col()}
					default:
						err = status.Error(codes.InvalidArgument, "invalid type, expected move type")
						Chans.ErrChan <- err
						gameLogger.Err(err).Msg("wrong type")
						return err
					}
				}
			} else {
				playerMove := <-Chans.CommChan
				CrWrapResponse(b, i9e.CreatorRespMsgMove, CreateMove(b, playerMove.I, playerMove.J))
				err = stream.Send(b)
				Chans.ErrChan <- err
				if err != nil {
					gameLogger.Debug().Dict("move", zerolog.Dict().
						Int16("i", playerMove.I).
						Int16("j", playerMove.J),
					).Err(err).Msgf("lwas attempt to send move, but got error")
					return err
				}
			}
		case game.StateCh:
			state := <-Chans.StateChan

			switch state.StateType {
			case models.Won:
				startLine := CreateMove(b, state.WinLine.Start.I, state.WinLine.Start.J)
				endLine := CreateMove(b, state.WinLine.End.I, state.WinLine.End.J)

				i9e.WinLineStart(b)
				i9e.WinLineAddMark(b, i9e.MarkType(state.WinLine.Mark))
				i9e.WinLineAddStart(b, startLine)
				i9e.WinLineAddEnd(b, endLine)
				winLine := i9e.WinLineEnd(b)

				createEvent(b, i9e.GameEventTypeWin)
				i9e.GameEventAddFollowUp(b, winLine)
			case models.Tie:
				createEvent(b, i9e.GameEventTypeTie)
			case models.Running:
				createEvent(b, i9e.GameEventTypeOK)
			}
			CrWrapResponse(b, i9e.CreatorRespMsgGameEvent, i9e.GameEventEnd(b))

			err = stream.Send(b)
			Chans.ErrChan <- err
			if err != nil {
				gameLogger.Err(err).Msg("can not send state")
				return err
			}

			if state.StateType != models.Running {
				break L
			}
		case game.InterruptCh:
			cause := <-Chans.InterruptChan
			switch cause {
			case game.OppInvalidMove:
				createEvent(b, i9e.GameEventTypeOppCheating)
			case game.InvalidMove:
				createEvent(b, i9e.GameEventTypeCheating)
			default:
				createEvent(b, i9e.GameEventTypeOppDisconnected)
			}
			CrWrapResponse(b, i9e.CreatorRespMsgGameEvent, i9e.GameEventEnd(b))
			return stream.Send(b)
		}
	}

	return nil
}

func (s *GameService) JoinGame(stream i9e.GameConfigurator_JoinGameServer) error {
	var Chans game.PlayerComm

	in, err := stream.Recv()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return err
	}

	var lobby game.GameLobby

	b := flatbuffers.NewBuilder(0)
	req := &flatbuffers.Table{}
	gameID := &i9e.GameId{}

	if in.Req(req) {
		if in.ReqType() == i9e.OpponentReqMsgGameId {
			gameID.Init(req.Bytes, req.Pos)
			s.logger.Debug().Int16("game", gameID.ID()).Msg("opponent request to join")

			lobby, err = s.lobbiesController.JoinLobby(gameID.ID())
			if err != nil {
				statusErr := status.Error(codes.InvalidArgument, err.Error())
				s.logger.Debug().Err(err).Msg("not found")
				return statusErr
			}
		} else {
			err = status.Error(codes.InvalidArgument, "Expected to get GameID")
			s.logger.Debug().Err(err).Msg("wrong type")
			return err
		}
	}

	gameLogger := models.CreateGameDebugLogger(lobby.GetID()).With().Str("player", models.EnumNamesMoveChoice[models.MoveChoice(lobby.GetOpponentMark())]).Logger()
	lobby.OpponentReadyChan() <- true

	if <-lobby.GameStartedChan() {
		createEvent(b, i9e.GameEventTypeGameStarted)
		OppWrapResponse(b, i9e.OpponentRespMsgGameEvent, i9e.GameEventEnd(b))

		if err = stream.Send(b); err != nil {
			s.logger.Err(err).Int16("game", lobby.GetID()).Msg("can not send game start event")
			return err
		}
		gameLogger.Debug().Msg("event started have sent")
	} else {
		return nil
	}

	gameLogger.Debug().Msg("started")

	Chans = lobby.GetGameController().GetOpponentComChannels()
L:
	for {
		gameLogger.Debug().Msg("listen")
		switch <-Chans.TypeChan {
		case game.MoveCh:
			if <-Chans.MoveRequesterChan {
				in, err = stream.Recv()
				if err != nil {
					gameLogger.Err(err)
					Chans.ErrChan <- err
					return err
				}
				req = &flatbuffers.Table{}
				if in.Req(req) {
					switch in.ReqType() {
					case i9e.OpponentReqMsgMove:
						move := &i9e.Move{}
						move.Init(req.Bytes, req.Pos)
						Chans.CommChan <- models.Move{I: move.Row(), J: move.Col()}
					default:
						err = status.Error(codes.InvalidArgument, "invalid type, expected move type")
						gameLogger.Err(err).Msg("wrong type")
						Chans.ErrChan <- err
						return err
					}
				}
			} else {
				playerMove := <-Chans.CommChan
				OppWrapResponse(b, i9e.OpponentRespMsgMove, CreateMove(b, playerMove.I, playerMove.J))

				err = stream.Send(b)
				Chans.ErrChan <- err
				if err != nil {
					gameLogger.Debug().Dict("move", zerolog.Dict().
						Int16("i", playerMove.I).
						Int16("j", playerMove.J),
					).Err(err).Msg("bwas attempt to send move, but got error")
					return err
				}
			}
		case game.StateCh:
			state := <-Chans.StateChan
			switch state.StateType {
			case models.Won:
				startLine := CreateMove(b, state.WinLine.Start.I, state.WinLine.Start.J)
				endLine := CreateMove(b, state.WinLine.End.I, state.WinLine.End.J)

				i9e.WinLineStart(b)
				i9e.WinLineAddMark(b, i9e.MarkType(state.WinLine.Mark))
				i9e.WinLineAddStart(b, startLine)
				i9e.WinLineAddEnd(b, endLine)
				winLine := i9e.WinLineEnd(b)

				createEvent(b, i9e.GameEventTypeWin)
				i9e.GameEventAddFollowUp(b, winLine)
			case models.Tie:
				createEvent(b, i9e.GameEventTypeTie)
			case models.Running:
				createEvent(b, i9e.GameEventTypeOK)
			}
			OppWrapResponse(b, i9e.OpponentRespMsgGameEvent, i9e.GameEventEnd(b))

			err = stream.Send(b)
			Chans.ErrChan <- err
			if err != nil {
				gameLogger.Err(err).Msg("can not send state")
				return err
			}

			if state.StateType != models.Running {
				break L
			}
		case game.InterruptCh:
			cause := <-Chans.InterruptChan
			switch cause {
			case game.OppInvalidMove:
				createEvent(b, i9e.GameEventTypeOppCheating)
			case game.InvalidMove:
				createEvent(b, i9e.GameEventTypeCheating)
			default:
				createEvent(b, i9e.GameEventTypeOppDisconnected)
			}
			OppWrapResponse(b, i9e.OpponentRespMsgGameEvent, i9e.GameEventEnd(b))
			return stream.Send(b)
		}
	}

	return nil
}

func createEvent(b *flatbuffers.Builder, eventType i9e.GameEventType) {
	i9e.GameEventStart(b)
	i9e.GameEventAddType(b, eventType)
}

func CreateMove(b *flatbuffers.Builder, i int16, j int16) flatbuffers.UOffsetT {
	i9e.MoveStart(b)
	i9e.MoveAddRow(b, i)
	i9e.MoveAddCol(b, j)
	return i9e.MoveEnd(b)
}

func CrWrapResponse(b *flatbuffers.Builder, respType i9e.CreatorRespMsg, resp flatbuffers.UOffsetT) {
	i9e.CrResponseStart(b)
	i9e.CrResponseAddRespType(b, respType)
	i9e.CrResponseAddResp(b, resp)
	b.Finish(i9e.CrResponseEnd(b))
}

func OppWrapResponse(b *flatbuffers.Builder, respType i9e.OpponentRespMsg, resp flatbuffers.UOffsetT) {
	i9e.OppResponseStart(b)
	i9e.OppResponseAddRespType(b, respType)
	i9e.OppResponseAddResp(b, resp)
	b.Finish(i9e.OppResponseEnd(b))
}
