package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/hardhacker/vaultr/internal/agent"
	"github.com/hardhacker/vaultr/internal/client"
	"github.com/spf13/cobra"
)

func newAgentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agent",
		Short: "Local coding agent adapters (requires running server)",
		Long: `Commands for integrating with locally installed coding CLIs exposed by the
Vaultr server (/api/agents, etc.). The server probes PATH and publishes model
hints compatible with Open Design.

Subcommands:
  agent list          — adapters and model counts
  agent models -a ID  — model ids and reasoning presets for chat
  agent chat          — POST /api/chat (SSE)`,
		SilenceUsage: true,
	}
	cmd.AddCommand(newAgentListCmd())
	cmd.AddCommand(newAgentModelsCmd())
	cmd.AddCommand(newAgentChatCmd())
	return cmd
}

func newAgentListCmd() *cobra.Command {
	var table bool

	cmd := &cobra.Command{
		Use:          "list",
		Short:        "List agent adapters from the server (GET /api/agents)",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := openClient()
			if err != nil {
				return err
			}
			agents, err := c.AgentsList(context.Background())
			if err != nil {
				return err
			}
			if table {
				return printAgentsTable(agents)
			}
			out := struct {
				Agents []agent.AgentInfo `json:"agents"`
			}{Agents: agents}
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(out)
		},
	}

	cmd.Flags().BoolVarP(&table, "table", "t", false, "output in human-readable table format")

	return cmd
}

func printAgentsTable(agents []agent.AgentInfo) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "ID\tNAME\tAVAILABLE\tVERSION\tSTREAM\tPATH\tMODELS")
	for _, a := range agents {
		path := a.Path
		if path == "" {
			path = "—"
		}
		ver := a.Version
		if ver == "" {
			ver = "—"
		}
		avail := "no"
		if a.Available {
			avail = "yes"
		}
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%d\n",
			a.ID, a.Name, avail, ver, string(a.StreamFormat), path, len(a.Models))
	}
	if err := w.Flush(); err != nil {
		return err
	}
	if v := os.Getenv("VAULTR_AGENT_LIST_MODELS"); v == "1" || v == "true" {
		for _, a := range agents {
			if len(a.Models) == 0 {
				continue
			}
			fmt.Printf("\n%s (%s):\n", a.ID, a.Name)
			for _, m := range a.Models {
				label := m.Label
				if label == "" {
					label = m.ID
				}
				fmt.Printf("  %s\n", label)
			}
		}
	}
	return nil
}

func newAgentModelsCmd() *cobra.Command {
	var (
		agentID string
		asJSON  bool
	)
	cmd := &cobra.Command{
		Use:          "models",
		Short:        "Print model ids for one adapter from GET /api/agents",
		Long: strings.TrimSpace(`
Calls GET /api/agents and shows models (and reasoning options, if any) for --agent.

These IDs are typical values for "agent chat --model". Omitting --model leaves the
CLI default (nothing is forwarded to the subprocess).`),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			id := strings.TrimSpace(agentID)
			if id == "" {
				return fmt.Errorf("--agent / -a is required")
			}
			c, err := openClient()
			if err != nil {
				return err
			}
			agents, err := c.AgentsList(cmd.Context())
			if err != nil {
				return err
			}
			var sel *agent.AgentInfo
			for i := range agents {
				if agents[i].ID == id {
					sel = &agents[i]
					break
				}
			}
			if sel == nil {
				return fmt.Errorf("no adapter %q in server list (run: vaultr agent list)", id)
			}
			out := cmd.OutOrStdout()
			if asJSON {
				enc := json.NewEncoder(out)
				enc.SetIndent("", "  ")
				return enc.Encode(map[string]any{
					"agentId":          sel.ID,
					"name":             sel.Name,
					"available":        sel.Available,
					"models":           sel.Models,
					"reasoningOptions": sel.ReasoningOptions,
				})
			}
			_, _ = fmt.Fprintf(out, "agent: %s (%s)\n", sel.ID, sel.Name)
			if len(sel.Models) == 0 {
				_, _ = fmt.Fprintf(out, "models: (none in API response — omit --model so %s uses its binary default)\n", sel.Name)
			} else {
				_, _ = fmt.Fprintln(out, "models (--model):")
				w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
				_, _ = fmt.Fprintln(w, "ID\tLABEL")
				for _, m := range sel.Models {
					label := m.Label
					if label == "" {
						label = m.ID
					}
					_, _ = fmt.Fprintf(w, "%s\t%s\n", m.ID, label)
				}
				_ = w.Flush()
				_, _ = fmt.Fprintf(out, "Empty --model: subprocess default (no flag). Other printable ids may still work via server sanitization.\n")
			}
			if len(sel.ReasoningOptions) > 0 {
				_, _ = fmt.Fprintln(out, "\nreasoning (--reasoning):")
				w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
				_, _ = fmt.Fprintln(w, "ID\tLABEL")
				for _, r := range sel.ReasoningOptions {
					lbl := r.Label
					if lbl == "" {
						lbl = r.ID
					}
					_, _ = fmt.Fprintf(w, "%s\t%s\n", r.ID, lbl)
				}
				_ = w.Flush()
				_, _ = fmt.Fprintln(out, "Omit or use \"default\" for adapter default.")
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&agentID, "agent", "a", "", "adapter id (e.g. claude, codex, cursor-agent)")
	cmd.Flags().BoolVar(&asJSON, "json", false, "print one JSON envelope (agentId, models, reasoningOptions)")
	_ = cmd.MarkFlagRequired("agent")
	return cmd
}

func newAgentChatCmd() *cobra.Command {
	var (
		agentID   string
		message   string
		model     string
		reasoning string
		system    string
		cwd       string
		jsonLines bool
	)

	cmd := &cobra.Command{
		Use:          "chat [message]",
		Short:        "Send one chat turn via POST /api/chat (server streams SSE)",
		Long: strings.TrimSpace(`
Message can be positional text, --message, or piped stdin (when stdin is not a TTY).

Requires a running Vaultr server that can spawn the chosen agent CLI (see agent list).

Model: if you omit --model, Vaultr does not pass --model (or equivalent) to the CLI;
each binary uses its own default. Allowed values are ids from server-reported catalogs
(use: vaultr agent models -a NAME) plus other ids that pass server sanitization when
not listed.

By default assistant text deltas are printed to stdout; auxiliary lines go to stderr.

chat --json: stream each SSE event as one JSON line. agent models --json: structured model JSON.`),
		Example: `  vaultr agent chat --agent claude "Summarize ~/.vaultr in one sentence."
  cat prompt.txt | vaultr agent chat -a codex`,
		SilenceUsage: true,
		Args:         cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			msg, err := resolveAgentChatMessage(message, args)
			if err != nil {
				return err
			}
			id := strings.TrimSpace(agentID)
			if id == "" {
				return fmt.Errorf("--agent is required")
			}
			c, err := openClient()
			if err != nil {
				return err
			}
			req := client.AgentChatRequest{
				AgentID: id, Message: msg, Model: strings.TrimSpace(model),
				Reasoning: strings.TrimSpace(reasoning), SystemPrompt: strings.TrimSpace(system),
				Cwd: strings.TrimSpace(cwd),
			}
			handler := chatHumanPrinter
			if jsonLines {
				handler = chatJSONLines
			}
			return c.AgentChatSSE(cmd.Context(), req, handler)
		},
	}

	cmd.Flags().StringVarP(&agentID, "agent", "a", "", `agent adapter id (e.g. claude, codex); required`)
	cmd.Flags().StringVarP(&message, "message", "m", "", "prompt text (else use args or stdin)")
	cmd.Flags().StringVar(&model, "model", "", "model id from \"agent models -a\"; omit for CLI binary default")
	cmd.Flags().StringVar(&reasoning, "reasoning", "", `reasoning id from "agent models -a"; omit or default for builtin default`)
	cmd.Flags().StringVar(&system, "system", "", "system / instructions prefix (sent as systemPrompt)")
	cmd.Flags().StringVar(&cwd, "cwd", "", "working directory for the agent subprocess (default: vault root on server)")
	cmd.Flags().BoolVar(&jsonLines, "json", false, "stream each SSE event as one JSON line (event + data)")
	_ = cmd.MarkFlagRequired("agent")

	return cmd
}

func resolveAgentChatMessage(flag string, args []string) (string, error) {
	s := strings.TrimSpace(flag)
	if s != "" {
		return s, nil
	}
	if len(args) > 0 {
		return strings.TrimSpace(strings.Join(args, " ")), nil
	}
	if stdinIsRedirected() {
		b, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("read stdin: %w", err)
		}
		return strings.TrimSpace(string(b)), nil
	}
	return "", fmt.Errorf("message required: pass as arguments, --message / -m, or pipe stdin")
}

func stdinIsRedirected() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice == 0
}

// chatSSELine mirrors one SSE payload for JSON output mode.
type chatSSELine struct {
	Event string          `json:"event"`
	Data  json.RawMessage `json:"data"`
}

func chatJSONLines(event string, data json.RawMessage) error {
	enc := json.NewEncoder(os.Stdout)
	return enc.Encode(chatSSELine{Event: event, Data: data})
}

func chatHumanPrinter(event string, data json.RawMessage) error {
	switch event {
	case "start", "heartbeat":
		return nil
	case "stderr", "stdout":
		var m struct {
			Chunk string `json:"chunk"`
		}
		if json.Unmarshal(data, &m) == nil && m.Chunk != "" {
			dst := os.Stderr
			if event == "stdout" {
				dst = os.Stdout
			}
			_, _ = fmt.Fprint(dst, m.Chunk)
		}
	case "agent":
		var m map[string]any
		if json.Unmarshal(data, &m) != nil {
			return nil
		}
		switch m["type"] {
		case "text_delta":
			if d := toAgentStr(m["delta"]); d != "" {
				fmt.Print(d)
			}
		case "thinking_delta":
			if d := toAgentStr(m["delta"]); d != "" {
				_, _ = fmt.Fprint(os.Stderr, d)
			}
		case "raw":
			if line := toAgentStr(m["line"]); line != "" {
				_, _ = fmt.Fprintln(os.Stderr, line)
			}
		case "tool_use":
			name := toAgentStr(m["name"])
			if name != "" {
				_, _ = fmt.Fprintf(os.Stderr, "[tool] %s\n", name)
			}
		case "tool_result":
			short := strings.TrimSpace(toAgentStr(m["content"]))
			if len(short) > 120 {
				short = short[:120] + "…"
			}
			if short != "" {
				_, _ = fmt.Fprintf(os.Stderr, "[tool result] %s\n", short)
			}
		case "error":
			_, _ = fmt.Fprintf(os.Stderr, "[agent stream] %s\n", toAgentStr(m["message"]))
		case "status":
			_, _ = fmt.Fprintf(os.Stderr, "[%s]\n", toAgentStr(m["label"]))
		}
	case "error":
		var m map[string]any
		if json.Unmarshal(data, &m) == nil && m["message"] != nil {
			_, _ = fmt.Fprintf(os.Stderr, "error: %v\n", m["message"])
		}
	case "end":
		var m map[string]any
		_ = json.Unmarshal(data, &m)
		switch st := m["status"].(type) {
		case string:
			if st != "succeeded" {
				code, _ := m["exitCode"].(float64)
				_, _ = fmt.Fprintf(os.Stderr, "\n(chat finished with status %s, exit %v)\n", st, int(code))
			}
		}
	}
	return nil
}

func toAgentStr(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case json.Number:
		return t.String()
	default:
		return ""
	}
}
