package tui

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetWidth(msg.Width / 2)
		m.list.SetHeight(14)
		headerHeight := lipgloss.Height(m.renderHeader())
		verticalMargins := 3
		vpHeight := msg.Height - headerHeight - verticalMargins
		if vpHeight < 1 {
			vpHeight = 1
		}
		m.viewport.Width = msg.Width
		m.viewport.Height = vpHeight

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			killPlayer(m.activePlayer)
			return m, tea.Quit
		}

		if msg.String() == "esc" && m.state != statusBoot {
			if m.state == statusBatchMenu || m.state == statusBatchFileSelect {
				m.state = statusMenu
				m.updateMainMenu()
				return m, nil
			}
			m.state = statusMenu
			m.updateMainMenu()
			return m, nil
		}

		if m.state == statusAuditReport || m.state == statusManual {
			if msg.String() == "enter" || msg.String() == "esc" {
				m.state = statusMenu
				m.updateMainMenu()
				return m, nil
			}
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd
		}

		// --- MAIN MENU ---
		if m.state == statusMenu {
			if msg.String() == "enter" {
				i, ok := m.list.SelectedItem().(item)
				if ok {
					switch i.id {
					case "GENESIS":
						m.state = statusGenesisInput
						m.opMode = "GENESIS"
						m.inputStep = 0
						m.inputs = []string{}
						m.textInput.Reset()
						m.textInput.Focus()
					case "BATCH":
						m.state = statusBatchMenu
						m.list.SetItems([]list.Item{
							item{title: "EXECUTE BATCH SEQUENCE", desc: "Select a JSON file to run", id: "RUN_BATCH"},
							item{title: "CREATE NEW TEMPLATE", desc: "Generate blank jobs(x).json", id: "NEW_BATCH"},
							item{title: "OPEN BATCH FOLDER", desc: "Open in Explorer", id: "OPEN_FOLDER"},
						})
						m.list.ResetSelected()
					case "PATCH", "PREVIEW", "AUDIT", "COMPRESS":
						m.opMode = i.id
						m.state = statusAction
						return m, loadFolders()
					case "HOSTINGER":
						m.state = statusProcessing
						m.statusMsg = "SCANNING HOSTINGER SATELLITES..."
						return m, runHostingerScan()
					case "TOGGLE_DEV":
						m.devMode = !m.devMode
						m.state = statusProcessing
						modeText := "SILENT MODE"
						if m.devMode {
							modeText = "DEBUG MODE (VISIBLE)"
						}
						m.statusMsg = fmt.Sprintf("REBOOTING DRIVER: %s...", modeText)
						return m, rebootDriver(m.devMode)
					case "MUSIC":
						m.state = statusMusic
						return m, loadMusic()
					case "MANUAL":
						m.state = statusManual
						m.setManualText()
						return m, nil
					case "EXIT":
						killPlayer(m.activePlayer)
						return m, tea.Quit
					}
				}
			}
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}

		// --- BATCH MENU LOGIC ---
		if m.state == statusBatchMenu {
			if msg.String() == "enter" {
				i, ok := m.list.SelectedItem().(item)
				if ok {
					switch i.id {
					case "NEW_BATCH":
						return m, createBatchTemplate()
					case "RUN_BATCH":
						m.state = statusBatchFileSelect
						return m, loadBatchFilesList()
					case "OPEN_FOLDER":
						openBatchFolder()
						return m, nil
					}
				}
			}
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}

		// --- BATCH FILE SELECTION ---
		if m.state == statusBatchFileSelect {
			if msg.String() == "enter" {
				i, ok := m.list.SelectedItem().(item)
				if ok && i.id != "NONE" {
					m.batchFile = i.title
					m.state = statusBatchRun
					m.batchLog = []string{}
					m.batchProgress = 0
					m.statusMsg = fmt.Sprintf("INITIALIZING SEQUENCE: %s...", m.batchFile)
					return m, loadBatchFile(m.batchFile)
				}
			}
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}

		// --- INPUT HANDLERS ---
		if m.state == statusGenesisInput {
			if msg.String() == "enter" {
				val := m.textInput.Value()
				if m.opMode == "PATCH" {
					executePatch(m.selectedTarget, m.patchField, val)
					m.state = statusDone
					m.statusMsg = fmt.Sprintf("PATCH APPLIED: %s -> %s", m.patchField, m.selectedTarget)
					return m, nil
				}
				if val == "" {
					val = "None"
				}
				m.inputs = append(m.inputs, val)
				m.inputStep++
				m.state = statusGenesisThemeSelect
				m.loadThemeOptions()
				return m, nil
			}
			m.textInput, cmd = m.textInput.Update(msg)
			return m, cmd
		}

		if m.state == statusGenesisThemeSelect {
			if msg.String() == "enter" {
				i, ok := m.list.SelectedItem().(item)
				if ok {
					if i.id == "CUSTOM" {
						m.state = statusGenesisCustomTheme
						m.textInput.Placeholder = "ENTER CUSTOM VISUAL STYLE..."
						m.textInput.Reset()
						return m, nil
					}
					m.inputs = append(m.inputs, i.title)
					m.inputStep++
					m.state = statusGenesisPlatformSelect
					m.loadPlatformOptions()
					return m, nil
				}
			}
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}

		if m.state == statusGenesisCustomTheme {
			if msg.String() == "enter" {
				val := m.textInput.Value()
				if val == "" {
					val = "Custom"
				}
				m.inputs = append(m.inputs, val)
				m.inputStep++
				m.state = statusGenesisPlatformSelect
				m.loadPlatformOptions()
				return m, nil
			}
			m.textInput, cmd = m.textInput.Update(msg)
			return m, cmd
		}

		if m.state == statusGenesisPlatformSelect {
			if msg.String() == "enter" {
				i, ok := m.list.SelectedItem().(item)
				if ok {
					m.inputs = append(m.inputs, i.title)
					m.inputStep++
					m.state = statusGenesisSocials
					m.textInput.Placeholder = "X / Twitter Link"
					m.textInput.Reset()
					return m, nil
				}
			}
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}

		if m.state == statusGenesisSocials {
			labels := []string{"TICKER SYMBOL", "THEME/VIBE", "BLOCKCHAIN PLATFORM", "X / Twitter Link", "Telegram Link", "Discord Link", "Dex Screener Link", "Tools/Utility Link", "Image Assets (URL)", "Extra Context"}
			if msg.String() == "enter" {
				val := m.textInput.Value()
				if val == "" {
					val = "None"
				}
				m.inputs = append(m.inputs, val)
				m.inputStep++
				m.textInput.Reset()
				nextStep := m.inputStep
				if nextStep < len(labels) {
					m.textInput.Placeholder = labels[nextStep]
				}
				if m.inputStep >= 10 {
					m.state = statusProcessing
					m.statusMsg = "TRANSMITTING DATA TO GEMINI..."
					return m, runGenesis(m.inputs)
				}
			}
			m.textInput, cmd = m.textInput.Update(msg)
			return m, cmd
		}

		if m.state == statusPostGen {
			if msg.String() == "enter" {
				i, ok := m.list.SelectedItem().(item)
				if ok {
					if i.id == "VIEW" {
						openBrowser(m.lastGenFolder)
					} else if i.id == "FOLDER" {
						openExplorer(m.lastGenFolder)
					} else if i.id == "MENU" {
						m.state = statusMenu
						m.updateMainMenu()
					}
				}
			}
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}

		// --- ACTION / SUBMENUS ---
		if m.state == statusAction {
			if msg.String() == "enter" {
				i, ok := m.list.SelectedItem().(item)
				if ok {
					if m.opMode == "PREVIEW" {
						openBrowser(i.title)
						m.state = statusDone
						m.statusMsg = "VISUALS ONLINE: " + i.title
					} else if m.opMode == "PATCH" {
						m.selectedTarget = i.title
						m.state = statusPatchMenu
						m.loadPatchOptions()
						return m, nil
					} else if m.opMode == "AUDIT" {
						m.selectedTarget = i.title
						m.state = statusProcessing
						m.statusMsg = fmt.Sprintf("AUDITING EXTERNAL LINKS FOR: %s", i.title)
						return m, runAudit(i.title)
					} else if m.opMode == "COMPRESS" {
						m.selectedTarget = i.title
						m.state = statusProcessing
						m.statusMsg = fmt.Sprintf("COMPRESSING IMAGES FOR: %s", i.title)
						return m, runImageCompress(i.title)
					}
				}
			}
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}

		if m.state == statusPatchMenu {
			if msg.String() == "enter" {
				i, ok := m.list.SelectedItem().(item)
				if ok {
					m.patchField = i.id
					m.state = statusGenesisInput
					m.textInput.Placeholder = "Enter new value..."
					m.textInput.Reset()
					m.textInput.Focus()
					return m, nil
				}
			}
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}

		if m.state == statusHostingerList {
			if msg.String() == "enter" {
				i, ok := m.list.SelectedItem().(item)
				if ok {
					m.state = statusProcessing
					m.statusMsg = "WARPING TO DASHBOARD..."
					return m, runHostingerNav(i.title)
				}
			}
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}

		if m.state == statusMusic {
			if msg.String() == "enter" {
				i, ok := m.list.SelectedItem().(item)
				if ok {
					killPlayer(m.activePlayer)
					if i.id != "STOP" {
						m.activePlayer = playAudio(i.title)
						m.musicPlaying = true
						m.currentSong = i.title
					} else {
						m.musicPlaying = false
					}
				}
			}
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}

		if m.state == statusDone {
			if msg.String() == "enter" {
				m.state = statusMenu
				m.updateMainMenu()
			}
		}

	case driverReadyMsg:
		m.driverModeStr = msg.mode
		if m.state == statusBoot {
			m.bootProgress = 30
			m.state = statusMenu
			m.updateMainMenu()
		} else {
			m.state = statusDone
			m.statusMsg = fmt.Sprintf("DRIVER REBOOTED: %s", msg.mode)
		}

	case tickMsg:
		m.globalTick++
		if m.state == statusBoot {
			if m.bootProgress < 30 {
				m.bootProgress++
			}
		}
		if m.globalTick%5 == 0 {
			if checkDriverPort() {
				m.driverConnected = true
			} else {
				m.driverConnected = false
			}
		}
		if m.globalTick%10 == 0 {
			m.blinkState = !m.blinkState
			m.waveHistory = append(m.waveHistory[1:], rand.Intn(8))
			m.ping = 30 + rand.Intn(40)

			if rand.Intn(100) > 80 {
				m.diagRPM = 8000 + rand.Intn(6000)
				m.diagTemp = 80 + rand.Intn(40)
			}
		}

		// --- ANIMATE AUDIT REPORT ICONS ---
		if m.state == statusAuditReport {
			iconOk := "●"
			iconFail := "■"
			iconWarn := "▲"
			if !m.blinkState {
				iconOk = "◌"
				iconFail = " "
				iconWarn = " "
			}
			report := "\n\n   AUDIT REPORT: EXTERNAL DEPENDENCY HEALTH\n   ----------------------------------------\n"
			for _, result := range m.auditResults {
				line := strings.ReplaceAll(result, "::OK::", iconOk)
				line = strings.ReplaceAll(line, "::FAIL::", iconFail)
				line = strings.ReplaceAll(line, "::WARN::", iconWarn)
				report += "   " + line + "\n"
			}
			m.viewport.SetContent(report)
		}

		if m.state == statusCompress {
			for i := range m.compressionEntries {
				if m.compressionEntries[i].Status == "PROCESSING" {
					if m.blinkState {
						m.compressionEntries[i].Progress = 0.65
					} else {
						m.compressionEntries[i].Progress = 0.35
					}
				}
			}
		}
		return m, tea.Tick(time.Millisecond*50, func(t time.Time) tea.Msg { return tickMsg(t) })

	// --- MSG HANDLERS ---
	case folderMsg:
		var items []list.Item
		if m.state == statusMusic {
			items = append(items, item{title: "SILENCE (STOP)", desc: "Stop Audio Signal", id: "STOP"})
		}
		sort.Strings(msg)
		for _, f := range msg {
			items = append(items, item{title: f, desc: "DATA", id: f})
		}
		if len(items) == 0 {
			items = append(items, item{title: "NO DATA FOUND", desc: "Create a project first", id: "NONE"})
		}
		m.list.SetItems(items)

	case hostingerMsg:
		m.state = statusHostingerList
		var items []list.Item
		for _, s := range msg {
			items = append(items, item{title: s, desc: "Active Site", id: s})
		}
		m.list.SetItems(items)

	case auditMsg:
		m.auditResults = msg
		m.state = statusAuditReport
		m.list.SetItems(make([]list.Item, 0))

	// --- BATCH LOADING ---
	case batchLoadMsg:
		m.batchQueue = msg
		if len(m.batchQueue) > 0 {
			job := m.batchQueue[0]
			m.statusMsg = fmt.Sprintf("PROCESSING JOB 1/%d: %s", len(m.batchQueue), job.Ticker)
			inputs := []string{job.Ticker, job.Theme, job.Platform, job.Twitter, job.Telegram, job.Discord, job.Dex, "None", job.Assets, job.Context}
			return m, runGenesis(inputs)
		}

	case browserMsg:
		// --- BATCH PROCESSING LOGIC ---
		if m.state == statusBatchRun {
			// SAFETY CHECK
			if len(m.batchQueue) == 0 {
				m.state = statusDone
				m.statusMsg = msg.msg
				return m, nil
			}

			// 1. Log Result
			statusIcon := "[✓]"
			if strings.Contains(msg.msg, "ERR") {
				statusIcon = "[X]"
			}
			logEntry := fmt.Sprintf("%s %s :: %s", statusIcon, m.batchQueue[m.batchProgress].Ticker, msg.msg)
			m.batchLog = append(m.batchLog, logEntry)

			// 2. Increment
			m.batchProgress++

			// 3. Check Sequence
			if m.batchProgress < len(m.batchQueue) {
				// START NEXT JOB
				job := m.batchQueue[m.batchProgress]
				m.statusMsg = fmt.Sprintf("PROCESSING JOB %d/%d: %s", m.batchProgress+1, len(m.batchQueue), job.Ticker)

				time.Sleep(5 * time.Second)

				inputs := []string{job.Ticker, job.Theme, job.Platform, job.Twitter, job.Telegram, job.Discord, job.Dex, "None", job.Assets, job.Context}
				return m, runGenesis(inputs)
			} else {
				// ALL DONE
				m.state = statusDone
				m.statusMsg = "BATCH SEQUENCE COMPLETE. CHECK LOGS."
				return m, nil
			}
		}

		if strings.Contains(msg.msg, "SUBJECT CONSTRUCTED") {
			m.state = statusPostGen
			m.lastGenFolder = msg.folder
			options := []list.Item{
				item{title: "VISUAL CONFIRMATION", desc: "Open Site in Browser", id: "VIEW"},
				item{title: "ACCESS DATA", desc: "Open Windows Folder", id: "FOLDER"},
				item{title: "NEW OPERATION", desc: "Return to Main Menu", id: "MENU"},
			}
			m.list.SetItems(options)
			m.list.ResetSelected()
		} else if strings.Contains(msg.msg, "MISSION REPORT") {
			m.state = statusDone
			m.statusMsg = msg.msg
		} else {
			m.state = statusDone
			m.statusMsg = msg.msg
		}

	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case compressionInitMsg:
		if len(msg.files) == 0 {
			m.state = statusDone
			m.statusMsg = fmt.Sprintf("NO PNG/JPEG ASSETS FOUND IN %s/images", msg.folder)
			return m, nil
		}
		m.compressionFolder = msg.folder
		m.compressionFiles = msg.files
		m.compressionEntries = make([]compressionEntry, len(msg.files))
		for i, file := range msg.files {
			m.compressionEntries[i] = compressionEntry{Name: file.Name, Status: "PENDING", Progress: 0}
		}
		m.compressionEntries[0].Status = "PROCESSING"
		m.compressionEntries[0].Progress = 0.35
		m.compressionProcessed = 0
		m.compressionSaved = 0
		m.state = statusCompress
		return m, compressFileCmd(msg.files[0], 0)

	case compressionProgressMsg:
		if msg.index < len(m.compressionEntries) {
			entry := &m.compressionEntries[msg.index]
			entry.Progress = 1
			if msg.err != nil {
				entry.Status = "ERROR"
				entry.Message = msg.err.Error()
			} else {
				entry.Status = "OPTIMIZED"
				entry.Original = msg.original
				entry.Optimized = msg.optimized
				entry.Saved = msg.original - msg.optimized
				m.compressionProcessed++
				m.compressionSaved += entry.Saved
			}
		}

		next := msg.index + 1
		if next < len(m.compressionEntries) && next < len(m.compressionFiles) {
			if m.compressionEntries[next].Status == "PENDING" {
				m.compressionEntries[next].Status = "PROCESSING"
				m.compressionEntries[next].Progress = 0.35
			}
			return m, compressFileCmd(m.compressionFiles[next], next)
		}

		// All done
		totalMB := float64(m.compressionSaved) / (1024 * 1024)
		m.state = statusDone
		m.statusMsg = fmt.Sprintf("GIGA-COMPRESS COMPLETE: %d FILES, %.2f MB RECLAIMED", m.compressionProcessed, totalMB)
		return m, nil
	}
	return m, cmd
}
