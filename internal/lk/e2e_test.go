//go:build e2e

package lk_test

import (
	"context"
	"fmt"
	"os"
	"slices"
	"strings"
	"testing"
	"time"

	lksdk "github.com/livekit/server-sdk-go/v2"

	"github.com/mariusgassen/lk9s/internal/lk"
)

const (
	alice = "alice"
	bob   = "bob"
)

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return def
}

// fake is the in-process stand-in used when LK9S_E2E_URL is unset; it lets
// the suite run without a LiveKit server. It has no media plane, so
// participants are registered directly instead of joining over WebRTC.
var fake *fakeServer

func TestMain(m *testing.M) {
	stop := func() {}

	if os.Getenv("LK9S_E2E_URL") == "" {
		fake, stop = newFakeServer()
	}

	code := m.Run()

	stop()
	os.Exit(code)
}

func serverURL(t *testing.T) string {
	t.Helper()

	if u := os.Getenv("LK9S_E2E_URL"); u != "" {
		return u
	}

	return fake.url
}

func newClient(t *testing.T) *lk.Client {
	t.Helper()

	return lk.NewClient(
		serverURL(t),
		env("LK9S_E2E_API_KEY", "devkey"),
		env("LK9S_E2E_API_SECRET", "secret"),
		nil,
	)
}

func newRoom(t *testing.T, c *lk.Client) string {
	t.Helper()

	name := fmt.Sprintf("e2e-%s-%d", t.Name(), time.Now().UnixNano())

	if _, err := c.CreateRoom(t.Context(), name, 60, 0); err != nil {
		t.Fatalf("create room: %v", err)
	}

	t.Cleanup(func() {
		_ = c.DeleteRoom(context.WithoutCancel(t.Context()), name)
	})

	return name
}

func join(t *testing.T, c *lk.Client, room, identity string) *lksdk.Room {
	t.Helper()

	token, err := c.CreateToken(identity, room, time.Minute, lk.TokenGrant{
		RoomJoin: true, CanPublish: true, CanSubscribe: true, CanPublishData: true,
	})
	if err != nil {
		t.Fatalf("create token: %v", err)
	}

	if os.Getenv("LK9S_E2E_URL") == "" {
		fake.addParticipant(room, identity)

		return nil
	}

	r, err := lksdk.ConnectToRoomWithToken(
		env("LK9S_E2E_WS_URL", "ws://localhost:7880"), token, lksdk.NewRoomCallback(),
	)
	if err != nil {
		t.Fatalf("connect %s: %v", identity, err)
	}

	t.Cleanup(r.Disconnect)

	return r
}

func waitParticipants(t *testing.T, c *lk.Client, room string, want int) []lk.Participant {
	t.Helper()

	var got []lk.Participant

	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		var err error

		got, err = c.ListParticipants(t.Context(), room)
		if err != nil {
			t.Fatalf("list participants: %v", err)
		}

		if len(got) == want {
			return got
		}

		time.Sleep(200 * time.Millisecond)
	}

	t.Fatalf("want %d participants in %s, got %d", want, room, len(got))

	return nil
}

func TestRoomLifecycle(t *testing.T) {
	t.Parallel()

	c := newClient(t)
	name := newRoom(t, c)

	rooms, err := c.ListRooms(t.Context())
	if err != nil {
		t.Fatalf("list rooms: %v", err)
	}

	var found bool

	for _, r := range rooms {
		if r.Name == name {
			found = true

			if r.EmptyTimeout != 60 {
				t.Errorf("empty timeout = %d, want 60", r.EmptyTimeout)
			}
		}
	}

	if !found {
		t.Fatalf("room %q not listed", name)
	}

	if err := c.UpdateRoomMetadata(t.Context(), name, "hello"); err != nil {
		t.Fatalf("update metadata: %v", err)
	}

	rooms, _ = c.ListRooms(t.Context())
	for _, r := range rooms {
		if r.Name == name && r.Metadata != "hello" {
			t.Errorf("metadata = %q, want hello", r.Metadata)
		}
	}

	if err := c.DeleteRoom(t.Context(), name); err != nil {
		t.Fatalf("delete room: %v", err)
	}

	rooms, _ = c.ListRooms(t.Context())
	for _, r := range rooms {
		if r.Name == name {
			t.Errorf("room %q still listed after delete", name)
		}
	}
}

func TestParticipants(t *testing.T) {
	t.Parallel()

	c := newClient(t)
	room := newRoom(t, c)

	join(t, c, room, alice)
	join(t, c, room, bob)

	pp := waitParticipants(t, c, room, 2)
	ids := []string{pp[0].Identity, pp[1].Identity}
	slices.Sort(ids)

	if ids[0] != alice || ids[1] != bob {
		t.Fatalf("unexpected identities: %q", ids)
	}

	perm := lk.Permission{CanSubscribe: true, CanPublish: false, CanPublishData: true}
	if err := c.UpdatePermission(t.Context(), room, alice, perm); err != nil {
		t.Fatalf("update permission: %v", err)
	}

	pp, err := c.ListParticipants(t.Context(), room)
	if err != nil {
		t.Fatalf("list participants: %v", err)
	}

	for _, p := range pp {
		if p.Identity == alice && p.Permission.CanPublish {
			t.Error("alice can still publish after permission update")
		}
	}

	if err := c.SendData(t.Context(), room, "e2e", []byte("ping"), nil, true); err != nil {
		t.Fatalf("send data: %v", err)
	}

	if err := c.RemoveParticipant(t.Context(), room, bob); err != nil {
		t.Fatalf("remove participant: %v", err)
	}

	pp = waitParticipants(t, c, room, 1)
	if pp[0].Identity != alice {
		t.Errorf("remaining participant = %q, want alice", pp[0].Identity)
	}
}

func TestEmptyListings(t *testing.T) {
	t.Parallel()

	c := newClient(t)
	room := newRoom(t, c)

	if _, err := c.ListEgresses(t.Context(), room); err != nil && !needsRedis(err) {
		t.Errorf("list egresses: %v", err)
	}

	if _, err := c.ListAgentDispatches(t.Context(), room); err != nil && !needsRedis(err) {
		t.Errorf("list agent dispatches: %v", err)
	}
}

// needsRedis reports whether err is how a bare dev server (no Redis) rejects
// egress and agent dispatch calls: v1.9 answers "redis required" or panics.
func needsRedis(err error) bool {
	msg := err.Error()

	return strings.Contains(msg, "redis required") || strings.Contains(msg, "Internal service panic")
}

func TestInvalidCredentials(t *testing.T) {
	t.Parallel()

	c := lk.NewClient(serverURL(t), "devkey", "wrong-secret-wrong-secret-wrong", nil)

	if _, err := c.ListRooms(t.Context()); err == nil {
		t.Fatal("expected an error with wrong credentials")
	}
}
