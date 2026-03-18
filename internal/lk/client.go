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
