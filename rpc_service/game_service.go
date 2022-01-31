package rpc_service

import (
	"github.com/google/flatbuffers/go"
	"github.com/izgib/tttserver/base"
	"github.com/izgib/tttserver/game"
	"github.com/izgib/tttserver/rpc_service/transport"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
)

type GameService struct {
	lobbiesController base.GameLobbyController
	logger            *zerolog.Logger
}

var EnumPlayerMark = map[transport.MarkType]game.PlayerMark{
	transport.MarkTypeCross:  game.CrossMark,
	transport.MarkTypeNought: game.NoughtMark,
}

// GetListOfGames returns list of Games
func (s *GameService) GetListOfGames(params *transport.GameFilter, stream transport.GameConfigurator_GetListOfGamesServer) error {
	b := flatbuffers.NewBuilder(0)

	filterRange := &transport.Range{}
	params.Rows(filterRange)
	rowsStart := filterRange.Start()
	rowsEnd := filterRange.End()

	filterRange = &transport.Range{}
	params.Cols(filterRange)
	colsStart := filterRange.Start()
	colsEnd := filterRange.End()

	filterRange = &transport.Range{}
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
		Str("mark", transport.EnumNamesMarkTypeFilter[params.Mark()]),
	).Msg("Request List of Games")

	filter := base.GameFilter{
		Rows: base.Filter{rowsStart, rowsEnd},
		Cols: base.Filter{colsStart, colsEnd},
		Win:  base.Filter{winStart, winEnd},
		Mark: game.MoveChoice(params.Mark()),
	}

	games, err := s.lobbiesController.ListLobbies(filter)
	if err != nil {
		return err
	}

	i := 0
	for _, v := range games {
		s.logger.Debug().Int16("game", v.ID).Msgf("%d:", i)
		transport.GameParamsStart(b)
		transport.GameParamsAddRows(b, v.Settings.Rows)
		transport.GameParamsAddCols(b, v.Settings.Cols)
		transport.GameParamsAddWin(b, v.Settings.Win)
		transport.GameParamsAddMark(b, transport.MarkType(v.Mark))
		gcParams := transport.GameParamsEnd(b)
		transport.ListItemStart(b)
		transport.ListItemAddID(b, v.ID)
		transport.ListItemAddParams(b, gcParams)
		b.Finish(transport.ListItemEnd(b))
		i++
		if err = stream.Send(b); err != nil {
			return err
		}
	}

	return nil
}

func NewGameService(gameController base.GameLobbyController, logger *zerolog.Logger) *GameService {
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

	var lobby base.GameLobby

	b := flatbuffers.NewBuilder(0)
	req := &flatbuffers.Table{}
	params := &transport.GameParams{}
	if in.Req(req) {
		if in.ReqType() == transport.CreatorReqMsgGameParams {
			params.Init(req.Bytes, req.Pos)

			s.logger.Debug().Int16("rows", params.Rows()).Int16("cols", params.Cols()).
				Int16("win", params.Win()).Str("mark", transport.EnumNamesMarkType[params.Mark()]).
				Msg("Request of creating game lobby")

			gameConfig := base.GameConfiguration{
				Settings: game.GameSettings{
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

			lobby.CreatorReadyChan() <- true

			transport.GameIdStart(b)
			transport.GameIdAddID(b, lobby.GetID())
			CrWrapResponse(b, transport.CreatorRespMsgGameId, transport.GameIdEnd(b))

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

	gameLogger := game.CreateGameDebugLogger(lobby.GetID()).With().
		Str("player", game.EnumNamesMoveChoice[game.MoveChoice(lobby.GetCreatorMark())]).Logger()

	gameLogger.Debug().Msg("initialized")

	var Chans *base.CommChannels
	select {
	case gs := <-lobby.GameStartedChan():
		if gs {
			Chans = lobby.GetGameController().GetCreatorChannels()
			createEvent(b, transport.GameEventTypeGameStarted)
			CrWrapResponse(b, transport.CreatorRespMsgGameEvent, transport.GameEventEnd(b))

			if err := stream.Send(b); err != nil {
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

	var interruptCause base.InterruptionCause
	for {
		select {
		case action := <-Chans.Actions:
			switch a := action.(type) {
			case base.Step:
				CrWrapResponse(b, transport.CreatorRespMsgMove, CreateMove(b, a.Pos.I, a.Pos.J))
				err = stream.Send(b)
				Chans.Reactions <- base.Status{Err: err}
				if err != nil {
					gameLogger.Debug().Dict("move", zerolog.Dict().
						Int16("i", a.Pos.I).
						Int16("j", a.Pos.J),
					).Err(err).Msg("can not send move")
					return err
				}
			case base.State:
				switch a.State.StateType {
				case game.Won:
					startLine := CreateMove(b, a.State.WinLine.Start.I, a.State.WinLine.Start.J)
					endLine := CreateMove(b, a.State.WinLine.End.I, a.State.WinLine.End.J)

					transport.WinLineStart(b)
					transport.WinLineAddMark(b, transport.MarkType(a.State.WinLine.Mark))
					transport.WinLineAddStart(b, startLine)
					transport.WinLineAddEnd(b, endLine)
					winLine := transport.WinLineEnd(b)

					createEvent(b, transport.GameEventTypeWin)
					transport.GameEventAddFollowUp(b, winLine)
				case game.Tie:
					createEvent(b, transport.GameEventTypeTie)
				case game.Running:
					createEvent(b, transport.GameEventTypeOK)
				}
				CrWrapResponse(b, transport.CreatorRespMsgGameEvent, transport.GameEventEnd(b))

				err = stream.Send(b)
				Chans.Reactions <- base.Status{Err: err}
				if err != nil {
					gameLogger.Err(err).Msg("can not send state")
					return err
				}

				if a.State.StateType != game.Running {
					gameLogger.Info().Msg("ended")
					return nil
				}
			case base.Interruption:
				interruptCause = a.Cause
				goto InterruptionEvent
			case base.ReceiveStep:
				respChan := make(chan base.ReceivedStep)
				go func() {
					in, err = stream.Recv()
					if err != nil {
						gameLogger.Error().Err(err).Msg("can not receive move")
						respChan <- base.ReceivedStep{Err: err}
						return
					}
					req = &flatbuffers.Table{}
					if in.Req(req) {
						switch in.ReqType() {
						case transport.CreatorReqMsgMove:
							move := &transport.Move{}
							move.Init(req.Bytes, req.Pos)
							respChan <- base.ReceivedStep{Pos: game.Move{I: move.Row(), J: move.Col()}, Err: nil}
						default:
							err = status.Error(codes.InvalidArgument, "invalid type, expected move type")
							gameLogger.Error().Err(err).Msg("wrong type")
							respChan <- base.ReceivedStep{Err: err}
						}
					}
				}()
				select {
				case action = <-Chans.Actions:
					interruption, ok := action.(base.Interruption)
					if !ok {
						gameLogger.Error().Msg("got something else instead of interruption")
						return status.Error(codes.Internal, "")
					}
					interruptCause = interruption.Cause
					goto InterruptionEvent
				case resp := <-respChan:
					Chans.Reactions <- resp
				}
			}
		case <-stream.Context().Done():
			ctxErr := status.FromContextError(stream.Context().Err()).Err()
			gameLogger.Error().Err(ctxErr).Msg("connection canceled")
			Chans.Err = ctxErr
			close(Chans.Done)
			return stream.Context().Err()
		}
	}
InterruptionEvent:
	var msg string
	var cause *InterruptionInfo
	switch interruptCause {
	case base.OppInvalidMove:
		msg = "invalid move from opponent"
		cause = &InterruptionInfo{Cause: InterruptionCause_OPP_INVALID_MOVE}
	case base.InvalidMove:
		msg = "invalid move"
		cause = &InterruptionInfo{Cause: InterruptionCause_INVALID_MOVE}
	case base.Disconnect:
		msg = "opponent disconnected"
		cause = &InterruptionInfo{Cause: InterruptionCause_DISCONNECT}
	case base.Leave:
		msg = "opponent leave the game"
		cause = &InterruptionInfo{Cause: InterruptionCause_LEAVE}
	}
	gameLogger.With().Str("cause", msg).Logger()
	gameLogger.Info().Msg("interrupted")
	code := status.New(codes.Unknown, msg)
	st, err := code.WithDetails(cause)
	if err != nil {
		gameLogger.Error().Err(err).Msg("can not attach details")
		return status.Error(codes.Internal, "")
	}
	return st.Err()
}

func (s *GameService) JoinGame(stream transport.GameConfigurator_JoinGameServer) error {
	var Chans *base.CommChannels

	in, err := stream.Recv()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return err
	}

	var lobby base.GameLobby

	b := flatbuffers.NewBuilder(0)
	req := &flatbuffers.Table{}
	gameID := &transport.GameId{}

	if in.Req(req) {
		if in.ReqType() == transport.OpponentReqMsgGameId {
			gameID.Init(req.Bytes, req.Pos)
			s.logger.Debug().Int16("game", gameID.ID()).Msg("opponent request to join")

			lobby, err = s.lobbiesController.JoinLobby(gameID.ID())
			if err != nil {
				statusErr := status.Error(codes.InvalidArgument, err.Error())
				s.logger.Error().Err(err).Msg("not found")
				return statusErr
			}
		} else {
			err = status.Error(codes.InvalidArgument, "Expected to get GameID")
			s.logger.Debug().Err(err).Msg("wrong type")
			return err
		}
	}

	gameLogger := game.CreateGameDebugLogger(lobby.GetID()).With().Str("player", game.EnumNamesMoveChoice[game.MoveChoice(lobby.GetOpponentMark())]).Logger()
	lobby.OpponentReadyChan() <- true

	if <-lobby.GameStartedChan() {
		createEvent(b, transport.GameEventTypeGameStarted)
		OppWrapResponse(b, transport.OpponentRespMsgGameEvent, transport.GameEventEnd(b))

		if err = stream.Send(b); err != nil {
			Chans.Err = err
			close(Chans.Done)
			s.logger.Err(err).Int16("game", lobby.GetID()).Msg("can not send game start event")
			return err
		}
		gameLogger.Debug().Msg("event started have sent")
	} else {
		return nil
	}

	gameLogger.Debug().Msg("started")

	Chans = lobby.GetGameController().GetOpponentChannels()

	var interruptCause base.InterruptionCause
	for {
		select {
		case action := <-Chans.Actions:
			switch a := action.(type) {
			case base.Step:
				OppWrapResponse(b, transport.OpponentRespMsgMove, CreateMove(b, a.Pos.I, a.Pos.J))
				err = stream.Send(b)
				Chans.Reactions <- base.Status{Err: err}
				if err != nil {
					gameLogger.Debug().Dict("move", zerolog.Dict().
						Int16("i", a.Pos.I).
						Int16("j", a.Pos.J),
					).Err(err).Msg("can not send move")
					return err
				}
			case base.State:
				switch a.State.StateType {
				case game.Won:
					startLine := CreateMove(b, a.State.WinLine.Start.I, a.State.WinLine.Start.J)
					endLine := CreateMove(b, a.State.WinLine.End.I, a.State.WinLine.End.J)

					transport.WinLineStart(b)
					transport.WinLineAddMark(b, transport.MarkType(a.State.WinLine.Mark))
					transport.WinLineAddStart(b, startLine)
					transport.WinLineAddEnd(b, endLine)
					winLine := transport.WinLineEnd(b)

					createEvent(b, transport.GameEventTypeWin)
					transport.GameEventAddFollowUp(b, winLine)
				case game.Tie:
					createEvent(b, transport.GameEventTypeTie)
				case game.Running:
					createEvent(b, transport.GameEventTypeOK)
				}
				OppWrapResponse(b, transport.OpponentRespMsgGameEvent, transport.GameEventEnd(b))

				err = stream.Send(b)
				Chans.Reactions <- base.Status{Err: err}
				if err != nil {
					gameLogger.Err(err).Msg("can not send state")
					return err
				}

				if a.State.StateType != game.Running {
					gameLogger.Info().Msg("ended")
					return nil
				}
			case base.Interruption:
				interruptCause = a.Cause
				goto InterruptionEvent
			case base.ReceiveStep:
				respChan := make(chan base.ReceivedStep)
				go func() {
					in, err = stream.Recv()
					if err != nil {
						gameLogger.Error().Err(err).Msg("can not receive move")
						respChan <- base.ReceivedStep{Err: err}
						return
					}
					req = &flatbuffers.Table{}
					if in.Req(req) {
						switch in.ReqType() {
						case transport.OpponentReqMsgMove:
							move := &transport.Move{}
							move.Init(req.Bytes, req.Pos)
							respChan <- base.ReceivedStep{Pos: game.Move{I: move.Row(), J: move.Col()}, Err: nil}
						default:
							err = status.Error(codes.InvalidArgument, "invalid type, expected move type")
							gameLogger.Error().Err(err).Msg("wrong type")
							respChan <- base.ReceivedStep{Err: err}
						}
					}
				}()
				select {
				case action = <-Chans.Actions:
					interruption, ok := action.(base.Interruption)
					if !ok {
						gameLogger.Error().Msg("got something else instead of interruption")
						return status.Error(codes.Internal, "")
					}
					interruptCause = interruption.Cause
					goto InterruptionEvent
				case resp := <-respChan:
					Chans.Reactions <- resp
				}
			}
		case <-stream.Context().Done():
			ctxErr := status.FromContextError(stream.Context().Err()).Err()
			gameLogger.Error().Err(ctxErr).Msg("connection canceled")
			Chans.Err = ctxErr
			close(Chans.Done)
			return stream.Context().Err()
		}
	}
InterruptionEvent:
	var msg string
	var cause *InterruptionInfo
	switch interruptCause {
	case base.OppInvalidMove:
		msg = "invalid move from opponent"
		cause = &InterruptionInfo{Cause: InterruptionCause_OPP_INVALID_MOVE}
	case base.InvalidMove:
		msg = "invalid move"
		cause = &InterruptionInfo{Cause: InterruptionCause_INVALID_MOVE}
	case base.Disconnect:
		msg = "opponent disconnected"
		cause = &InterruptionInfo{Cause: InterruptionCause_DISCONNECT}
	case base.Leave:
		msg = "opponent leave the game"
		cause = &InterruptionInfo{Cause: InterruptionCause_LEAVE}
	}

	interruptLogger := gameLogger.With().Str("cause", msg).Logger()
	(&interruptLogger).Info().Msg("interrupted")

	code := status.New(codes.Unknown, msg)
	st, err := code.WithDetails(cause)
	if err != nil {
		gameLogger.Error().Err(err).Msg("can not attach details")
		return status.Error(codes.Internal, "")
	}
	return st.Err()
}

func createEvent(b *flatbuffers.Builder, eventType transport.GameEventType) {
	transport.GameEventStart(b)
	transport.GameEventAddType(b, eventType)
}

func CreateMove(b *flatbuffers.Builder, i int16, j int16) flatbuffers.UOffsetT {
	transport.MoveStart(b)
	transport.MoveAddRow(b, i)
	transport.MoveAddCol(b, j)
	return transport.MoveEnd(b)
}

func CrWrapResponse(b *flatbuffers.Builder, respType transport.CreatorRespMsg, resp flatbuffers.UOffsetT) {
	transport.CrResponseStart(b)
	transport.CrResponseAddRespType(b, respType)
	transport.CrResponseAddResp(b, resp)
	b.Finish(transport.CrResponseEnd(b))
}

func OppWrapResponse(b *flatbuffers.Builder, respType transport.OpponentRespMsg, resp flatbuffers.UOffsetT) {
	transport.OppResponseStart(b)
	transport.OppResponseAddRespType(b, respType)
	transport.OppResponseAddResp(b, resp)
	b.Finish(transport.OppResponseEnd(b))
}
