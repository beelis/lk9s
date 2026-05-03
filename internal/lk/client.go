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
	CreationTime    int64 // Unix seconds
	Metadata        string
}

type TrackState uint8

const (
	TrackAbsent TrackState = iota
	TrackActive
	TrackMuted
)

func (t TrackState) String() string {
	switch t {
	case TrackActive:
		return "●"
	case TrackMuted:
		return "○"
	default:
		return "-"
	}
}

type Permission struct {
	CanPublish            bool
	CanSubscribe          bool
	CanPublishData        bool
	CanUpdateMetadata     bool
	Hidden                bool
	Recorder              bool
	CanPublishSources     []string
	CanManageAgentSession bool
	CanSubscribeMetrics   bool
}

type Participant struct {
	Identity    string
	Name        string
	Kind        livekit.ParticipantInfo_Kind
	State       string
	JoinedAt    int64
	Metadata    string
	Attributes  map[string]string
	Permission  Permission
	Mic         TrackState
	Camera      TrackState
	Screen      TrackState
	ScreenAudio TrackState
}

type Egress struct {
	ID        string
	Status    string
	Type      string
	StartedAt int64 // Unix nanoseconds
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
		tracks := p.GetTracks()
		pp[i] = Participant{
			Identity:    p.GetIdentity(),
			Name:        p.GetName(),
			Kind:        p.GetKind(),
			State:       p.GetState().String(),
			JoinedAt:    p.GetJoinedAt(),
			Metadata:    p.GetMetadata(),
			Attributes:  p.GetAttributes(),
			Permission:  participantPermission(p.GetPermission()),
			Mic:         trackState(tracks, livekit.TrackSource_MICROPHONE),
			Camera:      trackState(tracks, livekit.TrackSource_CAMERA),
			Screen:      trackState(tracks, livekit.TrackSource_SCREEN_SHARE),
			ScreenAudio: trackState(tracks, livekit.TrackSource_SCREEN_SHARE_AUDIO),
		}
	}

	return pp, nil
}

func participantPermission(perm *livekit.ParticipantPermission) Permission {
	if perm == nil {
		return Permission{}
	}

	sources := make([]string, len(perm.GetCanPublishSources()))
	for i, s := range perm.GetCanPublishSources() {
		sources[i] = s.String()
	}

	//nolint:staticcheck
	return Permission{
		CanPublish:            perm.GetCanPublish(),
		CanSubscribe:          perm.GetCanSubscribe(),
		CanPublishData:        perm.GetCanPublishData(),
		CanUpdateMetadata:     perm.GetCanUpdateMetadata(),
		CanManageAgentSession: perm.GetCanManageAgentSession(),
		CanSubscribeMetrics:   perm.GetCanSubscribeMetrics(),
		Hidden:                perm.GetHidden(),
		Recorder:              perm.GetRecorder(),
		CanPublishSources:     sources,
	}
}

func trackState(tracks []*livekit.TrackInfo, source livekit.TrackSource) TrackState {
	for _, t := range tracks {
		if t.GetSource() == source {
			if t.GetMuted() {
				return TrackMuted
			}

			return TrackActive
		}
	}

	return TrackAbsent
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
