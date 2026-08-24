package helps

import (
	"encoding/base64"
	"testing"

	"google.golang.org/protobuf/encoding/protowire"
)

func testStrictValidClaudeThinkingSignature() string {
	channelBlock := []byte{}
	channelBlock = protowire.AppendTag(channelBlock, 1, protowire.VarintType)
	channelBlock = protowire.AppendVarint(channelBlock, 12)
	channelBlock = protowire.AppendTag(channelBlock, 2, protowire.VarintType)
	channelBlock = protowire.AppendVarint(channelBlock, 2)
	channelBlock = protowire.AppendTag(channelBlock, 6, protowire.BytesType)
	channelBlock = protowire.AppendString(channelBlock, "claude-sonnet-4-6")

	container := []byte{}
	container = protowire.AppendTag(container, 1, protowire.BytesType)
	container = protowire.AppendBytes(container, channelBlock)

	payload := []byte{}
	payload = protowire.AppendTag(payload, 2, protowire.BytesType)
	payload = protowire.AppendBytes(payload, container)
	payload = protowire.AppendTag(payload, 3, protowire.VarintType)
	payload = protowire.AppendVarint(payload, 1)
	return base64.StdEncoding.EncodeToString(payload)
}

// observedFable5CAIS is the Claude CAIS envelope fixture from
// internal/signature (observed claude-fable-5 thinking signature).
const observedFable5CAIS = "CAISqwIKiAEIEBgCKkBHRlRBsNiptQUWfPoOhuQKwi5LnncZVO9bB5jqOs76D7uBtgktML0zqJtNmLHXHHcgD6lk4MQu4QBXzFd1lbC3Mg5jbGF1ZGUtZmFibGUtNTgBQgh0aGlua2luZ1okZDk3NDM5NzUtNGJiMC00OTM2LTllMjgtZDViMGQyMWJkYzQ4EgxCGh+XVFFFeySAjtAaDL/A1LltGu6MMJ+eXSIwsN0oBpDrqLv22UBfkMnTotnIbkvkOyb9xZHgigG6OZVHaI3gThm+maLKmgO5PrFLKlDFYp+YZksy/wKwszJlnLTPzAK+NUlfzagOE1ymtZTXhAYK260XyFYmg/te/C231+Fr/hoX+EJoUBnrn0gD7hqMISOT+TaFEuOXYsN517GfaxgB"

func TestClaudeThinkingReplayContentIsReplayable_AcceptsCAISSignedTurn(t *testing.T) {
	content := []byte(`[{"type":"thinking","thinking":"provider reasoning","signature":"` + observedFable5CAIS + `"},{"type":"text","text":"answer"}]`)
	if !ClaudeThinkingReplayContentIsReplayable(content) {
		t.Fatal("CAIS-signed thinking turn must be replayable so the cache can store it")
	}
}

func TestClaudeThinkingReplayContentIsReplayable_RejectsEgIFragment(t *testing.T) {
	content := []byte(`[{"type":"thinking","thinking":"provider reasoning","signature":"EgI="},{"type":"text","text":"answer"}]`)
	if ClaudeThinkingReplayContentIsReplayable(content) {
		t.Fatal("truncated EgI= fragment must not be written to the replay cache")
	}
}

func TestClaudeThinkingReplayContentIsReplayable_AcceptsStrictValidEnvelope(t *testing.T) {
	sig := testStrictValidClaudeThinkingSignature()
	content := []byte(`[{"type":"thinking","thinking":"provider reasoning","signature":"` + sig + `"},{"type":"text","text":"answer"}]`)
	if !ClaudeThinkingReplayContentIsReplayable(content) {
		t.Fatal("Strict-valid Claude envelope must be written to the replay cache")
	}
}
