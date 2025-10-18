package ws

import "time"

// MessageType represents the type of WebSocket message
type MessageType string

const (
	// Client → Server
	MsgJoinRoom         MessageType = "JOIN_ROOM"
	MsgStartGame        MessageType = "START_GAME"
	MsgSubmitDefinition MessageType = "SUBMIT_DEFINITION"
	MsgSubmitVote       MessageType = "SUBMIT_VOTE"
	MsgEndGame          MessageType = "END_GAME"

	// Server → Client
	MsgRoomState     MessageType = "ROOM_STATE"
	MsgGameState     MessageType = "GAME_STATE"
	MsgVotingOptions MessageType = "VOTING_OPTIONS"
	MsgRoundResults  MessageType = "ROUND_RESULTS"
	MsgError         MessageType = "ERROR"
)

// Message represents a WebSocket message
type Message struct {
	Type      MessageType `json:"type"`
	Payload   interface{} `json:"payload"`
	Timestamp time.Time   `json:"timestamp"`
}

// JoinRoomPayload is the payload for JOIN_ROOM messages
type JoinRoomPayload struct {
	RoomCode string `json:"roomCode"`
	Nickname string `json:"nickname"`
}

// SubmitDefinitionPayload is the payload for SUBMIT_DEFINITION messages
type SubmitDefinitionPayload struct {
	Text string `json:"text"`
}

// SubmitVotePayload is the payload for SUBMIT_VOTE messages
type SubmitVotePayload struct {
	DefinitionID string `json:"definitionId"`
}

// ErrorPayload is the payload for ERROR messages
type ErrorPayload struct {
	Message string `json:"message"`
}
