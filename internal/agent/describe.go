package agent

import (
	"context"
	"log"
	"os/exec"
	"strings"
	"time"
)

const describePrompt = `You are given the beginning of an AI agent's conversation in a software development workspace. Summarize what this workspace is doing in 3-8 words. Be specific and concrete — mention the actual feature, bug, or task. Output ONLY the summary, nothing else.

Examples of good summaries:
- "Fix auth token refresh bug"
- "Add dark mode to settings"
- "Refactor GraphQL subscription resolvers"
- "Implement workspace description generation"

Conversation:`

// buildDescribePrompt constructs the prompt for the description generator
// from a list of conversation messages. Exported for testing.
func buildDescribePrompt(messages []Message) string {
	var sb strings.Builder
	sb.WriteString(describePrompt)
	sb.WriteString("\n\n")

	for _, msg := range messages {
		switch msg.Role {
		case RoleUser:
			sb.WriteString("User: ")
			sb.WriteString(truncate(msg.Content, 500))
			sb.WriteString("\n\n")
		case RoleAssistant:
			sb.WriteString("Assistant: ")
			sb.WriteString(truncate(msg.Content, 500))
			sb.WriteString("\n\n")
		}
	}

	return sb.String()
}

// cleanDescription trims whitespace and strips surrounding quotes from
// a raw model response.
func cleanDescription(raw string) string {
	desc := strings.TrimSpace(raw)
	desc = strings.Trim(desc, "\"'")
	return desc
}

// truncate returns s truncated to maxLen characters with "..." appended if needed.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// GenerateDescription runs a lightweight Claude call to summarize what a workspace
// is doing based on its initial conversation. Returns the description or empty string on error.
func GenerateDescription(messages []Message) string {
	prompt := buildDescribePrompt(messages)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Pass prompt via stdin (not as CLI arg) to avoid leaking conversation
	// content into `ps` output and to avoid OS argument length limits.
	cmd := exec.CommandContext(ctx, "claude", "--print", "-m", "haiku")
	cmd.Env = buildClaudeEnv()
	cmd.Stdin = strings.NewReader(prompt)

	out, err := cmd.Output()
	if err != nil {
		log.Printf("[describe] failed to generate workspace description: %v", err)
		return ""
	}

	return cleanDescription(string(out))
}
