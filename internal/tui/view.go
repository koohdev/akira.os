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

func (m model) renderCompressionView() string {
	lines := []string{
		styles.TxtOrange.Render("// GIGA-COMPRESS // ASSET OPTIMIZER v1.0"),
		styles.TxtGray.Render(fmt.Sprintf("Target: ./websites/%s/images/", m.compressionFolder)),
		"",
	}

	for _, entry := range m.compressionEntries {
		statusIcon := "[ ]"
		color := styles.TxtGray
		percent := int(entry.Progress * 100)

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
			percent = 0
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

	contentStyle := lipgloss.NewStyle().PaddingLeft(3)
	return contentStyle.Render(strings.Join(lines, "\n"))
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
	return fmt.Sprintf("[%s%s] %3d%%", strings.Repeat("=", filled), strings.Repeat(".", width-filled), percent)
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

	if m.state == statusBoot {
		bar := strings.Repeat("█", m.bootProgress)
		return styles.TxtOrange.Render(fmt.Sprintf("\n\n  INITIALIZING NEO-TOKYO KERNEL...\n  [%-25s]\n\n  > MOUNTING CHROME DRIVER (SELENIUM MODE)...\n  > LOADING AUDIO DRIVER (.WAV)...", bar))
	}

	header := m.renderHeader()
	contentStyle := lipgloss.NewStyle().PaddingLeft(3)

	// --- BATCH RUN SCREEN (Mission Control) ---
	if m.state == statusBatchRun {
		total := len(m.batchQueue)
		current := m.batchProgress
		percent := 0
		if total > 0 {
			percent = int((float64(current) / float64(total)) * 100)
		}
		barLen := 20
		filled := int((float64(percent) / 100.0) * float64(barLen))
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

	// --- COMPRESSION VIEW ---
	if m.state == statusCompress {
		return lipgloss.JoinVertical(lipgloss.Left, header, "\n", m.renderCompressionView())
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

		// --- NEW DIAGNOSTICS PANEL (CLEAN / NO BOX) ---
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

			// Removed lines
			musicStatus,
		)

		// Styled Right Panel (Just padding, no border)
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
