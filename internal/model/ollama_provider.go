package model

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

// OllamaProvider implements ModelProvider using a local Ollama daemon or the ollama CLI as fallback.
type OllamaProvider struct {
	host    string
	modelID string
	timeout time.Duration
	useCLI  bool
}

func NewOllamaProvider(host, modelID string, timeoutSeconds int, useCLI bool) *OllamaProvider {
	if host == "" {
		host = "http://localhost:11434"
	}
	if timeoutSeconds <= 0 {
		timeoutSeconds = 30
	}
	return &OllamaProvider{host: host, modelID: modelID, timeout: time.Duration(timeoutSeconds) * time.Second, useCLI: useCLI}
}

func (p *OllamaProvider) Name() string { return "ollama" }

// Generate calls the Ollama HTTP generate endpoint synchronously, or falls back to CLI.
func (p *OllamaProvider) Generate(ctx context.Context, prompt string) (string, error) {
	if !p.useCLI {
		res, err := p.generateHTTP(ctx, prompt)
		if err == nil {
			return res, nil
		}
		// fall through to CLI if HTTP fails
	}
	return p.generateCLI(ctx, prompt)
}

func (p *OllamaProvider) Stream(ctx context.Context, prompt string) (<-chan string, error) {
	out := make(chan string, 16)
	if !p.useCLI {
		ch, err := p.streamHTTP(ctx, prompt)
		if err == nil {
			return ch, nil
		}
		// fall through to CLI
	}
	// CLI streaming
	go func() {
		defer close(out)
		if err := p.streamCLI(ctx, prompt, out); err != nil {
			// send error as single chunk
			out <- fmt.Sprintf("error: %v", err)
		}
	}()
	return out, nil
}

func (p *OllamaProvider) generateHTTP(ctx context.Context, prompt string) (string, error) {
	client := &http.Client{Timeout: p.timeout}
	reqBody := map[string]interface{}{"model": p.modelID, "prompt": prompt}
	b, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, "POST", strings.TrimRight(p.host, "/")+"/api/generate", strings.NewReader(string(b)))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if key := os.Getenv("OPENAI_API_KEY"); key != "" {
		req.Header.Set("Authorization", "Bearer "+key)
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama HTTP error: %d %s", resp.StatusCode, string(body))
	}
	var parsed struct {
		Output string `json:"output"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", err
	}
	return parsed.Output, nil
}

func (p *OllamaProvider) streamHTTP(ctx context.Context, prompt string) (<-chan string, error) {
	client := &http.Client{Timeout: 0}
	// no timeout for streaming; use ctx deadline
	reqBody := map[string]interface{}{"model": p.modelID, "prompt": prompt, "stream": true}
	b, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, "POST", strings.TrimRight(p.host, "/")+"/api/generate", strings.NewReader(string(b)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if key := os.Getenv("OPENAI_API_KEY"); key != "" {
		req.Header.Set("Authorization", "Bearer "+key)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("ollama stream error: %d %s", resp.StatusCode, string(body))
	}

	out := make(chan string, 16)
	go func() {
		defer resp.Body.Close()
		defer close(out)
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			// try parse JSON chunk
			var chunk map[string]interface{}
			if err := json.Unmarshal([]byte(line), &chunk); err == nil {
				if t, ok := chunk["token"].(string); ok {
					out <- t
					continue
				}
				if o, ok := chunk["output"].(string); ok {
					out <- o
					continue
				}
			}
			// fallback: emit raw line
			out <- line
			select {
			case <-ctx.Done():
				return
			default:
			}
		}
		if err := scanner.Err(); err != nil && !errors.Is(err, io.EOF) {
			out <- fmt.Sprintf("error: %v", err)
		}
	}()
	return out, nil
}

func (p *OllamaProvider) generateCLI(ctx context.Context, prompt string) (string, error) {
	if p.modelID == "" {
		return "", errors.New("model id is required for CLI fallback")
	}
	cmd := exec.CommandContext(ctx, "ollama", "run", p.modelID, "-p", prompt)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("ollama cli error: %w - %s", err, string(out))
	}
	return string(out), nil
}

func (p *OllamaProvider) streamCLI(ctx context.Context, prompt string, out chan<- string) error {
	if p.modelID == "" {
		return errors.New("model id is required for CLI fallback")
	}
	cmd := exec.CommandContext(ctx, "ollama", "run", p.modelID, "-p", prompt)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	reader := bufio.NewReader(stdout)
	for {
		line, err := reader.ReadString('\n')
		if line != "" {
			out <- strings.TrimRight(line, "\n")
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}
		select {
		case <-ctx.Done():
			_ = cmd.Process.Kill()
			return ctx.Err()
		default:
		}
	}
	return cmd.Wait()
}
