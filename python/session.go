// session.go - mixnet session client
// Copyright (C) 2017  Yawning Angel, Ruben Pollan, David Stainton
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package client

import (
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/katzenpost/client"
	"github.com/katzenpost/core/crypto/ecdh"
	"github.com/katzenpost/minclient/block"
	"github.com/op/go-logging"
)

// StorageStub implements the Storage interface
// as defined in the client library.
// XXX This should be replaced by something useful.
type StorageStub struct {
}

// GetBlocks returns a slice of blocks
func (s StorageStub) GetBlocks(*[block.MessageIDLength]byte) ([][]byte, error) {
	return nil, errors.New("failure: StorageStub GetBlocks not yet implemented")
}

// PutBlock puts a block into storage
func (s StorageStub) PutBlock(*[block.MessageIDLength]byte, []byte) error {
	return errors.New("failure: StorageStub PutBlock not yet implemented")
}

// Session holds the client session
type Session struct {
	client          *client.Client
	log             *logging.Logger
	clientCfg       *client.Config
	sessionCfg      *client.SessionConfig
	session         *client.Session
	ingressMsgQueue chan string
}

// NewSession stablishes a session with provider using key
func (c Client) NewSession(user string, provider string, key Key) (Session, error) {
	var err error
	var session Session
	clientCfg := &client.Config{
		User:       user,
		Provider:   provider,
		LinkKey:    key.priv,
		LogBackend: c.log,
		PKIClient:  c.pki,
	}
	gClient, err := client.New(clientCfg)
	if err != nil {
		return session, err
	}
	session.client = gClient
	session.ingressMsgQueue = make(chan string, 100)
	session.log = c.log.GetLogger(fmt.Sprintf("session_%s@%s", user, provider))
	return session, err
}

// ReceivedMessage is used to receive a message.
// This is a method on the MessageConsumer interface
// which is defined in the client library.
// XXX fix me
func (s Session) ReceivedMessage(senderPubKey *ecdh.PublicKey, message []byte) {
	s.log.Debug("ReceivedMessage")
	s.ingressMsgQueue <- string(message)
}

// GetMessage blocks until there is a message in the inbox
func (s Session) GetMessage() string {
	s.log.Debug("GetMessage")
	return <-s.ingressMsgQueue
}

// ReceivedACK is used to receive a signal that a message was received by
// the recipient Provider. This is a method on the MessageConsumer interface
// which is defined in the client library.
// XXX fix me
func (s Session) ReceivedACK(messageID *[block.MessageIDLength]byte, message []byte) {
	s.log.Debug("ReceivedACK")
}

// Get returns the identity public key for a given identity.
// This is part of the UserKeyDiscovery interface defined
// in the client library.
// XXX fix me
func (s Session) Get(identity string) (*ecdh.PublicKey, error) {
	s.log.Debugf("Get identity %s", identity)
	return nil, nil
}

// Connect connects the client to the Provider
func (s Session) Connect(identityKey Key) error {
	sessionCfg := client.SessionConfig{
		User:             s.clientCfg.User,
		Provider:         s.clientCfg.Provider,
		IdentityPrivKey:  identityKey.priv,
		LinkPrivKey:      s.clientCfg.LinkKey,
		MessageConsumer:  s,
		Storage:          new(StorageStub),
		UserKeyDiscovery: s,
	}
	s.sessionCfg = &sessionCfg
	var err error
	s.session, err = s.client.NewSession(&sessionCfg)
	return err
}

// Shutdown the session
func (s Session) Shutdown() {
	s.Shutdown()
}

// Send into the mix network
func (s Session) Send(recipient, provider, msg string) error {
	raw, err := hex.DecodeString(msg)
	if err != nil {
		return err
	}
	messageID, err := s.session.Send(recipient, provider, raw)
	if err != nil {
		return err
	}
	s.log.Debugf("sent message with messageID %x", messageID)
	return nil
}

// SendUnreliable into the mix network
func (s Session) SendUnreliable(recipient, provider, msg string) error {
	raw, err := hex.DecodeString(msg)
	if err != nil {
		return err
	}
	return s.session.SendUnreliable(recipient, provider, raw)
}
