package lk

import (
	"context"
	"fmt"
	"strings"

	"github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go/v2"
)

type Room struct {
	Name            string
	SID             string
	NumParticipants uint32
	CreationTime    int64
	Metadata        string
}

type Participant struct {
	Identity string
	Name     string
	State    string
	Tracks   int
	JoinedAt int64
	Metadata string
}

type Egress struct {
	ID        string
	Status    string
	Type      string
	StartedAt int64
	Error     string
}

type Client struct {
	rooms    *lksdk.RoomServiceClient
	egresses *lksdk.EgressClient
}

func NewClient(url, apiKey, apiSecret string) *Client {
	return &Client{
		rooms:    lksdk.NewRoomServiceClient(url, apiKey, apiSecret),
		egresses: lksdk.NewEgressClient(url, apiKey, apiSecret),
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
			Metadata:        r.GetMetadata(),
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
			Metadata: p.GetMetadata(),
		}
	}

	return pp, nil
}

func (c *Client) ListEgresses(ctx context.Context, room string) ([]Egress, error) {
	res, err := c.egresses.ListEgress(ctx, &livekit.ListEgressRequest{RoomName: room})
	if err != nil {
		return nil, fmt.Errorf("list egresses: %w", err)
	}

	eg := make([]Egress, len(res.GetItems()))

	for i, e := range res.GetItems() {
		eg[i] = Egress{
			ID:        e.GetEgressId(),
			Status:    egressStatus(e.GetStatus()),
			Type:      egressType(e),
			StartedAt: e.GetStartedAt(),
			Error:     e.GetError(),
		}
	}

	return eg, nil
}

func egressStatus(s livekit.EgressStatus) string {
	name := s.String()

	return strings.TrimPrefix(name, "EGRESS_")
}

func egressType(e *livekit.EgressInfo) string {
	switch e.GetRequest().(type) {
	case *livekit.EgressInfo_RoomComposite:
		return "ROOM_COMPOSITE"
	case *livekit.EgressInfo_Web:
		return "WEB"
	case *livekit.EgressInfo_Participant:
		return "PARTICIPANT"
	case *livekit.EgressInfo_TrackComposite:
		return "TRACK_COMPOSITE"
	case *livekit.EgressInfo_Track:
		return "TRACK"
	default:
		return "UNKNOWN"
	}
}
