// This file verifies coding-agent helper behavior that does not require
// database state.

package catalog

import "testing"

// TestNewAgentIDReturnsOpaqueAgentPrefix verifies public Agent IDs use the expected prefix.
func TestNewAgentIDReturnsOpaqueAgentPrefix(t *testing.T) {
	id := newAgentID()
	if len(id) != len("agt-")+32 || id[:4] != "agt-" {
		t.Fatalf("unexpected agent id %q", id)
	}
}

// TestNormalizeAgentInputTrimsAndNormalizes verifies request normalization.
func TestNormalizeAgentInputTrimsAndNormalizes(t *testing.T) {
	input := normalizeAgentInput(AgentInput{
		Name:          " Agent ",
		ModelName:     " gpt-5 ",
		ModelProtocol: " OpenAI ",
		AgentType:     " Codex ",
		IconKey:       " code ",
		Notes:         " notes ",
	})
	if input.Name != "Agent" ||
		input.ModelName != "gpt-5" ||
		input.ModelProtocol != ProtocolOpenAI ||
		input.AgentType != AgentTypeCodex ||
		input.IconKey != "code" ||
		input.Notes != "notes" {
		t.Fatalf("unexpected normalized input: %#v", input)
	}
}
