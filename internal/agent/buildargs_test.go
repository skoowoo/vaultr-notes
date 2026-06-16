package agent

import "testing"

func TestBuildClaudeSessionArgs(t *testing.T) {
	def := GetAgentDef("claude")
	if def == nil {
		t.Fatal("claude def missing")
	}
	first := BuildInvocationArgs(def, BuildArgsContext{SessionID: "abc-123", FirstSession: true})
	if !containsSeq(first, "--session-id", "abc-123") {
		t.Fatalf("first session args = %v, want --session-id abc-123", first)
	}
	resume := BuildInvocationArgs(def, BuildArgsContext{SessionID: "abc-123"})
	if !containsSeq(resume, "--resume", "abc-123") {
		t.Fatalf("resume args = %v, want --resume abc-123", resume)
	}
}

func TestBuildCodexSessionArgs(t *testing.T) {
	def := GetAgentDef("codex")
	if def == nil {
		t.Fatal("codex def missing")
	}
	resume := BuildInvocationArgs(def, BuildArgsContext{SessionID: "thread-1", Cwd: "/tmp"})
	if !containsSeq(resume, "exec", "resume", "thread-1") {
		t.Fatalf("resume args = %v", resume)
	}
}

func TestBuildOpenCodeSessionArgs(t *testing.T) {
	def := GetAgentDef("opencode")
	if def == nil {
		t.Fatal("opencode def missing")
	}
	args := BuildInvocationArgs(def, BuildArgsContext{SessionID: "sess-9", Cwd: "/tmp"})
	if !containsSeq(args, "-s", "sess-9") {
		t.Fatalf("args = %v", args)
	}
	for _, a := range args {
		if a == "-" {
			t.Fatalf("session resume must not include stdin sentinel '-': %v", args)
		}
	}
}

func containsSeq(argv []string, seq ...string) bool {
outer:
	for i := 0; i+len(seq) <= len(argv); i++ {
		for j, s := range seq {
			if argv[i+j] != s {
				continue outer
			}
		}
		return true
	}
	return false
}
