// Learn Agent the Hard Way — 练习 10：误删保护——删除前备份
//
// 练习 9 拦住了模型不该做的事。但 write_file 覆盖一个已存在的文件，从练习 6 到现在
// 从来都是该做的事——问题不在"要不要允许"，在"允许之后，旧内容还有没有第二次机会"。
// 这一章补上覆盖前留一份退路。
package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ---- base prompt：给模型的说明书 ----

const basePrompt = `你是一个能操作本地文件和 shell 的助手，通过工具真正执行动作，而不是描述打算做什么。

- 能用 read_file / write_file / edit_file 完成的事，优先用它们；bash 留给专用工具做不到的事（跑测试、跑 git、装依赖、查系统信息）。
- 修改一个已经存在的文件前，必须先用 read_file 读过它一遍——这条规矩不因为你换了工具执行修改就不算数：用 bash 的 echo / sed / tee 等方式直接改文件内容，同样要先读一遍再动手。能用 edit_file 完成的局部修改，优先用 edit_file 而不是 sed -i，这样改动会经过校验，而不是绕开它。
- 只做任务要求的改动，不顺手重构、不改无关代码。`

// ---- 权限层：拦下危险命令 ----

type decision int

const (
	decisionAllow decision = iota
	decisionAsk
	decisionDeny
)

type permRule struct {
	pattern string
	decide  decision
}

var bashRules = []permRule{
	{"rm -rf /", decisionDeny},
	{"rm -rf ~", decisionDeny},
	{"rm -rf", decisionAsk},
	{"sudo ", decisionAsk},
	{"git push --force", decisionAsk},
	{"curl ", decisionAsk},
	{"ls", decisionAllow},
	{"cat ", decisionAllow},
	{"pwd", decisionAllow},
	{"echo ", decisionAllow},
	{"git status", decisionAllow},
}

func classifyBash(cmd string) decision {
	for _, r := range bashRules {
		if r.decide == decisionDeny && strings.Contains(cmd, r.pattern) {
			return decisionDeny
		}
	}
	for _, r := range bashRules {
		if r.decide == decisionAsk && strings.Contains(cmd, r.pattern) {
			return decisionAsk
		}
	}
	for _, r := range bashRules {
		if r.decide != decisionAllow || !strings.Contains(cmd, r.pattern) {
			continue
		}
		trimmed := strings.TrimLeft(cmd, " \t")
		if strings.HasPrefix(trimmed, r.pattern) && !containsShellChain(cmd) {
			return decisionAllow
		}
	}
	return decisionAsk
}

func containsShellChain(cmd string) bool {
	return strings.ContainsAny(cmd, ";|&$()`\n")
}

func askApproval(cmd string) bool {
	fmt.Fprintf(os.Stderr, "\n⚠️  模型想执行: %s\n允许吗？(y/N) ", cmd)
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return false
	}
	answer := strings.ToLower(strings.TrimSpace(line))
	return answer == "y" || answer == "yes"
}

func commandOf(args string) string {
	var in struct {
		Command string `json:"command"`
	}
	_ = json.Unmarshal([]byte(args), &in)
	return in.Command
}

// ---- 备份层：覆盖前留一份 ----

// trashDir 是备份落地的地方，就在工作目录底下——足够找、足够简单。
const trashDir = ".trash"

// backupIfExists 在覆盖一个已存在的文件前，把旧内容原样复制进 trashDir，
// 文件名前缀时间戳避免撞名。目标文件本来就不存在时什么都不做，返回空字符串。
func backupIfExists(path string) (string, error) {
	if !fileExists(path) {
		return "", nil
	}
	if err := os.MkdirAll(trashDir, 0o755); err != nil {
		return "", err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	ts := time.Now().Format("20060102-150405")
	dest := filepath.Join(trashDir, ts+"_"+filepath.Base(path))
	if err := os.WriteFile(dest, data, 0o644); err != nil {
		return "", err
	}
	return dest, nil
}

// restore 找 trashDir 里这个文件名最新的一份备份，写回原路径。恢复动作本身也先给
// "现在这份"备份一次——误删保护对自己也生效，不会因为你手滑恢复错了版本就丢掉当前内容。
func restore(path string) int {
	entries, err := os.ReadDir(trashDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "错误: 没有找到", trashDir, "目录，或读取失败:", err)
		return 1
	}
	suffix := "_" + filepath.Base(path)
	var newest string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), suffix) && e.Name() > newest {
			newest = e.Name()
		}
	}
	if newest == "" {
		fmt.Fprintf(os.Stderr, "错误: %s 里没有 %s 的备份\n", trashDir, filepath.Base(path))
		return 1
	}
	if _, err := backupIfExists(path); err != nil {
		fmt.Fprintln(os.Stderr, "错误: 备份当前版本失败，为安全起见拒绝恢复:", err)
		return 1
	}
	data, err := os.ReadFile(filepath.Join(trashDir, newest))
	if err != nil {
		fmt.Fprintln(os.Stderr, "错误:", err)
		return 1
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "错误:", err)
		return 1
	}
	fmt.Printf("已从 %s 恢复到 %s\n", filepath.Join(trashDir, newest), path)
	return 0
}

// ---- 工具层 ----

type toolSpec struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

type tool interface {
	definition() toolSpec
	execute(args string) string
}

type readFileTool struct{}

func (readFileTool) definition() toolSpec {
	return toolSpec{
		Name:        "read_file",
		Description: "读取一个本地文件，返回它的文本内容。修改文件前必须先用它读一遍。",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{"type": "string", "description": "要读取的文件路径"},
			},
			"required": []string{"path"},
		},
	}
}

func (readFileTool) execute(args string) string {
	var in struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal([]byte(args), &in); err != nil {
		return "错误: 参数不是合法 JSON: " + err.Error()
	}
	data, err := os.ReadFile(in.Path)
	if err != nil {
		return "错误: " + err.Error()
	}
	return string(data)
}

type writeFileTool struct{}

func (writeFileTool) definition() toolSpec {
	return toolSpec{
		Name:        "write_file",
		Description: "把内容完整写入一个文件。文件不存在就创建，存在就整个覆盖。",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path":    map[string]any{"type": "string", "description": "目标文件路径"},
				"content": map[string]any{"type": "string", "description": "要写入的完整内容"},
			},
			"required": []string{"path", "content"},
		},
	}
}

func (writeFileTool) execute(args string) string {
	var in struct {
		Path    string `json:"path"`
		Content string `json:"content"`
	}
	if err := json.Unmarshal([]byte(args), &in); err != nil {
		return "错误: 参数不是合法 JSON: " + err.Error()
	}
	// 覆盖前先把旧内容备份进 .trash/，留一份退路。备份失败则拒绝覆盖，安全第一。
	backup, err := backupIfExists(in.Path)
	if err != nil {
		return "错误: 备份旧内容失败，为安全起见拒绝覆盖: " + err.Error()
	}
	if err := os.WriteFile(in.Path, []byte(in.Content), 0o644); err != nil {
		return "错误: " + err.Error()
	}
	if backup != "" {
		return fmt.Sprintf("已把旧内容备份到 %s，然后写入 %s（%d 字节）", backup, in.Path, len(in.Content))
	}
	return fmt.Sprintf("已写入 %s（%d 字节）", in.Path, len(in.Content))
}

type editFileTool struct{}

func (editFileTool) definition() toolSpec {
	return toolSpec{
		Name: "edit_file",
		Description: "在已有文件里做一次精确替换。old_string 必须与文件现有内容逐字一致，" +
			"且只出现一次——不唯一时请带上足够的上下文再试。文件必须已存在（创建用 write_file）。",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path":       map[string]any{"type": "string", "description": "目标文件路径"},
				"old_string": map[string]any{"type": "string", "description": "要找到的原文，必须唯一"},
				"new_string": map[string]any{"type": "string", "description": "替换成的新文本，可以为空（等于删除）"},
			},
			"required": []string{"path", "old_string", "new_string"},
		},
	}
}

func (editFileTool) execute(args string) string {
	var in struct {
		Path      string `json:"path"`
		OldString string `json:"old_string"`
		NewString string `json:"new_string"`
	}
	if err := json.Unmarshal([]byte(args), &in); err != nil {
		return "错误: 参数不是合法 JSON: " + err.Error()
	}
	data, err := os.ReadFile(in.Path)
	if err != nil {
		return "错误: " + err.Error()
	}
	text := string(data)
	switch n := strings.Count(text, in.OldString); {
	case in.OldString == "":
		return "错误: old_string 不能为空"
	case n == 0:
		return "错误: old_string 在文件里找不到——和 read_file 看到的原文逐字对一下"
	case n > 1:
		return fmt.Sprintf("错误: old_string 出现了 %d 次，无法确定改哪一处——多带几行上下文让它唯一", n)
	}
	text = strings.Replace(text, in.OldString, in.NewString, 1)
	if err := os.WriteFile(in.Path, []byte(text), 0o644); err != nil {
		return "错误: " + err.Error()
	}
	return "已替换 " + in.Path + " 中的一处文本"
}

// ---- bash：特权工具 ----

const (
	defaultBashTimeout = 30 * time.Second
	maxBashTimeout     = 120 * time.Second
	maxBashOutput      = 8 * 1024
)

var workDir, _ = os.Getwd()

type bashTool struct{}

func (bashTool) definition() toolSpec {
	return toolSpec{
		Name: "bash",
		Description: "在系统 shell 里运行一条命令，返回 stdout 和 stderr。" +
			"命令总是在固定的工作目录执行，cd 不会跨调用生效。" +
			"默认 30 秒超时；预计更久就传 timeout（整数秒，上限 120）。" +
			"能用 read_file / write_file / edit_file 完成的事，优先用那些专用工具。",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"command": map[string]any{"type": "string", "description": "要执行的 shell 命令"},
				"timeout": map[string]any{"type": "integer", "description": "超时秒数，可选，默认 30，上限 120"},
			},
			"required": []string{"command"},
		},
	}
}

func (bashTool) execute(args string) string {
	var in struct {
		Command string `json:"command"`
		Timeout int    `json:"timeout"`
	}
	if err := json.Unmarshal([]byte(args), &in); err != nil {
		return "错误: 参数不是合法 JSON: " + err.Error()
	}
	if strings.TrimSpace(in.Command) == "" {
		return "错误: command 不能为空"
	}
	d := defaultBashTimeout
	if in.Timeout > 0 {
		d = time.Duration(in.Timeout) * time.Second
		if d > maxBashTimeout {
			return fmt.Sprintf("错误: timeout 最大 %d 秒。要跑更久的命令，把它拆小，或者放弃在一次调用里等它",
				int(maxBashTimeout.Seconds()))
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), d)
	defer cancel()
	cmd := exec.CommandContext(ctx, "sh", "-c", in.Command)
	cmd.Dir = workDir
	out, err := cmd.CombinedOutput()

	text := tail(string(out), maxBashOutput)
	if ctx.Err() == context.DeadlineExceeded {
		return fmt.Sprintf("错误: 命令超过 %s 被终止。被杀前的输出：\n%s", d, text)
	}
	if err != nil {
		return fmt.Sprintf("%s\n[%v]", text, err)
	}
	if text == "" {
		return "(命令成功，无输出)"
	}
	return text
}

func tail(s string, max int) string {
	if len(s) <= max {
		return s
	}
	cut := s[len(s)-max:]
	if i := strings.IndexByte(cut, '\n'); i >= 0 {
		cut = cut[i+1:]
	}
	return fmt.Sprintf("[... 前面 %d 字节被截断，只保留结尾 ...]\n%s", len(s)-len(cut), cut)
}

// ---- 注册表层 ----

type registry struct {
	tools   map[string]tool
	order   []string
	hasRead map[string]bool
}

func newRegistry(ts ...tool) *registry {
	r := &registry{tools: map[string]tool{}, hasRead: map[string]bool{}}
	for _, t := range ts {
		spec := t.definition()
		r.tools[spec.Name] = t
		r.order = append(r.order, spec.Name)
	}
	return r
}

func (r *registry) definitions() []map[string]any {
	var out []map[string]any
	for _, name := range r.order {
		out = append(out, map[string]any{
			"type":     "function",
			"function": r.tools[name].definition(),
		})
	}
	return out
}

func (r *registry) execute(name, args string) string {
	t, ok := r.tools[name]
	if !ok {
		return "错误: 未知工具 " + name
	}
	if name == "write_file" || name == "edit_file" {
		if path := pathOf(args); path != "" && fileExists(path) && !r.hasRead[path] {
			return "错误: " + path + " 已存在但这个会话里还没读过它。先用 read_file 看一眼，再来修改。"
		}
	}
	if name == "bash" {
		cmd := commandOf(args)
		switch classifyBash(cmd) {
		case decisionDeny:
			return "错误: 权限拒绝——这条命令匹配了硬性禁止规则，不会执行，也不会询问。"
		case decisionAsk:
			if !askApproval(cmd) {
				return "错误: 权限拒绝——用户没有批准这条命令。"
			}
		}
	}
	result := t.execute(args)
	if path := pathOf(args); path != "" && !strings.HasPrefix(result, "错误:") {
		r.hasRead[path] = true
	}
	return result
}

func pathOf(args string) string {
	var in struct {
		Path string `json:"path"`
	}
	_ = json.Unmarshal([]byte(args), &in)
	return in.Path
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// ---- 协议层 ----

type message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content"`
	ToolCalls  []toolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
}

type toolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type request struct {
	Model     string           `json:"model"`
	Messages  []message        `json:"messages"`
	MaxTokens int              `json:"max_tokens,omitempty"`
	Tools     []map[string]any `json:"tools,omitempty"`
}

type response struct {
	Choices []struct {
		Message      message `json:"message"`
		FinishReason string  `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		PromptTokensDetails struct {
			CachedTokens int `json:"cached_tokens"`
		} `json:"prompt_tokens_details"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

func main() {
	// 恢复模式从正常任务里分出去：./ex10 -restore list.txt
	if len(os.Args) == 3 && os.Args[1] == "-restore" {
		os.Exit(restore(os.Args[2]))
	}
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, `用法: ./ex10 "你的任务"  |  ./ex10 -restore <file>`)
		os.Exit(1)
	}
	apiKey := os.Getenv("OPENAI_API_KEY")
	model := os.Getenv("MODEL")
	if apiKey == "" || model == "" {
		fmt.Fprintln(os.Stderr, "需要环境变量 OPENAI_API_KEY 和 MODEL")
		fmt.Fprintln(os.Stderr, `例: export OPENAI_API_KEY=sk-xxxx`)
		fmt.Fprintln(os.Stderr, `    export MODEL=deepseek-v4-flash`)
		fmt.Fprintln(os.Stderr, `    export OPENAI_BASE_URL=https://api.deepseek.com/v1  # 不设则默认 OpenAI 官方`)
		os.Exit(1)
	}
	base := os.Getenv("OPENAI_BASE_URL")
	if base == "" {
		base = "https://api.openai.com/v1"
	}

	reg := newRegistry(
		readFileTool{},
		writeFileTool{},
		editFileTool{},
		bashTool{},
	)

	history := []message{
		{Role: "system", Content: basePrompt},
		{Role: "user", Content: os.Args[1]},
	}

	const maxRounds = 10
	for round := 1; round <= maxRounds; round++ {
		r, err := send(base, apiKey, model, history, reg.definitions())
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		msg := r.Choices[0].Message
		history = append(history, msg)

		if r.Choices[0].FinishReason != "tool_calls" {
			fmt.Println(msg.Content)
			fmt.Fprintf(os.Stderr, "\n[共 %d 轮 · 最后一轮输入 %d tokens（命中缓存 %d）· finish_reason=%s]\n",
				round, r.Usage.PromptTokens, r.Usage.PromptTokensDetails.CachedTokens, r.Choices[0].FinishReason)
			return
		}

		for _, tc := range msg.ToolCalls {
			fmt.Fprintf(os.Stderr, "[round %d 输入 %d tokens，命中缓存 %d]\n",
				round, r.Usage.PromptTokens, r.Usage.PromptTokensDetails.CachedTokens)
			fmt.Fprintf(os.Stderr, "[round %d] %s(%s)\n", round, tc.Function.Name, tc.Function.Arguments)
			result := reg.execute(tc.Function.Name, tc.Function.Arguments)
			history = append(history, message{
				Role:       "tool",
				ToolCallID: tc.ID,
				Content:    result,
			})
		}
	}
	fmt.Fprintf(os.Stderr, "达到 %d 轮上限，停止。\n", maxRounds)
	os.Exit(1)
}

func send(base, apiKey, model string, history []message, tools []map[string]any) (response, error) {
	var r response
	body, _ := json.Marshal(request{
		Model:     model,
		MaxTokens: 4096,
		Messages:  history,
		Tools:     tools,
	})
	req, err := http.NewRequest("POST", base+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return r, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return r, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return r, fmt.Errorf("HTTP %d: %s", resp.StatusCode, raw)
	}
	if err := json.Unmarshal(raw, &r); err != nil {
		return r, fmt.Errorf("解析失败: %w\n原始响应: %s", err, raw)
	}
	if r.Error != nil {
		return r, fmt.Errorf("API 错误 [%s]: %s", r.Error.Type, r.Error.Message)
	}
	if len(r.Choices) == 0 {
		return r, fmt.Errorf("空响应: %s", raw)
	}
	return r, nil
}
