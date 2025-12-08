package tui

import (
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shirou/gopsutil/v3/cpu"
)

// --- DEV LIFE OF QUALITY: BOOT CONFIGURATION ---
const (
	BootTickSpeed = time.Millisecond * 5 // Speed of one "tick" (Lower = Faster)
	BootMaxSteps  = 1000                 // Total length of boot sequence (Increase if text gets cut off)
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetWidth(msg.Width / 2)
		m.list.SetHeight(14)
		// Update viewer viewport dimensions
		if m.state == statusViewer {
			leftWidth := m.width / 4
			rightWidth := m.width - leftWidth - 3
			m.codeViewport.Width = rightWidth - 4
			m.codeViewport.Height = m.height - 15
		}
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

		if m.state == statusAuditReport || m.state == statusManual || m.state == statusCompress {
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
					case "SYNC":
						m.state = statusNeuralSync
						m.typingPhase = 1
						m.typingTargets = getRandomizedTargets()
						m.typingTarget = m.typingTargets[0]
						m.typingInput = ""
						m.typingStartTime = time.Now()
						m.typingTimerDuration = 30 * time.Second
						m.typingCurrentWPM = 0
						m.typingAccuracy = 0
						m.typingPhaseWPM = make([]int, 5)
						m.typingPhaseAccuracy = make([]int, 5)
						m.textInput.Reset()
						m.textInput.Focus()
						return m, nil
					case "OVERWATCH":
						m.state = statusOverwatch
						m.overwatchSelected = 0
						return m, runOverwatchScan()
					case "VIEWER":
						m.state = statusViewer
						m.viewerActivePane = 0
						m.fileExplorer.CurrentDirectory, _ = os.Getwd()
						return m, m.fileExplorer.Init()
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

		// --- NEURAL SYNC TYPING GAME ---
		if m.state == statusNeuralSync {
			if msg.String() == "esc" {
				m.state = statusMenu
				m.updateMainMenu()
				return m, nil
			}
			// Update text input
			m.textInput, cmd = m.textInput.Update(msg)
			m.typingInput = m.textInput.Value()

			// Calculate WPM and accuracy
			elapsed := time.Since(m.typingStartTime)
			if elapsed.Seconds() > 0 {
				charsTyped := len(m.typingInput)
				minutes := elapsed.Minutes()
				if minutes > 0 {
					m.typingCurrentWPM = int(float64(charsTyped) / 5.0 / minutes)
				}
			}

			// Calculate accuracy
			correct := 0
			for i := 0; i < len(m.typingInput) && i < len(m.typingTarget); i++ {
				if m.typingInput[i] == m.typingTarget[i] {
					correct++
				}
			}
			if len(m.typingInput) > 0 {
				m.typingAccuracy = int((float64(correct) / float64(len(m.typingInput))) * 100)
			}

			// Check if completed
			if m.typingInput == m.typingTarget {
				// Store results for this phase
				phaseIndex := m.typingPhase - 1
				if phaseIndex >= 0 && phaseIndex < len(m.typingPhaseWPM) {
					m.typingPhaseWPM[phaseIndex] = m.typingCurrentWPM
					m.typingPhaseAccuracy[phaseIndex] = m.typingAccuracy
				}

				// Check if this is the last phase
				if m.typingPhase >= m.typingTotalPhases {
					// Calculate average WPM and accuracy
					avgWPM := 0
					avgAccuracy := 0
					for i := 0; i < len(m.typingPhaseWPM); i++ {
						avgWPM += m.typingPhaseWPM[i]
						avgAccuracy += m.typingPhaseAccuracy[i]
					}
					if len(m.typingPhaseWPM) > 0 {
						avgWPM /= len(m.typingPhaseWPM)
						avgAccuracy /= len(m.typingPhaseAccuracy)
					}
					m.typingLastWPM = avgWPM
					m.state = statusDone
					m.statusMsg = fmt.Sprintf("NEURAL SYNC COMPLETE: %d WPM, %d%% ACCURACY", avgWPM, avgAccuracy)
					return m, nil
				} else {
					// Move to next phase
					m.typingPhase++
					m.typingTarget = m.typingTargets[m.typingPhase-1]
					m.typingInput = ""
					m.typingStartTime = time.Now()
					m.typingCurrentWPM = 0
					m.typingAccuracy = 0
					m.textInput.Reset()
					m.textInput.Focus()
					return m, nil
				}
			}

			// Check timeout
			if time.Since(m.typingStartTime) > m.typingTimerDuration {
				m.state = statusDone
				m.statusMsg = fmt.Sprintf("NEURAL SYNC TIMEOUT ON SEQUENCE %d. TRY AGAIN.", m.typingPhase)
				return m, nil
			}

			return m, cmd
		}

		// --- OVERWATCH SYSTEM MONITOR ---
		if m.state == statusOverwatch {
			if msg.String() == "esc" {
				m.state = statusMenu
				m.updateMainMenu()
				return m, nil
			}
			if msg.String() == "r" || msg.String() == "R" {
				// Refresh scan
				return m, runOverwatchScan()
			}
			if msg.String() == "k" || msg.String() == "K" {
				// Kill selected process (immediately, no confirmation)
				if m.overwatchSelected >= 0 && m.overwatchSelected < len(m.overwatchProcesses) {
					proc := m.overwatchProcesses[m.overwatchSelected]
					return m, killProcessCmd(proc.PID)
				}
			}
			// Arrow keys to select process (but also allow viewport scrolling)
			if msg.String() == "up" {
				// Check if we're in the process list area - if so, select process
				// Otherwise, scroll viewport
				if m.overwatchSelected > 0 {
					m.overwatchSelected--
					m.viewport.SetContent(m.renderOverwatchView())
					return m, nil
				}
			}
			if msg.String() == "down" {
				if m.overwatchSelected < len(m.overwatchProcesses)-1 {
					m.overwatchSelected++
					m.viewport.SetContent(m.renderOverwatchView())
					return m, nil
				}
			}
			// Allow viewport to handle scrolling (page up/down, mouse wheel)
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd
		}

		// --- VIEWER (ZEN MODE) ---
		if m.state == statusViewer {
			if msg.String() == "esc" {
				m.state = statusMenu
				m.updateMainMenu()
				return m, nil
			}
			if msg.String() == "tab" {
				// Toggle active pane
				m.viewerActivePane = 1 - m.viewerActivePane
				return m, nil
			}

			// Handle filepicker messages
			var fpCmd tea.Cmd
			m.fileExplorer, fpCmd = m.fileExplorer.Update(msg)

			// Check if a file was selected
			didSelect, filePath := m.fileExplorer.DidSelectFile(msg)
			if didSelect && filePath != "" {
				// Load the selected file
				return m, tea.Batch(loadFileCmd(filePath), fpCmd)
			}

			// Handle viewport messages
			var vpCmd tea.Cmd
			m.codeViewport, vpCmd = m.codeViewport.Update(msg)

			// Update cursor position for code viewport
			if m.viewerActivePane == 1 {
				// Calculate line/col from viewport position
				// This is approximate - viewport doesn't expose exact cursor position
				m.viewerLineNum = m.codeViewport.YOffset + 1
				m.viewerColNum = 1
			}

			return m, tea.Batch(fpCmd, vpCmd)
		}

	case fileLoadedMsg:
		if m.state == statusViewer {
			if msg.err == nil {
				m.viewerCurrentFile = msg.path
				m.codeViewport.SetContent(msg.content)
				m.codeViewport.YOffset = 0
				// Switch focus to code viewport
				m.viewerActivePane = 1
			}
		}
		return m, nil

	case driverReadyMsg:
		m.driverModeStr = msg.mode
		if m.state == statusBoot {
			// Ensure we wait for the full animation even if driver loads fast
			m.bootProgress = BootMaxSteps
			m.state = statusMenu
			m.updateMainMenu()
		} else {
			m.state = statusDone
			m.statusMsg = fmt.Sprintf("DRIVER REBOOTED: %s", msg.mode)
		}

	case tickMsg:
		m.globalTick++
		if m.state == statusBoot {
			// Use the Config variable here
			if m.bootProgress < BootMaxSteps {
				m.bootProgress++
			}
		}
		if m.globalTick%100 == 0 {
			if checkDriverPort() {
				m.driverConnected = true
			} else {
				m.driverConnected = false
			}
		}
		if m.globalTick%100 == 0 {
			m.blinkState = !m.blinkState
			m.waveHistory = append(m.waveHistory[1:], rand.Intn(8))
			m.ping = 30 + rand.Intn(40)

			// Update real hardware monitor data
			cpuPercents, err := cpu.Percent(100*time.Millisecond, false)
			if err == nil && len(cpuPercents) > 0 {
				m.diagTemp = int(cpuPercents[0]) // CPU percentage (0-100) - using diagTemp field for CPU display
			} else {
				m.diagTemp = 0
			}
			// RPM - not used, set to 0
			m.diagRPM = 0
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
			report := fmt.Sprintf("\n\n   AUDIT REPORT: EXTERNAL DEPENDENCY HEALTH | FOLDER: %s | ONLINE: %d | FAILED: %d\n   ----------------------------------------\n", m.auditFolder, m.auditOnline, m.auditFailed)
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
			// Update viewport content with latest compression status
			m.viewport.SetContent(m.renderCompressionView())
		}
		return m, tea.Tick(BootTickSpeed, func(t time.Time) tea.Msg { return tickMsg(t) })

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
		m.auditResults = msg.results
		m.state = statusAuditReport
		m.list.SetItems(make([]list.Item, 0))
		m.auditFolder = msg.folder
		m.auditOnline = msg.online
		m.auditFailed = msg.failed

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
		m.viewport.SetContent(m.renderCompressionView())
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

		// Update viewport with latest status
		m.viewport.SetContent(m.renderCompressionView())

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

	case overwatchDataMsg:
		// Calculate network rates (bytes per second)
		now := time.Now()
		netRXRate := uint64(0)
		netTXRate := uint64(0)
		if !m.overwatchNetTime.IsZero() {
			elapsed := now.Sub(m.overwatchNetTime).Seconds()
			if elapsed > 0 {
				if msg.netRX > m.overwatchNetRXPrev {
					netRXRate = (msg.netRX - m.overwatchNetRXPrev) / uint64(elapsed)
				}
				if msg.netTX > m.overwatchNetTXPrev {
					netTXRate = (msg.netTX - m.overwatchNetTXPrev) / uint64(elapsed)
				}
			}
		}
		// Store current values as previous for next calculation
		m.overwatchNetRXPrev = msg.netRX
		m.overwatchNetTXPrev = msg.netTX
		m.overwatchNetTime = now
		// Store rates in the fields (reusing them for rate display)
		m.overwatchNetRX = netRXRate
		m.overwatchNetTX = netTXRate

		m.overwatchCPU = msg.cpu
		m.overwatchRAMUsed = msg.ramUsed
		m.overwatchRAMTotal = msg.ramTotal
		m.overwatchSwapUsed = msg.swapUsed
		m.overwatchSwapTotal = msg.swapTotal
		m.overwatchBattery = msg.battery
		m.overwatchUptime = msg.uptime
		m.overwatchHostname = msg.hostname
		m.overwatchProcesses = msg.processes
		// Update viewport with latest data
		m.viewport.SetContent(m.renderOverwatchView())
		// Auto-refresh every 2 seconds
		return m, tea.Tick(2*time.Second, func(t time.Time) tea.Msg { return runOverwatchScan()() })

	case overwatchKillMsg:
		// Silently handle process kill - just refresh the process list
		if msg.error == nil {
			// Process killed successfully, refresh the list
			return m, runOverwatchScan()
		}
		// If error, still refresh to update the list (process might be gone)
		return m, runOverwatchScan()
	}
	return m, cmd
}
