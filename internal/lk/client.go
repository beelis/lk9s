package lk

import (
	"context"
	"fmt"

	"github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go/v2"
)

type Room struct {
	Name            string
	SID             string
	NumParticipants uint32
	CreationTime    int64
}

type Participant struct {
	Identity string
	Name     string
	State    string
	Tracks   int
	JoinedAt int64
}

type Client struct {
	rooms *lksdk.RoomServiceClient
}

func NewClient(url, apiKey, apiSecret string) *Client {
	return &Client{
		rooms: lksdk.NewRoomServiceClient(url, apiKey, apiSecret),
	}
}

func (c *Client) ListRooms(ctx context.Context) ([]Room, error) {
	res, err := c.rooms.ListRooms(ctx, &livekit.ListRoomsRequest{})
	if err != nil {
		return nil, fmt.Errorf("list rooms: %w", err)
	}

	rooms := make([]Room, len(res.GetRooms()))
	for i, r := range res.GetRooms() {
		rooms[i] = Room{
			Name:            r.GetName(),
			SID:             r.GetSid(),
			NumParticipants: r.GetNumParticipants(),
			CreationTime:    r.GetCreationTime(),
		}
	}

	return rooms, nil
}

func (c *Client) ListParticipants(ctx context.Context, room string) ([]Participant, error) {
	res, err := c.rooms.ListParticipants(ctx, &livekit.ListParticipantsRequest{Room: room})
	if err != nil {
		return nil, fmt.Errorf("list participants: %w", err)
	}

	pp := make([]Participant, len(res.GetParticipants()))
	for i, p := range res.GetParticipants() {
		pp[i] = Participant{
			Identity: p.GetIdentity(),
			Name:     p.GetName(),
			State:    p.GetState().String(),
			Tracks:   len(p.GetTracks()),
			JoinedAt: p.GetJoinedAt(),
		}
	}

	return pp, nil
}
