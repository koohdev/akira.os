package tui

import (
	"fmt"
	"math/rand"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"meme-factory/internal/config"
	"meme-factory/internal/styles"
)

func (m model) renderHeader() string {
	graph := ""
	for _, h := range m.waveHistory {
		graph += styles.Bars[h]
	}

	titleBlock := styles.TxtRed.Render(fmt.Sprintf(`
    ▄████▄ ██ ▄█▀ ██ █████▄  ▄████▄ ©
    ██▄▄██ ████   ██ ██▄▄██▄ ██▄▄██ 
    ██  ██ ██ ▀█▄ ██ ██   ██ ██  ██ 											[+] 1988.
                                 
  %s // %s
  %s`, config.APP_AUTHOR, config.APP_VERSION, config.APP_DESC))
	return "\n" + titleBlock
}

func (m model) renderMusicStatus() string {
	musicStatus := styles.TxtGray.Render("AUDIO: STANDBY")
	if m.musicPlaying {
		musicStatus = styles.TxtPink.Render("♫ NOW PLAYING: " + m.currentSong)
	}
	return musicStatus
}

func (m model) renderNeuralSync() string {
	// Header
	header := styles.TxtOrange.Render(fmt.Sprintf("// NEURAL LINK CALIBRATION // SEQUENCE: %d/%d", m.typingPhase, m.typingTotalPhases))

	// Target Logic
	target := styles.TxtBlue.Render("TARGET:  " + m.typingTarget)

	// Input Render Logic (Colorizing correct/incorrect chars)
	coloredInput := ""
	for i, char := range m.typingInput {
		if i < len(m.typingTarget) {
			if byte(char) == m.typingTarget[i] {
				coloredInput += styles.TxtGreen.Render(string(char))
			} else {
				coloredInput += styles.TxtRed.Render(string(char))
			}
		} else {
			coloredInput += styles.TxtRed.Render(string(char)) // Extra chars are wrong
		}
	}

	// Cursor
	cursor := styles.TxtWhite.Blink(true).Render("█")
	inputLine := fmt.Sprintf("INPUT:   %s%s", coloredInput, cursor)

	// Time Remaining Bar (Simulated based on Phase Start + 30s)
	// For now, let's just make it a visual style bar or simple countdown if needed
	// The prompt showed a bar: [████████████] TIME REMAINING
	elapsed := time.Since(m.typingStartTime)
	timeLeft := m.typingTimerDuration - elapsed
	if timeLeft < 0 {
		timeLeft = 0
	}

	percentLeft := float64(timeLeft) / float64(m.typingTimerDuration)
	barWidth := 20
	filledLen := int(percentLeft * float64(barWidth))
	if filledLen < 0 {
		filledLen = 0
	}

	timeBar := fmt.Sprintf("[%s%s] TIME REMAINING", strings.Repeat("█", filledLen), strings.Repeat("░", barWidth-filledLen))
	if timeLeft < 5*time.Second {
		timeBar = styles.TxtRed.Render(timeBar)
	} else {
		timeBar = styles.TxtGreen.Render(timeBar)
	}

	stats := fmt.Sprintf("WPM: %d    ACCURACY: %d%%", m.typingCurrentWPM, m.typingAccuracy)

	footer := styles.TxtGray.Render("> KEEP TYPING...")

	return fmt.Sprintf("%s\n\n%s\n%s\n\n%s\n%s\n\n%s",
		header,
		target,
		inputLine,
		timeBar,
		styles.TxtWhite.Render(stats),
		footer,
	)
}

func (m model) renderCompressionView() string {
	lines := []string{
		styles.TxtOrange.Render("// GIGA-COMPRESS // ASSET OPTIMIZER v1.0"),
		styles.TxtGray.Render(fmt.Sprintf("Target: ./websites/%s/images/", m.compressionFolder)),
		"",
	}

	for _, entry := range m.compressionEntries {
		statusIcon := "[ ]"
		color := styles.TxtGray
		var percent int

		switch entry.Status {
		case "PROCESSING":
			statusIcon = "[>]"
			color = styles.TxtOrange
			if m.blinkState {
				percent = 65
			} else {
				percent = 35
			}
		case "OPTIMIZED":
			statusIcon = "[✓]"
			color = styles.TxtGreen
			percent = 100
		case "ERROR":
			statusIcon = "[X]"
			color = styles.TxtRed
			percent = 100
		default:
			percent = int(entry.Progress * 100)
		}

		bar := renderProgressBar(percent, 20)
		nameLine := color.Render(fmt.Sprintf("%-3s %-20s %s", statusIcon, entry.Name, bar))

		detailLine := ""
		switch entry.Status {
		case "OPTIMIZED":
			detailLine = styles.TxtBlue.Render(fmt.Sprintf("    %s ──> %s  (SAVED %s)  :: OPTIMIZED ::",
				formatBytesHuman(entry.Original),
				formatBytesHuman(entry.Optimized),
				formatBytesHuman(entry.Saved),
			))
		case "ERROR":
			detailLine = styles.TxtRed.Render("    ERROR: " + entry.Message)
		case "PROCESSING":
			detailLine = styles.TxtOrange.Render("    PROCESSING...")
		default:
			detailLine = styles.TxtGray.Render("    PENDING")
		}

		lines = append(lines, nameLine, detailLine, "")
	}

	total := len(m.compressionEntries)
	progress := 0
	if total > 0 {
		progress = int((float64(m.compressionProcessed) / float64(total)) * 100)
		if progress > 100 {
			progress = 100
		}
	}

	lines = append(lines,
		styles.TxtOrange.Render(strings.Repeat("-", 55)),
		styles.TxtWhite.Render(fmt.Sprintf("TOTAL PROGRESS: %s", renderProgressBar(progress, 20))),
		styles.TxtWhite.Render(fmt.Sprintf("SPACE RECLAIMED: %s", formatBytesHuman(m.compressionSaved))),
	)

	return strings.Join(lines, "\n")
}

func (m model) renderOverwatchView() string {
	lines := []string{
		"",
		styles.TxtOrange.Render("  // OVERWATCH // SYSTEM TELEMETRY"),
		"",
	}

	// Host/Uptime/Battery line
	uptimeStr := formatDuration(m.overwatchUptime)
	batteryStr := "N/A"
	batteryPercent := 0
	if m.overwatchBattery >= 0 {
		batteryPercent = m.overwatchBattery
		batteryStr = fmt.Sprintf("%d%%", batteryPercent)
	}
	hostLine := fmt.Sprintf("  HOST: %s   UPTIME: %s   BATTERY: %s", m.overwatchHostname, uptimeStr, batteryStr)
	if m.overwatchBattery >= 0 {
		batteryBar := renderProgressBar(batteryPercent, 20)
		hostLine = fmt.Sprintf("  HOST: %s   UPTIME: %s   BATTERY: %s", m.overwatchHostname, uptimeStr, batteryBar)
	}
	lines = append(lines, styles.TxtWhite.Render(hostLine), "")

	// Overall CPU (average of all cores)
	cpuPercent := 0.0
	if len(m.overwatchCPU) > 0 {
		sum := 0.0
		for _, p := range m.overwatchCPU {
			sum += p
		}
		cpuPercent = sum / float64(len(m.overwatchCPU))
	}
	percent := int(cpuPercent)
	if percent > 100 {
		percent = 100
	}
	bar := renderProgressBar(percent, 20)
	color := styles.TxtGreen
	if percent >= 80 {
		color = styles.TxtRed
	} else if percent >= 50 {
		color = styles.TxtOrange
	}
	// Temperature - we'll show N/A for now (requires additional library on Windows)
	tempStr := "N/A"
	cpuLine := color.Render(fmt.Sprintf("  [ CPU ] %s %d%%  :: %s°C", bar, percent, tempStr))
	lines = append(lines, cpuLine)

	// RAM
	ramPercent := 0
	if m.overwatchRAMTotal > 0 {
		ramPercent = int((float64(m.overwatchRAMUsed) / float64(m.overwatchRAMTotal)) * 100)
	}
	ramBar := renderProgressBar(ramPercent, 20)
	ramColor := styles.TxtGreen
	if ramPercent >= 80 {
		ramColor = styles.TxtRed
	} else if ramPercent >= 50 {
		ramColor = styles.TxtOrange
	}
	ramUsedGB := float64(m.overwatchRAMUsed) / (1024 * 1024 * 1024)
	ramTotalGB := float64(m.overwatchRAMTotal) / (1024 * 1024 * 1024)
	ramLine := ramColor.Render(fmt.Sprintf("  [ RAM USAGE  ] %s %d%%  :: %.1fGB / %.1fGB", ramBar, ramPercent, ramUsedGB, ramTotalGB))
	lines = append(lines, "", ramLine)

	// Swap
	swapPercent := 0
	if m.overwatchSwapTotal > 0 {
		swapPercent = int((float64(m.overwatchSwapUsed) / float64(m.overwatchSwapTotal)) * 100)
	}
	swapBar := renderProgressBar(swapPercent, 20)
	swapColor := styles.TxtGreen
	if swapPercent >= 80 {
		swapColor = styles.TxtRed
	} else if swapPercent >= 50 {
		swapColor = styles.TxtOrange
	}
	swapLine := swapColor.Render(fmt.Sprintf("  [ SWAP MEM   ] %s %d%%", swapBar, swapPercent))
	lines = append(lines, swapLine)

	// Network (real-time rate in bytes per second, converted to MB/s)
	netRXMBps := float64(m.overwatchNetRX) / (1024 * 1024)
	netTXMBps := float64(m.overwatchNetTX) / (1024 * 1024)
	netLine := styles.TxtBlue.Render(fmt.Sprintf("  [ NET RX ] ▼ %.1f MB/s   [ NET TX ] ▲ %.1f MB/s", netRXMBps, netTXMBps))
	lines = append(lines, "", netLine)


	// Top processes
	lines = append(lines, "", styles.TxtOrange.Render("  // TOP PROCESSES (RESOURCE HOGS)"), styles.TxtWhite.Render("  PID    NAME           CPU%    MEM%    STATUS"), "")

	for i, proc := range m.overwatchProcesses {
		selected := ""
		if i == m.overwatchSelected {
			selected = "  ► "
		} else {
			selected = "    "
		}
		statusStr := "RUNNING"
		if proc.Status != "R" && proc.Status != "running" {
			statusStr = proc.Status
		}
		procLine := fmt.Sprintf("%s%-5d   %-15s   %5.1f%%   %5.1f%%   [%s]", selected, proc.PID, truncateString(proc.Name, 15), proc.CPU, proc.Memory, statusStr)
		lines = append(lines, styles.TxtWhite.Render(procLine))
	}

	// Footer
	lines = append(lines, "", styles.TxtGray.Render("  [R] REFRESH   [K] KILL PROCESS   [ESC] RETURN TO BASE"))

	return strings.Join(lines, "\n")
}

func (m model) renderViewer() string {
	// Calculate pane widths (25% left, 75% right)
	leftWidth := m.width / 4
	rightWidth := m.width - leftWidth - 3 // -3 for divider and padding

	// Set viewport dimensions
	m.codeViewport.Width = rightWidth - 4
	m.codeViewport.Height = m.height - 15 // Adjust for header and footer

	// Dimmed style for inactive pane
	dimmedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	// Render file explorer (left pane)
	explorerView := m.fileExplorer.View()
	if m.viewerActivePane == 0 {
		// Explorer is active - render normally
		explorerView = lipgloss.NewStyle().Width(leftWidth).Render(explorerView)
	} else {
		// Explorer is inactive - dim it
		explorerView = dimmedStyle.Width(leftWidth).Render(explorerView)
	}

	// Divider
	divider := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(" │ ")

	// Render code viewport (right pane)
	codeContent := m.codeViewport.View()
	if m.viewerActivePane == 1 {
		// Code viewport is active - render normally
		codeContent = lipgloss.NewStyle().Width(rightWidth).Render(codeContent)
	} else {
		// Code viewport is inactive - dim it
		codeContent = dimmedStyle.Width(rightWidth).Render(codeContent)
	}

	// Join panes horizontally
	splitView := lipgloss.JoinHorizontal(lipgloss.Top, explorerView, divider, codeContent)

	// Footer with cursor position and controls
	footer := fmt.Sprintf("  [TAB] SWITCH PANE   [ENTER] OPEN   Ln %d, Col %d", m.viewerLineNum, m.viewerColNum)
	if m.viewerCurrentFile != "" {
		footer = fmt.Sprintf("  [TAB] SWITCH PANE   [ENTER] OPEN   %s   Ln %d, Col %d", 
			m.viewerCurrentFile, m.viewerLineNum, m.viewerColNum)
	}
	footerStyle := styles.TxtGray.Render(footer)

	return lipgloss.JoinVertical(lipgloss.Left, splitView, "\n", footerStyle)
}

func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	return fmt.Sprintf("%dh %dm", hours, minutes)
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func renderProgressBar(percent int, width int) string {
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}
	filled := int((float64(percent) / 100.0) * float64(width))
	if filled > width {
		filled = width
	}
	return fmt.Sprintf("[%s%s] %3d%%", strings.Repeat("█", filled), strings.Repeat("░", width-filled), percent)
}

func formatBytesHuman(b int64) string {
	if b < 1024 {
		return fmt.Sprintf("%d B", b)
	}
	kb := float64(b) / 1024
	if kb < 1024 {
		return fmt.Sprintf("%.1f KB", kb)
	}
	mb := kb / 1024
	if mb < 1024 {
		return fmt.Sprintf("%.2f MB", mb)
	}
	gb := mb / 1024
	return fmt.Sprintf("%.2f GB", gb)
}

func (m model) View() string {

	// --- CUSTOM BOOT SEQUENCE ---
	if m.state == statusBoot {
		output := "\n\n"

		// 1. DEFINE LOG LINES FIRST
		// We need these to calculate the total loading time dynamically
		logLines := []string{
			"[OK] MOUNTING CHROME DRIVER (SELENIUM MODE)",
			"[OK] LOADING AUDIO DRIVER (.WAV)",
			"[OK] MEMORY INTEGRITY CHECK... 16GB",
			"[OK] MOUNTING FILESYSTEM... /websites",
			"[OK] INITIALIZING ROD DRIVER...",
			"[OK] ESTABLISHING SECURE CONNECTION...",
			"", // Spacer
			"> WAKE UP, NEO..",
		}

		// YOUR CONFIGURATION
		stepsPerLine := 50

		// 2. CALCULATE PROGRESS BAR MATH
		// The bar should finish when the last text line appears.
		// Calculation: (Number of Lines * Steps Per Line)
		totalLoadingSteps := float64(len(logLines) * stepsPerLine)

		// Calculate Ratio based on the CONTENT, not the total max steps
		progressRatio := float64(m.bootProgress) / totalLoadingSteps

		// Clamp the ratio (So it stays full during the pulsing phase)
		if progressRatio < 0 {
			progressRatio = 0
		}
		if progressRatio > 1.0 {
			progressRatio = 1.0
		}

		// Render the Bar
		barWidth := 25
		filledLen := int(progressRatio * float64(barWidth))
		barStr := strings.Repeat("█", filledLen)

		// 3. RENDER HEADER
		output += styles.TxtOrange.Render(" // NEURAL LINK BIOS v2.4") + "\n\n"
		output += styles.TxtOrange.Render(" INITIALIZING NEO-TOKYO KERNEL...") + "\n"
		output += styles.TxtOrange.Render(fmt.Sprintf(" [%-25s]", barStr)) + "\n\n"

		// 4. RENDER SCROLLING LOG LINES
		for i, line := range logLines {
			// Offset by 5 so the bar starts just a tiny bit before text
			threshold := (i * stepsPerLine) + 5

			if m.bootProgress >= threshold {
				if strings.Contains(line, "WAKE UP") {
					// PULSE EFFECT: Bright Orange <-> Dim Gray
					if m.blinkState {
						output += styles.TxtOrange.Render(" "+line) + "\n"
					} else {
						output += styles.TxtGray.Faint(true).Render(" "+line) + "\n"
					}
				} else {
					output += styles.TxtOrange.Render(" "+line) + "\n"
				}
			}
		}

		return output
	}

	header := m.renderHeader()
	contentStyle := lipgloss.NewStyle().PaddingLeft(3)

	if m.state == statusNeuralSync {
		header := m.renderHeader()
		contentStyle := lipgloss.NewStyle().PaddingLeft(3)
		return lipgloss.JoinVertical(lipgloss.Left, header, "\n", contentStyle.Render(m.renderNeuralSync()))
	}

	// --- BATCH RUN SCREEN (Mission Control) ---
	if m.state == statusBatchRun {
		total := len(m.batchQueue)
		current := m.batchProgress
		percent := 0
		if total > 0 {
			percent = int((float64(current) / float64(total)) * 100)
		}

		// PROGRESS BAR RENDERER
		barLen := 20
		filled := int((float64(percent) / 100.0) * float64(barLen))

		// SAFETY CLAMP for Batch Bar
		if filled < 0 {
			filled = 0
		}
		if filled > barLen {
			filled = barLen
		}

		progressBar := strings.Repeat("█", filled) + strings.Repeat("░", barLen-filled)

		logView := ""
		for _, l := range m.batchLog {
			logView += fmt.Sprintf("   %s\n", l)
		}

		currentJobName := "COMPLETE"
		if current < total {
			currentJobName = m.batchQueue[current].Ticker
		}

		spin := m.spinner.View()
		flux := strings.Repeat("█", (m.globalTick%10)+5) + strings.Repeat("▒", 5)
		activeIcon := "►"
		if !m.blinkState {
			activeIcon = " "
		}

		rawContent := fmt.Sprintf(`
   BATCH SEQUENCE: LAZY_DEGEN_MODE
   -------------------------------
%s
   [%s] PROCESSING: %s  %s
   -------------------------------
   STATUS: %s
   PROGRESS: [%s] %d%%
 
   DATA FLUX: %s
`, styles.TxtGray.Render(logView), activeIcon, styles.TxtBlue.Render(currentJobName), spin, styles.TxtOrange.Render(m.statusMsg), progressBar, percent, flux)

		return lipgloss.JoinVertical(lipgloss.Left, header, "\n", contentStyle.Render(rawContent))
	}

	// --- COMPRESSION VIEW (Scrollable) ---
	if m.state == statusCompress {
		footerText := fmt.Sprintf("   [PRESS ESC or ENTER TO RETURN]                       %s\n", m.renderMusicStatus())
		footer := styles.TxtWhite.Render(footerText)
		screenContent := m.viewport.View()
		return lipgloss.JoinVertical(
			lipgloss.Left,
			header,
			screenContent,
			footer,
		)
	}

	// --- VIEWER (ZEN MODE) ---
	if m.state == statusViewer {
		return lipgloss.JoinVertical(lipgloss.Left, header, "\n", m.renderViewer())
	}

	// --- OVERWATCH SYSTEM MONITOR ---
	// Viewport settings - adjust these values to change scrollable area size
	// Width: controls horizontal padding (smaller = more padding)
	// Height: controls visible lines (smaller = more scrolling needed)
	if m.state == statusOverwatch {
		m.viewport.Width = m.width - 6  // EDIT HERE: Adjust viewport width (default: m.width - 6)
		headerHeight := lipgloss.Height(header)
		footerHeight := 0 // Footer takes ~2 lines
		m.viewport.Height = m.height - headerHeight - footerHeight - 2 // EDIT HERE: Adjust viewport height (default: calculated)
		footerText := fmt.Sprintf("   [PRESS ESC to RETURN]                       %s\n", m.renderMusicStatus())
		footer := styles.TxtWhite.Render(footerText)
		screenContent := m.viewport.View()
		return lipgloss.JoinVertical(
			lipgloss.Left,
			header,
			screenContent,
			footer,
		)
	}

	// --- STANDARD LIST SCREENS ---
	if m.state == statusMenu || m.state == statusAction || m.state == statusMusic || m.state == statusPatchMenu || m.state == statusHostingerList || m.state == statusGenesisThemeSelect || m.state == statusGenesisPlatformSelect || m.state == statusPostGen || m.state == statusBatchMenu || m.state == statusBatchFileSelect {

		driverStatus := ""
		if m.driverConnected {
			indicator := "●"
			if !m.blinkState {
				indicator = "◌"
			}
			driverStatus = styles.TxtGreen.Render(fmt.Sprintf("%s DRIVER: %s", indicator, m.driverModeStr))
		} else {
			indicator := "■"
			if !m.blinkState {
				indicator = " "
			}
			driverStatus = styles.TxtRed.Render(fmt.Sprintf("%s DRIVER: OFFLINE", indicator))
		}

		musicStatus := m.renderMusicStatus()

		graph := ""
		for _, h := range m.waveHistory {
			graph += styles.Bars[h]
		}

		// --- DIAGNOSTICS PANEL ---
		var mStats runtime.MemStats
		runtime.ReadMemStats(&mStats)
		memVal := mStats.Alloc / 1024 / 1024 // MB
		uptime := time.Since(m.startTime).Round(time.Second)

		diagContent := fmt.Sprintf(`
%s
%s

%s
CPU: %d%%   RPM: %d
MEM: %d MB  PING: %dms
TIME: %s
%s

%s`,
			styles.TxtRed.Bold(true).Render("SYSTEM_DIAGNOSTICS"),
			styles.TxtBlue.Render(graph),

			styles.TxtOrange.Render("HARDWARE_MONITOR"),
			m.diagTemp, m.diagRPM,
			memVal, m.ping,
			uptime.String(),
			driverStatus,
			musicStatus,
		)

		rightPanel := lipgloss.NewStyle().
			PaddingLeft(2).
			Render(diagContent)

		listHeader := ""
		if m.state == statusPatchMenu {
			listHeader = styles.TxtOrange.Render(fmt.Sprintf(" INJECTING INTO: %s", strings.ToUpper(m.selectedTarget))) + "\n"
		}
		if m.state == statusBatchMenu {
			listHeader = styles.TxtOrange.Render(" BATCH OPERATIONS (FOLDER: /batch_jobs):") + "\n"
		}
		if m.state == statusBatchFileSelect {
			listHeader = styles.TxtOrange.Render(" SELECT TARGET JSON FILE:") + "\n"
		}
		if m.state == statusHostingerList {
			listHeader = styles.TxtOrange.Render(" SELECT HOSTINGER NODE:") + "\n"
		}
		if m.state == statusGenesisThemeSelect {
			listHeader = styles.TxtOrange.Render(" SELECT VISUAL STYLE:") + "\n"
		}
		if m.state == statusGenesisPlatformSelect {
			listHeader = styles.TxtOrange.Render(" SELECT BLOCKCHAIN NETWORK:") + "\n"
		}
		if m.state == statusPostGen {
			listHeader = styles.TxtGreen.Render(fmt.Sprintf("  SUCCESS: %s", m.lastGenFolder)) + "\n"
		}

		body := lipgloss.JoinHorizontal(lipgloss.Top,
			lipgloss.NewStyle().Width(60).Render(listHeader+m.list.View()),
			rightPanel)

		finalHeader := header

		// STICKY FOOTER LOGIC
		if m.state == statusBatchMenu || m.state == statusBatchFileSelect {
			footerText := fmt.Sprintf("   [PRESS ESC TO RETURN]                                  %s", m.renderMusicStatus())
			footer := styles.TxtWhite.Render(footerText)

			headerH := lipgloss.Height(finalHeader)
			contentH := lipgloss.Height(body) + 2
			footerH := lipgloss.Height(footer)

			gap := m.height - headerH - contentH - footerH - 3
			gapFill := ""
			if gap > 0 {
				gapFill = strings.Repeat("\n", gap)
			}

			return lipgloss.JoinVertical(lipgloss.Left, finalHeader, "\n\n", contentStyle.Render(body), gapFill, footer)
		}

		return lipgloss.JoinVertical(lipgloss.Left, finalHeader, "\n\n", contentStyle.Render(body))
	}

	// --- AUDIT REPORT VIEW ---
	if m.state == statusAuditReport {
		footerText := fmt.Sprintf("   [PRESS ESC or ENTER TO RETURN]                       %s\n", m.renderMusicStatus())
		footer := styles.TxtWhite.Render(footerText)
		screenContent := m.viewport.View()
		return lipgloss.JoinVertical(
			lipgloss.Left,
			header,
			screenContent,
			footer,
		)
	}

	// MANUAL SCREEN
	if m.state == statusManual {
		return lipgloss.JoinVertical(lipgloss.Left, header, "\n", m.viewport.View())
	}

	// VIEW: GENESIS/PATCH INPUT SCREENS
	if m.state == statusGenesisInput || m.state == statusGenesisSocials || m.state == statusGenesisCustomTheme {

		labels := []string{"TICKER SYMBOL", "THEME/VIBE", "BLOCKCHAIN PLATFORM", "TWITTER LINK", "TELEGRAM LINK", "DISCORD", "DEX SCREENER", "TOOLS/UTILITY", "IMAGE ASSETS (URL)", "EXTRA CONTEXT"}

		history := ""
		for i, input := range m.inputs {
			if i < len(labels) {
				history += styles.TxtGray.Faint(true).Render(fmt.Sprintf("  ▵ %s: %s\n", labels[i], input))
			}
		}

		lbl := "INPUT VALUE"
		if m.state == statusGenesisCustomTheme {
			lbl = "CUSTOM VISUAL STYLE"
		} else if m.opMode == "GENESIS" && m.inputStep < len(labels) {
			lbl = labels[m.inputStep]
		} else if m.opMode == "PATCH" {
			lbl = fmt.Sprintf("NEW VALUE FOR %s", m.patchField)
		}

		rawContent := fmt.Sprintf(`
SUBJECT INITIALIZATION PROTOCOL
-------------------------------
%s
ENTER PARAMETER (%s):
%s
`, history, styles.TxtOrange.Render(lbl), m.textInput.View())

		contentStyle := lipgloss.NewStyle().PaddingLeft(4)
		screenContent := contentStyle.Render(rawContent)

		footerText := fmt.Sprintf("   [ESC] ABORT      [ENTER] INITIATE SEQUENCE                  %s", m.renderMusicStatus())
		footer := styles.TxtWhite.Render(footerText)

		headerH := lipgloss.Height(header)
		contentH := lipgloss.Height(screenContent)
		footerH := lipgloss.Height(footer)
		gap := m.height - headerH - contentH - footerH - 3

		gapFill := ""
		if gap > 0 {
			gapFill = strings.Repeat("\n", gap)
		}

		return lipgloss.JoinVertical(lipgloss.Left, "\n\n\n", header, screenContent, gapFill, footer, "\n")
	}

	// VIEW: PROCESSING SCREEN
	if m.state == statusProcessing {
		flux := strings.Repeat("█", rand.Intn(15)) + strings.Repeat("▒", 5)
		rawContent := fmt.Sprintf(`
		
⚠  CRITICAL ALERT  ⚠  //  SYNCHRONIZING...

%s %s

DATA FLUX:
%s  [OVERFLOW]

AWAITING NEURAL LINK RESPONSE...
`, m.spinner.View(), m.statusMsg, flux)

		screenContent := contentStyle.Render(styles.TxtOrange.Render(rawContent))
		headerH := lipgloss.Height(header)
		contentH := lipgloss.Height(screenContent)
		gap := m.height - headerH - contentH - 2
		gapFill := ""
		if gap > 0 {
			gapFill = strings.Repeat("\n", gap)
		}

		return lipgloss.JoinVertical(lipgloss.Left, header, screenContent, gapFill)
	}

	if m.state == statusDone {
		return lipgloss.JoinVertical(lipgloss.Left, header, contentStyle.Render(styles.TxtBlue.Render(fmt.Sprintf("\n  MISSION REPORT:\n  %s\n\n  [PRESS ENTER TO RETURN]", m.statusMsg))))
	}

	return ""
}
