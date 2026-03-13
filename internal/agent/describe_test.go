package agent

import (
	"strings"
	"testing"
	"time"
)

func TestBuildDescribePrompt(t *testing.T) {
	messages := []Message{
		{Role: RoleUser, Content: "Fix the auth bug"},
		{Role: RoleAssistant, Content: "I'll look into the authentication issue."},
		{Role: RoleTool, Content: "Read: src/auth.go"},
	}

	prompt := buildDescribePrompt(messages)

	// Should contain the system prompt
	if !strings.Contains(prompt, "Summarize what this workspace is doing") {
		t.Error("prompt should contain system instructions")
	}

	// Should include user and assistant messages
	if !strings.Contains(prompt, "User: Fix the auth bug") {
		t.Error("prompt should contain user message")
	}
	if !strings.Contains(prompt, "Assistant: I'll look into the authentication issue.") {
		t.Error("prompt should contain assistant message")
	}

	// Should NOT include tool messages
	if strings.Contains(prompt, "Read: src/auth.go") {
		t.Error("prompt should not contain tool messages")
	}
}

func TestBuildDescribePromptTruncation(t *testing.T) {
	longContent := strings.Repeat("x", 600)
	messages := []Message{
		{Role: RoleUser, Content: longContent},
	}

	prompt := buildDescribePrompt(messages)

	// Should be truncated to 500 chars + "..."
	if strings.Contains(prompt, strings.Repeat("x", 501)) {
		t.Error("long messages should be truncated to 500 characters")
	}
	if !strings.Contains(prompt, strings.Repeat("x", 500)+"...") {
		t.Error("truncated messages should end with '...'")
	}
}

func TestCleanDescription(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{`"Fix auth token refresh bug"`, "Fix auth token refresh bug"},
		{`'Add dark mode to settings'`, "Add dark mode to settings"},
		{"  Refactor resolvers  \n", "Refactor resolvers"},
		{"No quotes here", "No quotes here"},
		{`""`, ""},
	}

	for _, tt := range tests {
		got := cleanDescription(tt.input)
		if got != tt.want {
			t.Errorf("cleanDescription(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestReadOutputFirstResponseCallback(t *testing.T) {
	// Simulate a first-spawn session where isFirstSpawn=true.
	// Verify the onFirstResponse callback fires after the first eventResult.
	lines := strings.Join([]string{
		`{"type":"content_block_start","content_block":{"type":"text","text":""}}`,
		`{"type":"content_block_delta","delta":{"type":"text_delta","text":"Hello, I'll help you."}}`,
		`{"type":"result","session_id":"sess-first"}`,
	}, "\n")

	callbackCalled := make(chan struct{}, 1)

	m := &Manager{
		sessions:    make(map[string]*Session),
		processes:   make(map[string]*runningProcess),
		subscribers: make(map[string][]chan struct{}),
		onFirstResponse: func(beanID string, messages []Message) {
			if beanID != "wt-test" {
				t.Errorf("expected beanID 'wt-test', got %q", beanID)
			}
			if len(messages) < 2 {
				t.Errorf("expected at least 2 messages, got %d", len(messages))
			}
			callbackCalled <- struct{}{}
		},
	}

	session := &Session{
		ID:           "wt-test",
		AgentType:    "claude",
		Status:       StatusRunning,
		Messages:     []Message{{Role: RoleUser, Content: "Fix the auth bug"}},
		streamingIdx: -1,
	}
	m.sessions["wt-test"] = session

	// isFirstSpawn=true should trigger the callback
	m.readOutput("wt-test", strings.NewReader(lines), "", true)

	// The callback runs in a goroutine — wait for it with a timeout
	select {
	case <-callbackCalled:
		// success
	case <-time.After(time.Second):
		t.Fatal("onFirstResponse callback was not called within timeout")
	}
}

func TestReadOutputNoCallbackOnSubsequentSpawn(t *testing.T) {
	lines := strings.Join([]string{
		`{"type":"content_block_start","content_block":{"type":"text","text":""}}`,
		`{"type":"content_block_delta","delta":{"type":"text_delta","text":"Resumed."}}`,
		`{"type":"result","session_id":"sess-2"}`,
	}, "\n")

	callbackCalled := false

	m := &Manager{
		sessions:    make(map[string]*Session),
		processes:   make(map[string]*runningProcess),
		subscribers: make(map[string][]chan struct{}),
		onFirstResponse: func(beanID string, messages []Message) {
			callbackCalled = true
		},
	}

	session := &Session{
		ID:           "wt-test2",
		AgentType:    "claude",
		Status:       StatusRunning,
		Messages:     []Message{{Role: RoleUser, Content: "Continue"}},
		streamingIdx: -1,
	}
	m.sessions["wt-test2"] = session

	// isFirstSpawn=false should NOT trigger the callback
	m.readOutput("wt-test2", strings.NewReader(lines), "", false)

	if callbackCalled {
		t.Error("onFirstResponse should not be called when isFirstSpawn=false")
	}
}
