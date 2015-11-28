package guble

import (
	assert "github.com/stretchr/testify/assert"
	"testing"
)

var aSendCommand = `send /foo
{"meta": "data"}
Hello World`

var aSubscribeCommand = "subscribe /foo/bar"

func TestParsingASendCommand(t *testing.T) {
	assert := assert.New(t)

	cmd, err := ParseCmd([]byte(aSendCommand))
	assert.NoError(err)

	assert.Equal(CMD_SEND, cmd.Name)
	assert.Equal("/foo", cmd.Arg)
	assert.Equal(`{"meta": "data"}`, cmd.HeaderJson)
	assert.Equal("Hello World", string(cmd.Body))
}

func TestSerializeASendCommand(t *testing.T) {
	cmd := &Cmd{
		Name:       CMD_SEND,
		Arg:        "/foo",
		HeaderJson: `{"meta": "data"}`,
		Body:       []byte("Hello World"),
	}

	assert.Equal(t, aSendCommand, string(cmd.Bytes()))
}

func TestParsingASubscribeCommand(t *testing.T) {
	assert := assert.New(t)

	cmd, err := ParseCmd([]byte(aSubscribeCommand))
	assert.NoError(err)

	assert.Equal(CMD_SUBSCRIBE, cmd.Name)
	assert.Equal("/foo/bar", cmd.Arg)
	assert.Equal("", cmd.HeaderJson)
	assert.Nil(cmd.Body)
}

func TestSerializeASubscribeCommand(t *testing.T) {
	cmd := &Cmd{
		Name: CMD_SUBSCRIBE,
		Arg:  "/foo/bar",
	}

	assert.Equal(t, aSubscribeCommand, string(cmd.Bytes()))
}
