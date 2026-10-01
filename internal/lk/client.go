package lk

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/livekit/protocol/auth"
	"github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go/v2"
)

type Room struct {
	Name             string
	SID              string
	NumParticipants  uint32
	NumPublishers    uint32
	CreationTime     int64 // Unix seconds
	Metadata         string
	ActiveRecording  bool
	MaxParticipants  uint32
	EmptyTimeout     uint32 // seconds
	DepartureTimeout uint32 // seconds
	EnabledCodecs    []Codec
}

type Codec struct {
	MimeType string
	FmtpLine string
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
	Kind        string
	State       string
	JoinedAt    int64
	Metadata    string
	Attributes  map[string]string
	Permission  Permission
	Mic         TrackState
	Camera      TrackState
	Screen      TrackState
	ScreenAudio TrackState
	Tracks      []Track
}

type VideoLayer struct {
	Quality      string
	Width        uint32
	Height       uint32
	Bitrate      uint32 // bps
	SSRC         uint32
	RepairSSRC   uint32
	SpatialLayer int32
	RID          string
}

type TrackCodec struct {
	MimeType  string
	MID       string
	CID       string
	SDPCID    string
	LayerMode string
	Layers    []VideoLayer
}

type Track struct {
	SID               string
	Name              string
	Type              string
	Source            string
	MimeType          string
	MID               string
	Stream            string
	Muted             bool
	Width             uint32
	Height            uint32
	Simulcast         bool
	DisableDTX        bool
	Stereo            bool
	DisableRED        bool
	Encryption        string
	BackupCodecPolicy string
	AudioFeatures     []string
	Version           int64 // Unix microseconds
	Layers            []VideoLayer
	Codecs            []TrackCodec
}

type Egress struct {
	ID        string
	Status    string
	Type      string
	StartedAt int64 // Unix nanoseconds
	Error     string
}

type Client struct {
	rooms         *lksdk.RoomServiceClient
	egresses      *lksdk.EgressClient
	sip           *lksdk.SIPClient
	agentDispatch *lksdk.AgentDispatchClient
	logger        *slog.Logger
}

// NewClient creates a LiveKit API client. logger receives debug-level
// diagnostics (e.g. the exact room identifiers sent and returned by the
// API) useful for tracking down server-side inconsistencies; pass nil to
// discard them.
func NewClient(url, apiKey, apiSecret string, logger *slog.Logger) *Client {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}

	return &Client{
		rooms:         lksdk.NewRoomServiceClient(url, apiKey, apiSecret),
		egresses:      lksdk.NewEgressClient(url, apiKey, apiSecret),
		sip:           lksdk.NewSIPClient(url, apiKey, apiSecret),
		agentDispatch: lksdk.NewAgentDispatchServiceClient(url, apiKey, apiSecret),
		logger:        logger,
	}
}

func (c *Client) ListRooms(ctx context.Context) ([]Room, error) {
	res, err := c.rooms.ListRooms(ctx, &livekit.ListRoomsRequest{})
	if err != nil {
		c.logger.Error("list rooms", "error", err)

		return nil, fmt.Errorf("list rooms: %w", err)
	}

	rooms := make([]Room, len(res.GetRooms()))
	for i, r := range res.GetRooms() {
		rooms[i] = Room{
			Name:             r.GetName(),
			SID:              r.GetSid(),
			NumParticipants:  r.GetNumParticipants(),
			NumPublishers:    r.GetNumPublishers(),
			CreationTime:     r.GetCreationTime(),
			Metadata:         r.GetMetadata(),
			ActiveRecording:  r.GetActiveRecording(),
			MaxParticipants:  r.GetMaxParticipants(),
			EmptyTimeout:     r.GetEmptyTimeout(),
			DepartureTimeout: r.GetDepartureTimeout(),
			EnabledCodecs:    roomCodecs(r.GetEnabledCodecs()),
		}

		c.logger.Debug("list rooms: got room", "name", rooms[i].Name, "sid", rooms[i].SID)
	}

	return rooms, nil
}

func (c *Client) CreateRoom(ctx context.Context, name string, emptyTimeout, maxParticipants uint32) (Room, error) {
	res, err := c.rooms.CreateRoom(ctx, &livekit.CreateRoomRequest{
		Name:            name,
		EmptyTimeout:    emptyTimeout,
		MaxParticipants: maxParticipants,
	})
	if err != nil {
		c.logger.Error("create room", "name", name, "error", err)

		return Room{}, fmt.Errorf("create room: %w", err)
	}

	return Room{
		Name:             res.GetName(),
		SID:              res.GetSid(),
		NumParticipants:  res.GetNumParticipants(),
		NumPublishers:    res.GetNumPublishers(),
		CreationTime:     res.GetCreationTime(),
		Metadata:         res.GetMetadata(),
		ActiveRecording:  res.GetActiveRecording(),
		MaxParticipants:  res.GetMaxParticipants(),
		EmptyTimeout:     res.GetEmptyTimeout(),
		DepartureTimeout: res.GetDepartureTimeout(),
		EnabledCodecs:    roomCodecs(res.GetEnabledCodecs()),
	}, nil
}

func (c *Client) DeleteRoom(ctx context.Context, room string) error {
	c.logger.Debug("delete room: request", "name", room)

	if _, err := c.rooms.DeleteRoom(ctx, &livekit.DeleteRoomRequest{Room: room}); err != nil {
		c.logger.Error("delete room", "name", room, "error", err)

		return fmt.Errorf("delete room: %w", err)
	}

	c.logger.Debug("delete room: ok", "name", room)

	return nil
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
			Kind:        p.GetKind().String(),
			State:       p.GetState().String(),
			JoinedAt:    p.GetJoinedAt(),
			Metadata:    p.GetMetadata(),
			Attributes:  p.GetAttributes(),
			Permission:  participantPermission(p.GetPermission()),
			Mic:         trackState(tracks, livekit.TrackSource_MICROPHONE),
			Camera:      trackState(tracks, livekit.TrackSource_CAMERA),
			Screen:      trackState(tracks, livekit.TrackSource_SCREEN_SHARE),
			ScreenAudio: trackState(tracks, livekit.TrackSource_SCREEN_SHARE_AUDIO),
			Tracks:      trackInfos(tracks),
		}
	}

	return pp, nil
}

func (c *Client) RemoveParticipant(ctx context.Context, room, identity string) error {
	if _, err := c.rooms.RemoveParticipant(ctx, &livekit.RoomParticipantIdentity{Room: room, Identity: identity}); err != nil {
		c.logger.Error("remove participant", "room", room, "identity", identity, "error", err)

		return fmt.Errorf("remove participant: %w", err)
	}

	return nil
}

func (c *Client) SetTrackMuted(ctx context.Context, room, identity, trackSID string, muted bool) error {
	if _, err := c.rooms.MutePublishedTrack(ctx, &livekit.MuteRoomTrackRequest{
		Room:     room,
		Identity: identity,
		TrackSid: trackSID,
		Muted:    muted,
	}); err != nil {
		c.logger.Error("mute track", "room", room, "identity", identity, "track", trackSID, "error", err)

		return fmt.Errorf("mute track: %w", err)
	}

	return nil
}

func (c *Client) UpdatePermission(ctx context.Context, room, identity string, perm Permission) error {
	//nolint:staticcheck
	req := &livekit.UpdateParticipantRequest{
		Room:     room,
		Identity: identity,
		Permission: &livekit.ParticipantPermission{
			CanSubscribe:          perm.CanSubscribe,
			CanPublish:            perm.CanPublish,
			CanPublishData:        perm.CanPublishData,
			CanPublishSources:     parseTrackSources(perm.CanPublishSources),
			Hidden:                perm.Hidden,
			Recorder:              perm.Recorder,
			CanUpdateMetadata:     perm.CanUpdateMetadata,
			CanSubscribeMetrics:   perm.CanSubscribeMetrics,
			CanManageAgentSession: perm.CanManageAgentSession,
		},
	}

	if _, err := c.rooms.UpdateParticipant(ctx, req); err != nil {
		c.logger.Error("update permission", "room", room, "identity", identity, "error", err)

		return fmt.Errorf("update permission: %w", err)
	}

	return nil
}

// parseTrackSources converts the display names produced by
// participantPermission back into the enum values the API expects,
// preserving the participant's existing publish-source restriction across
// an otherwise unrelated permission edit.
func parseTrackSources(names []string) []livekit.TrackSource {
	if len(names) == 0 {
		return nil
	}

	out := make([]livekit.TrackSource, 0, len(names))

	for _, name := range names {
		if v, ok := livekit.TrackSource_value[name]; ok {
			out = append(out, livekit.TrackSource(v))
		}
	}

	return out
}

// TokenGrant is the subset of auth.VideoGrant exposed by the "generate
// token" UI action. RoomJoin and RoomAdmin are scoped to Room (the token
// mints room-join/publish permissions, or admin rights, for that one
// room); RoomCreate, RoomList, RoomRecord and IngressAdmin are project-wide
// and ignore Room entirely.
type TokenGrant struct {
	RoomJoin             bool
	CanPublish           bool
	CanSubscribe         bool
	CanPublishData       bool
	CanUpdateOwnMetadata bool
	Hidden               bool
	Recorder             bool
	RoomAdmin            bool
	RoomCreate           bool
	RoomList             bool
	RoomRecord           bool
	IngressAdmin         bool
}

// CreateToken mints a signed access token for identity, valid for ttl, with
// the given grant scoped to room (room is only meaningful when
// grant.RoomJoin or grant.RoomAdmin is set). This is a local signing
// operation using the context's configured API key/secret (no LiveKit API
// call), so it's available regardless of the context's write setting.
func (c *Client) CreateToken(identity, room string, ttl time.Duration, grant TokenGrant) (string, error) {
	at := c.rooms.CreateToken()
	at.SetIdentity(identity).
		SetValidFor(ttl).
		SetVideoGrant(&auth.VideoGrant{
			Room:                 room,
			RoomJoin:             grant.RoomJoin,
			RoomAdmin:            grant.RoomAdmin,
			RoomCreate:           grant.RoomCreate,
			RoomList:             grant.RoomList,
			RoomRecord:           grant.RoomRecord,
			IngressAdmin:         grant.IngressAdmin,
			CanPublish:           &grant.CanPublish,
			CanSubscribe:         &grant.CanSubscribe,
			CanPublishData:       &grant.CanPublishData,
			CanUpdateOwnMetadata: &grant.CanUpdateOwnMetadata,
			Hidden:               grant.Hidden,
			Recorder:             grant.Recorder,
		})

	token, err := at.ToJWT()
	if err != nil {
		return "", fmt.Errorf("create token: %w", err)
	}

	return token, nil
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

func roomCodecs(cc []*livekit.Codec) []Codec {
	out := make([]Codec, len(cc))
	for i, c := range cc {
		out[i] = Codec{MimeType: c.GetMime(), FmtpLine: c.GetFmtpLine()}
	}

	return out
}

func trackInfos(tracks []*livekit.TrackInfo) []Track {
	out := make([]Track, len(tracks))

	for i, t := range tracks {
		features := make([]string, len(t.GetAudioFeatures()))
		for j, f := range t.GetAudioFeatures() {
			features[j] = f.String()
		}

		codecs := make([]TrackCodec, len(t.GetCodecs()))
		for j, c := range t.GetCodecs() {
			codecs[j] = TrackCodec{
				MimeType:  c.GetMimeType(),
				MID:       c.GetMid(),
				CID:       c.GetCid(),
				SDPCID:    c.GetSdpCid(),
				LayerMode: c.GetVideoLayerMode().String(),
				Layers:    videoLayers(c.GetLayers()),
			}
		}

		//nolint:staticcheck
		out[i] = Track{
			SID:               t.GetSid(),
			Name:              t.GetName(),
			Type:              t.GetType().String(),
			Source:            t.GetSource().String(),
			MimeType:          t.GetMimeType(),
			MID:               t.GetMid(),
			Stream:            t.GetStream(),
			Muted:             t.GetMuted(),
			Width:             t.GetWidth(),
			Height:            t.GetHeight(),
			Simulcast:         t.GetSimulcast(),
			DisableDTX:        t.GetDisableDtx(),
			Stereo:            t.GetStereo(),
			DisableRED:        t.GetDisableRed(),
			Encryption:        t.GetEncryption().String(),
			BackupCodecPolicy: t.GetBackupCodecPolicy().String(),
			AudioFeatures:     features,
			Version:           t.GetVersion().GetUnixMicro(),
			Layers:            videoLayers(t.GetLayers()),
			Codecs:            codecs,
		}
	}

	return out
}

func videoLayers(layers []*livekit.VideoLayer) []VideoLayer {
	out := make([]VideoLayer, len(layers))
	for i, l := range layers {
		out[i] = VideoLayer{
			Quality:      l.GetQuality().String(),
			Width:        l.GetWidth(),
			Height:       l.GetHeight(),
			Bitrate:      l.GetBitrate(),
			SSRC:         l.GetSsrc(),
			RepairSSRC:   l.GetRepairSsrc(),
			SpatialLayer: l.GetSpatialLayer(),
			RID:          l.GetRid(),
		}
	}

	return out
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

type AgentDispatch struct {
	ID        string
	AgentName string
	Room      string
	Metadata  string
	CreatedAt int64 // Unix seconds
	DeletedAt int64 // Unix seconds
	JobStatus string
	JobError  string
}

func (c *Client) ListAgentDispatches(ctx context.Context, room string) ([]AgentDispatch, error) {
	res, err := c.agentDispatch.ListDispatch(ctx, &livekit.ListAgentDispatchRequest{Room: room})
	if err != nil {
		return nil, fmt.Errorf("list agent dispatches: %w", err)
	}

	dd := make([]AgentDispatch, len(res.GetAgentDispatches()))
	for i, d := range res.GetAgentDispatches() {
		dd[i] = AgentDispatch{
			ID:        d.GetId(),
			AgentName: d.GetAgentName(),
			Room:      d.GetRoom(),
			Metadata:  d.GetMetadata(),
			CreatedAt: d.GetState().GetCreatedAt(),
			DeletedAt: d.GetState().GetDeletedAt(),
		}

		if jobs := d.GetState().GetJobs(); len(jobs) > 0 {
			dd[i].JobStatus = jobs[0].GetState().GetStatus().String()
			dd[i].JobError = jobs[0].GetState().GetError()
		}
	}

	return dd, nil
}

// SIPKind identifies which kind of SIP configuration entry a SIPEntry
// represents (inbound trunk, outbound trunk, or dispatch rule); the three
// share few fields, so they're normalized into one listing.
type SIPKind string

const (
	SIPInboundTrunk  SIPKind = "INBOUND"
	SIPOutboundTrunk SIPKind = "OUTBOUND"
	SIPDispatchRule  SIPKind = "DISPATCH"
)

type SIPEntry struct {
	Kind     SIPKind
	ID       string
	Name     string
	Numbers  []string
	Address  string // outbound trunks only
	Metadata string
}

func (c *Client) ListSIP(ctx context.Context) ([]SIPEntry, error) {
	inbound, err := c.sip.ListSIPInboundTrunk(ctx, &livekit.ListSIPInboundTrunkRequest{})
	if err != nil {
		return nil, fmt.Errorf("list sip inbound trunks: %w", err)
	}

	outbound, err := c.sip.ListSIPOutboundTrunk(ctx, &livekit.ListSIPOutboundTrunkRequest{})
	if err != nil {
		return nil, fmt.Errorf("list sip outbound trunks: %w", err)
	}

	rules, err := c.sip.ListSIPDispatchRule(ctx, &livekit.ListSIPDispatchRuleRequest{})
	if err != nil {
		return nil, fmt.Errorf("list sip dispatch rules: %w", err)
	}

	entries := make([]SIPEntry, 0, len(inbound.GetItems())+len(outbound.GetItems())+len(rules.GetItems()))

	for _, t := range inbound.GetItems() {
		entries = append(entries, SIPEntry{
			Kind:     SIPInboundTrunk,
			ID:       t.GetSipTrunkId(),
			Name:     t.GetName(),
			Numbers:  t.GetNumbers(),
			Metadata: t.GetMetadata(),
		})
	}

	for _, t := range outbound.GetItems() {
		entries = append(entries, SIPEntry{
			Kind:     SIPOutboundTrunk,
			ID:       t.GetSipTrunkId(),
			Name:     t.GetName(),
			Numbers:  t.GetNumbers(),
			Address:  t.GetAddress(),
			Metadata: t.GetMetadata(),
		})
	}

	for _, r := range rules.GetItems() {
		entries = append(entries, SIPEntry{
			Kind:     SIPDispatchRule,
			ID:       r.GetSipDispatchRuleId(),
			Name:     r.GetName(),
			Numbers:  r.GetNumbers(),
			Metadata: r.GetMetadata(),
		})
	}

	return entries, nil
}
