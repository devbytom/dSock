package dsock_test

import (
	"encoding/json"
	dsock "github.com/Cretezy/dSock-go"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"testing"
)

type ConnectSuite struct {
	suite.Suite
}

func TestConnectSuite(t *testing.T) {
	suite.Run(t, new(ConnectSuite))
}

func (suite *ConnectSuite) TestClaimConnect() {
	claim, err := dSockClient.CreateClaim(dsock.CreateClaimOptions{
		User:    "connect",
		Session: "claim",
	})
	if !checkRequestError(suite.Suite, err, "claim creation") {
		return
	}

	conn, resp, err := websocket.DefaultDialer.Dial("ws://worker/connect?claim="+claim.Id, nil)
	if !checkConnectionError(suite.Suite, err, resp) {
		return
	}

	defer conn.Close()

	info, err := dSockClient.GetInfo(dsock.GetInfoOptions{
		Target: dsock.Target{
			User:    "connect",
			Session: "claim",
		},
	})
	if !checkRequestError(suite.Suite, err, "getting info") {
		return
	}

	connections := info.Connections
	if !suite.Len(connections, 1, "Incorrect number of connections") {
		return
	}

	connection := connections[0]

	suite.Equal("connect", claim.User, "Incorrect claim user")
	suite.Equal("connect", connection.User, "Incorrect connection user")

	suite.Equal("claim", claim.Session, "Incorrect claim user session")
	suite.Equal("claim", connection.Session, "Incorrect connection user session")
}

func (suite *ConnectSuite) TestInvalidClaim() {
	_, resp, err := websocket.DefaultDialer.Dial("ws://worker/connect?claim=invalid-claim", nil)
	if !suite.Error(err, "Did not error when expected during connection") {
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if !suite.NoError(err, "Could not read body") {
		return
	}

	var parsedBody map[string]interface{}

	err = json.Unmarshal(body, &parsedBody)
	if !suite.NoError(err, "Could not parse body") {
		return
	}

	if !suite.Equal(false, parsedBody["success"], "Succeeded when should have failed") {
		return
	}

	if !suite.Equal("MISSING_CLAIM", parsedBody["errorCode"], "Incorrect error code") {
		return
	}
}

func (suite *ConnectSuite) TestJwtConnect() {
	// Hard coded JWT with max expiry:
	// {
	//  "sub": "connect",
	//  "sid": "jwt",
	//  "exp": 2147485546
	//}
	jwt := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJjb25uZWN0Iiwic2lkIjoiand0IiwiZXhwIjoyMTQ3NDg1NTQ2fQ.oMbgPfg86I1sWs6IK25AP0H4ftzUVt9asKr9W9binW0"

	conn, resp, err := websocket.DefaultDialer.Dial("ws://worker/connect?jwt="+jwt, nil)
	if !checkConnectionError(suite.Suite, err, resp) {
		return
	}

	defer conn.Close()

	info, err := dSockClient.GetInfo(dsock.GetInfoOptions{
		Target: dsock.Target{
			User:    "connect",
			Session: "jwt",
		},
	})
	if !checkRequestError(suite.Suite, err, "getting info") {
		return
	}

	connections := info.Connections
	if !suite.Len(connections, 1, "Incorrect number of connections") {
		return
	}

	connection := connections[0]

	suite.Equal("connect", connection.User, "Incorrect connection user")
	suite.Equal("jwt", connection.Session, "Incorrect connection user session")
}

func (suite *ConnectSuite) TestInvalidJwt() {
	// Hard coded JWT with invalid expiry:
	// {
	//  "sub": "connect",
	//  "sid": "invalid",
	//  "exp": "invalid"
	//}
	jwt := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJjb25uZWN0Iiwic2lkIjoiaW52YWxpZCIsImV4cCI6ImludmFsaWQifQ.afZ4Mi-K0FeS35n7sivpNlq41JUi-QKVEjkH6mGWOrk"

	_, resp, err := websocket.DefaultDialer.Dial("ws://worker/connect?jwt="+jwt, nil)
	if !suite.Error(err, "Did not error when expecting during connection") {
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if !suite.NoError(err, "Could not read body") {
		return
	}

	var parsedBody map[string]interface{}

	err = json.Unmarshal(body, &parsedBody)
	if !suite.NoError(err, "Could not parse body") {
		return
	}

	if !suite.Equal(false, parsedBody["success"], "Application succeeded when expected to fail") {
		return
	}

	if !suite.Equal("INVALID_JWT", parsedBody["errorCode"], "Incorrect error code") {
		return
	}
}

func (suite *ConnectSuite) TestJwtConnectChannel() {
	// Hard coded JWT with max expiry:
	// {
	//  "sub": "connect",
	//  "sid": "jwt_channel",
	//  "exp": 2147485546,
	//  "channels": ["connect_jwt"]
	//}
	jwt := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJjb25uZWN0Iiwic2lkIjoiand0X2NoYW5uZWwiLCJleHAiOjIxNDc0ODU1NDYsImNoYW5uZWxzIjpbImNvbm5lY3Rfand0Il19.LdKHWk1W6DLMR02T0g1lGfhPdyyKqDJHqvUL3YQ9tLQ"

	conn, resp, err := websocket.DefaultDialer.Dial("ws://worker/connect?jwt="+jwt, nil)
	if !checkConnectionError(suite.Suite, err, resp) {
		return
	}

	defer conn.Close()

	info, err := dSockClient.GetInfo(dsock.GetInfoOptions{
		Target: dsock.Target{
			Channel: "connect_jwt",
		},
	})
	if !checkRequestError(suite.Suite, err, "getting info") {
		return
	}

	connections := info.Connections
	if !suite.Len(connections, 1, "Incorrect number of connections") {
		return
	}

	connection := connections[0]

	if !suite.Equal("connect", connection.User, "Incorrect connection user") {
		return
	}
	if !suite.Equal("jwt_channel", connection.Session, "Incorrect connection user session") {
		return
	}

	// Includes default_channels in info
	if !suite.Equal([]string{"connect_jwt", "global"}, interfaceToStringSlice(connection.Channels), "Incorrect connection channels") {
		return
	}
}
