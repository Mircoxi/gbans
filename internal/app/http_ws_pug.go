package app

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/leighmacdonald/gbans/pkg/fp"
	"github.com/leighmacdonald/gbans/pkg/mm"
	"github.com/leighmacdonald/steamid/v3/steamid"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
)

type classMapping map[string]steamid.SID64

type pugLobby struct {
	*sync.RWMutex
	logger    *zap.Logger
	Leader    *wsClient                `json:"leader"`
	LobbyID   string                   `json:"lobby_id"`
	Clients   wsClients                `json:"clients"`
	Messages  []pugUserMessageResponse `json:"messages"`
	Options   createLobbyOpts          `json:"options"`
	Classes   classMapping             `json:"classes"`
	ClassKeys []string                 `json:"class_keys"`
}

func (lobby *pugLobby) lobbyType() LobbyType {
	return lobbyTypePug
}

func newPugLobby(logger *zap.Logger, creator *wsClient, id string, opts createLobbyOpts) *pugLobby {
	lobby := &pugLobby{
		logger:   logger.Named(fmt.Sprintf("lobby-%s", id)),
		Leader:   creator,
		RWMutex:  &sync.RWMutex{},
		LobbyID:  id,
		Clients:  wsClients{creator},
		Messages: []pugUserMessageResponse{},
		Options:  opts,
	}

	switch opts.GameType {
	case mm.Sixes:
		lobby.ClassKeys = mm.ClassMappingKeysSixes
	case mm.Highlander:
		lobby.ClassKeys = mm.ClassMappingKeysHL
	case mm.Ultiduo:
		lobby.ClassKeys = mm.ClassMappingKeysUltiduo
	}

	creator.lobbies = append(creator.lobbies, lobby)

	return lobby
}

func (lobby *pugLobby) clientCount() int {
	lobby.RLock()
	defer lobby.RUnlock()

	return len(lobby.Clients)
}

func (lobby *pugLobby) id() string {
	lobby.RLock()
	defer lobby.RUnlock()

	return lobby.LobbyID
}

func (lobby *pugLobby) joinSlot(client *wsClient, slot string) error {
	lobby.Lock()
	defer lobby.Unlock()

	if !fp.Contains(lobby.ClassKeys, slot) {
		return ErrSlotInvalid
	}

	_, found := lobby.Classes[slot]
	if found {
		return ErrSlotInvalid
	}

	lobby.Classes[slot] = client.User.SteamID

	return nil
}

func (lobby *pugLobby) join(client *wsClient) error {
	lobby.Lock()
	defer lobby.Unlock()

	if slices.Contains(lobby.Clients, client) {
		return ErrDuplicateClient
	}

	lobby.Clients = append(lobby.Clients, client)
	client.lobbies = append(client.lobbies, lobby)

	lobby.logger.Info("User joined lobby", zap.String("lobby", lobby.LobbyID),
		zap.Int("clients", len(lobby.Clients)), zap.Bool("leader", len(lobby.Clients) == 1))

	if len(lobby.Clients) == 1 {
		return lobby.promote(client)
	}

	client.send(
		wsMsgTypePugJoinLobbyResponse,
		true,
		wsJoinLobbyResponse{Lobby: lobby},
	)

	return nil
}

func (lobby *pugLobby) promote(client *wsClient) error {
	lobby.Lock()
	defer lobby.Unlock()
	lobby.Leader = client

	return nil
}

func (lobby *pugLobby) leave(client *wsClient) error {
	lobby.RLock()
	if !slices.Contains(lobby.Clients, client) {
		lobby.RUnlock()

		return ErrUnknownClient
	}
	lobby.RUnlock()
	lobby.broadcast(wsMsgTypePugLeaveLobbyResponse, true, struct {
		LobbyID string `json:"lobby_id"`
		SteamID string `json:"steam_id"`
	}{
		LobbyID: lobby.id(),
		SteamID: client.User.SteamID.String(),
	},
	)

	lobby.Clients = fp.Remove(lobby.Clients, client)

	client.removeLobby(lobby)

	return nil
}

func (lobby *pugLobby) broadcast(msgType wsMsgType, status bool, payload any) {
	lobby.Clients.broadcast(msgType, status, payload)
}

func (lobby *pugLobby) sendUserMessage(client *wsClient, msg lobbyUserMessageRequest) {
	lobby.Lock()
	defer lobby.Unlock()

	userMessage := pugUserMessageResponse{
		User:      client.User,
		Message:   msg.Message,
		CreatedAt: time.Now(),
	}

	lobby.Messages = append(lobby.Messages, userMessage)
	lobby.broadcast(wsMsgTypePugUserMessageResponse, true, userMessage)
}

func leavePugLobby(connectionManager *wsConnectionManager, client *wsClient, _ json.RawMessage) error {
	lobby, found := client.currentPugLobby()
	if !found {
		return ErrInvalidLobbyID
	}

	if errLeave := lobby.leave(client); errLeave != nil {
		return errLeave
	}

	if lobby.clientCount() == 0 {
		if errRemove := connectionManager.removeLobby(lobby.LobbyID); errRemove != nil {
			connectionManager.logger.Error("Failed to remove empty lobby", zap.Error(errRemove))

			return nil
		}
	}

	return nil
}

func joinPugLobby(connectionManager *wsConnectionManager, client *wsClient, payload json.RawMessage) error {
	var req wsJoinLobbyRequest
	if errUnmarshal := json.Unmarshal(payload, &req); errUnmarshal != nil {
		connectionManager.logger.Error("Failed to unmarshal create request", zap.Error(errUnmarshal))

		return errors.Wrapf(errUnmarshal, "Failed to decode payload")
	}

	lobby, findErr := connectionManager.findLobby(req.LobbyID)
	if findErr != nil {
		return findErr
	}

	if errJoin := lobby.join(client); errJoin != nil {
		return errors.Wrap(errJoin, "Failed to join lobby")
	}

	return nil
}

func joinPugLobbySlot(connectionManager *wsConnectionManager, client *wsClient, payload json.RawMessage) error {
	var req wsJoinLobbySlotRequest
	if errUnmarshal := json.Unmarshal(payload, &req); errUnmarshal != nil {
		connectionManager.logger.Error("Failed to unmarshal create request", zap.Error(errUnmarshal))

		return errors.Wrapf(errUnmarshal, "Failed to decode payload")
	}

	lobby, findErr := connectionManager.findLobby(req.LobbyID)
	if findErr != nil {
		return findErr
	}

	if errJoin := lobby.joinSlot(client, req.Slot); errJoin != nil {
		return errors.Wrap(errJoin, "Failed to join lobby slot")
	}

	return nil
}

func createPugLobby(connectionManager *wsConnectionManager, client *wsClient, payload json.RawMessage) error {
	var req createLobbyOpts
	if errUnmarshal := json.Unmarshal(payload, &req); errUnmarshal != nil {
		connectionManager.logger.Error("Failed to unmarshal create request", zap.Error(errUnmarshal))

		return errors.Wrapf(errUnmarshal, "Failed to decode payload")
	}

	lobby, errCreate := connectionManager.createPugLobby(client, req)
	if errCreate != nil {
		return errCreate
	}

	sendPugCreateLobbyResponse(client, lobby)

	return nil
}

func sendPugUserMessage(cm *wsConnectionManager, client *wsClient, payload json.RawMessage) error {
	var req lobbyUserMessageRequest
	if errUnmarshal := json.Unmarshal(payload, &req); errUnmarshal != nil {
		cm.logger.Error("Failed to unmarshal user message request", zap.Error(errUnmarshal))

		return errors.New("Invalid request")
	}

	lobby, found := client.currentPugLobby()
	if !found {
		return ErrInvalidLobbyID
	}

	lobby.sendUserMessage(client, req)

	return nil
}

func sendPugLobbyListStates(cm *wsConnectionManager, client *wsClient, _ json.RawMessage) {
	client.send(wsMsgTypePugLobbyListStatesResponse, true, wsPugLobbyListStatesResponse{Lobbies: cm.pubLobbyList()})
}

func sendPugCreateLobbyResponse(client *wsClient, lobby *pugLobby) {
	client.send(
		wsMsgTypePugCreateLobbyResponse,
		true,
		wsJoinLobbyResponse{Lobby: lobby},
	)
}
