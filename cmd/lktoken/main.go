package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/livekit/protocol/auth"
)

func main() {
	apiKey := flag.String("api-key", "", "LiveKit API key")
	apiSecret := flag.String("api-secret", "", "LiveKit API secret")
	room := flag.String("room", "", "Room name")
	identity := flag.String("identity", "", "Participant identity")
	flag.Parse()

	if *apiKey == "" || *apiSecret == "" || *room == "" || *identity == "" {
		fmt.Fprintln(os.Stderr, "usage: lktoken -api-key <key> -api-secret <secret> -room <room> -identity <identity>")
		os.Exit(1)
	}

	token, err := generateToken(*apiKey, *apiSecret, *room, *identity)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println(token)
}

func generateToken(apiKey, apiSecret, room, identity string) (string, error) {
	at := auth.NewAccessToken(apiKey, apiSecret)
	at.SetVideoGrant(&auth.VideoGrant{
		RoomCreate: true,
		RoomJoin:   true,
		Room:       room,
	}).SetIdentity(identity)

	token, err := at.ToJWT()
	if err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	return token, nil
}
