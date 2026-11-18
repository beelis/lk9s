//go:build e2e

package lk_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/livekit/protocol/auth"
	"github.com/livekit/protocol/livekit"
	"github.com/twitchtv/twirp"
)

type fakeServer struct {
	url string

	mu    sync.Mutex
	rooms map[string]*livekit.Room
	parts map[string][]*livekit.ParticipantInfo
}

func newFakeServer() (*fakeServer, func()) {
	f := &fakeServer{rooms: map[string]*livekit.Room{}, parts: map[string][]*livekit.ParticipantInfo{}}

	mux := http.NewServeMux()
	room := livekit.NewRoomServiceServer(roomSvc{f: f})
	egress := livekit.NewEgressServer(egressSvc{})
	dispatch := livekit.NewAgentDispatchServiceServer(dispatchSvc{})

	mux.Handle(room.PathPrefix(), room)
	mux.Handle(egress.PathPrefix(), egress)
	mux.Handle(dispatch.PathPrefix(), dispatch)

	srv := httptest.NewServer(requireAuth(mux))
	f.url = srv.URL

	return f, srv.Close
}

func (f *fakeServer) addParticipant(room, identity string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.parts[room] = append(f.parts[room], &livekit.ParticipantInfo{
		Identity:   identity,
		State:      livekit.ParticipantInfo_ACTIVE,
		Permission: &livekit.ParticipantPermission{CanPublish: true, CanSubscribe: true},
	})
}

type roomSvc struct {
	livekit.RoomService

	f *fakeServer
}

func (s roomSvc) CreateRoom(_ context.Context, r *livekit.CreateRoomRequest) (*livekit.Room, error) {
	s.f.mu.Lock()
	defer s.f.mu.Unlock()

	room := &livekit.Room{
		Name: r.GetName(), Sid: "RM_" + r.GetName(), EmptyTimeout: r.GetEmptyTimeout(),
		MaxParticipants: r.GetMaxParticipants(), CreationTime: time.Now().Unix(),
	}
	s.f.rooms[room.GetName()] = room

	return room, nil
}

func (s roomSvc) ListRooms(context.Context, *livekit.ListRoomsRequest) (*livekit.ListRoomsResponse, error) {
	s.f.mu.Lock()
	defer s.f.mu.Unlock()

	res := &livekit.ListRoomsResponse{}
	for _, r := range s.f.rooms {
		res.Rooms = append(res.Rooms, r)
	}

	return res, nil
}

func (s roomSvc) DeleteRoom(_ context.Context, r *livekit.DeleteRoomRequest) (*livekit.DeleteRoomResponse, error) {
	s.f.mu.Lock()
	defer s.f.mu.Unlock()

	delete(s.f.rooms, r.GetRoom())
	delete(s.f.parts, r.GetRoom())

	return &livekit.DeleteRoomResponse{}, nil
}

func (s roomSvc) UpdateRoomMetadata(_ context.Context, r *livekit.UpdateRoomMetadataRequest) (*livekit.Room, error) {
	s.f.mu.Lock()
	defer s.f.mu.Unlock()

	room, ok := s.f.rooms[r.GetRoom()]
	if !ok {
		return nil, twirp.NotFoundError("room not found")
	}

	room.Metadata = r.GetMetadata()

	return room, nil
}

func (s roomSvc) ListParticipants(
	_ context.Context, r *livekit.ListParticipantsRequest,
) (*livekit.ListParticipantsResponse, error) {
	s.f.mu.Lock()
	defer s.f.mu.Unlock()

	return &livekit.ListParticipantsResponse{Participants: slices.Clone(s.f.parts[r.GetRoom()])}, nil
}

func (s roomSvc) RemoveParticipant(
	_ context.Context, r *livekit.RoomParticipantIdentity,
) (*livekit.RemoveParticipantResponse, error) {
	s.f.mu.Lock()
	defer s.f.mu.Unlock()

	s.f.parts[r.GetRoom()] = slices.DeleteFunc(s.f.parts[r.GetRoom()], func(p *livekit.ParticipantInfo) bool {
		return p.GetIdentity() == r.GetIdentity()
	})

	return &livekit.RemoveParticipantResponse{}, nil
}

func (s roomSvc) UpdateParticipant(
	_ context.Context, r *livekit.UpdateParticipantRequest,
) (*livekit.ParticipantInfo, error) {
	s.f.mu.Lock()
	defer s.f.mu.Unlock()

	for _, p := range s.f.parts[r.GetRoom()] {
		if p.GetIdentity() == r.GetIdentity() {
			p.Permission = r.GetPermission()

			return p, nil
		}
	}

	return nil, twirp.NotFoundError("participant not found")
}

func (s roomSvc) SendData(context.Context, *livekit.SendDataRequest) (*livekit.SendDataResponse, error) {
	return &livekit.SendDataResponse{}, nil
}

type egressSvc struct{ livekit.Egress }

func (egressSvc) ListEgress(context.Context, *livekit.ListEgressRequest) (*livekit.ListEgressResponse, error) {
	return &livekit.ListEgressResponse{}, nil
}

type dispatchSvc struct{ livekit.AgentDispatchService }

func (dispatchSvc) ListDispatch(
	context.Context, *livekit.ListAgentDispatchRequest,
) (*livekit.ListAgentDispatchResponse, error) {
	return &livekit.ListAgentDispatchResponse{}, nil
}

// requireAuth rejects requests not signed with the dev secret, as a real
// server does.
func requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")

		v, err := auth.ParseAPIToken(raw)
		if err == nil {
			_, _, err = v.Verify(env("LK9S_E2E_API_SECRET", "secret"))
		}

		if err != nil {
			http.Error(w, `{"code":"unauthenticated","msg":"invalid token"}`, http.StatusUnauthorized)

			return
		}

		next.ServeHTTP(w, r)
	})
}
