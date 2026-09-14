package backend

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"ant-chrome/backend/internal/automation"

	"github.com/google/uuid"
)

const (
	automationScriptRunDefaultTimeout = 5 * time.Minute
	automationScriptRunMinTimeout     = 1 * time.Second
	automationScriptRunMaxTimeout     = 30 * time.Minute
)

func (a *App) automationScriptRunStore() *automation.ScriptRunStore {
	return automation.NewScriptRunStore(a.resolveAppPath(filepath.ToSlash(filepath.Join("data", "automation", "runs"))))
}

func (a *App) AutomationScriptRunList(limit int) ([]automation.ScriptRunRecord, error) {
	return a.automationScriptRunStore().List(limit)
}

func (a *App) AutomationScriptRun(scriptID string) (*automation.ScriptRunRecord, error) {
	return a.AutomationScriptRunWithOptions(automation.ScriptRunRequest{
		ScriptID:          scriptID,
		UseScriptSelector: true,
		UseScriptParams:   true,
	})
}

func (a *App) AutomationScriptRunGet(runID string) (*automation.ScriptRunRecord, error) {
	record, err := a.automationScriptRunStore().Get(runID)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("run not found")
		}
		return nil, err
	}
	return &record, nil
}

func (a *App) AutomationScriptStartRun(input automation.ScriptRunRequest) (*automation.ScriptRunRecord, error) {
	scriptID := strings.TrimSpace(input.ScriptID)
	if scriptID == "" {
		return nil, fmt.Errorf("scriptId is required")
	}

	store := a.automationScriptStore()
	if err := a.ensureAutomationScriptDefaults(store); err != nil {
		return nil, err
	}
	script, err := store.Get(scriptID)
	if err != nil {
		return nil, err
	}

	startedAt := time.Now()
	pending := automation.ScriptRunRecord{
		ID:         uuid.NewString(),
		ScriptID:   script.ID,
		ScriptName: script.Name,
		ScriptType: script.Type,
		Status:     "running",
		Summary:    "脚本执行中",
		StartedAt:  startedAt.Format(time.RFC3339),
	}
	saved, err := a.automationScriptRunStore().Save(pending)
	if err != nil {
		return nil, err
	}

	input.RunID = saved.ID
	go func() {
		defer func() {
			if recovered := recover(); recovered != nil {
				failed := saved
				failed.Status = "failed"
				failed.Summary = "脚本执行失败"
				failed.Error = fmt.Sprintf("%v", recovered)
				failed.FinishedAt = time.Now().Format(time.RFC3339)
				failed.DurationMs = time.Since(startedAt).Milliseconds()
				_, _ = a.automationScriptRunStore().Save(failed)
			}
		}()
		_, _ = a.AutomationScriptRunWithOptions(input)
	}()

	return &saved, nil
}

func (a *App) AutomationScriptRunWithOptions(input automation.ScriptRunRequest) (*automation.ScriptRunRecord, error) {
	startedAt := time.Now()
	run := automation.ScriptRunRecord{
		ID:        strings.TrimSpace(input.RunID),
		ScriptID:  strings.TrimSpace(input.ScriptID),
		Status:    "failed",
		StartedAt: startedAt.Format(time.RFC3339),
	}

	store := a.automationScriptStore()
	if err := a.ensureAutomationScriptDefaults(store); err != nil {
		run.Summary = "脚本读取失败"
		run.Error = err.Error()
		return a.finalizeAutomationScriptRun(run, startedAt)
	}

	script, err := store.Get(run.ScriptID)
	if err != nil {
		run.Summary = "脚本读取失败"
		run.Error = err.Error()
		return a.finalizeAutomationScriptRun(run, startedAt)
	}

	run.ScriptName = script.Name
	run.ScriptType = script.Type

	runCtx := a.ctx
	if runCtx == nil {
		runCtx = context.Background()
	}
	runCtx, cancel := context.WithTimeout(runCtx, automationScriptRunTimeout(input))
	defer cancel()

	switch script.Type {
	case "launch-api":
		resultText, summary, errText := a.runLaunchAPIScript(runCtx, script, input)
		run.ResultText = resultText
		run.Summary = summary
		run.Error = errText
		if errText == "" {
			run.Status = "success"
		}
	case "playwright-cdp":
		resultText, logText, summary, errText := a.runPlaywrightScript(runCtx, script, input)
		run.ResultText = resultText
		run.LogText = logText
		run.Summary = summary
		run.Error = errText
		if errText == "" {
			run.Status = "success"
		}
	default:
		run.Summary = "当前脚本类型暂不支持直接执行"
		run.Error = fmt.Sprintf("script type %q is not supported yet", script.Type)
	}

	return a.finalizeAutomationScriptRun(run, startedAt)
}

func automationScriptRunTimeout(input automation.ScriptRunRequest) time.Duration {
	if input.TimeoutMs <= 0 {
		return automationScriptRunDefaultTimeout
	}
	timeout := time.Duration(input.TimeoutMs) * time.Millisecond
	if timeout < automationScriptRunMinTimeout {
		return automationScriptRunMinTimeout
	}
	if timeout > automationScriptRunMaxTimeout {
		return automationScriptRunMaxTimeout
	}
	return timeout
}

func automationRunContextErrorMessage(err error) string {
	if err == context.DeadlineExceeded {
		return "自动化任务超时，已终止"
	}
	if err == context.Canceled {
		return "自动化任务已取消"
	}
	return err.Error()
}

func (a *App) finalizeAutomationScriptRun(run automation.ScriptRunRecord, startedAt time.Time) (*automation.ScriptRunRecord, error) {
	run.FinishedAt = time.Now().Format(time.RFC3339)
	run.DurationMs = time.Since(startedAt).Milliseconds()
	saved, err := a.automationScriptRunStore().Save(run)
	if err != nil {
		return nil, err
	}
	return &saved, nil
}
