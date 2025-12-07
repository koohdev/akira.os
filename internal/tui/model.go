package tui

import (
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"meme-factory/internal/domain"
	"meme-factory/internal/styles"
)

type driverReadyMsg struct {
	success bool
	mode    string
}

type (
	tickMsg            time.Time
	folderMsg          []string
	hostingerMsg       []string
	browserMsg         struct{ msg, folder string }
	auditMsg           []string
	batchLoadMsg       []domain.BatchJob
	compressionInitMsg struct {
		folder string
		files  []compressionFile
	}
	compressionProgressMsg struct {
		index     int
		original  int64
		optimized int64
		err       error
	}
)

type status int

const (
	statusBoot status = iota
	statusMenu
	statusGenesisInput
	statusGenesisThemeSelect
	statusGenesisCustomTheme
	statusGenesisPlatformSelect
	statusGenesisSocials
	statusAction
	statusProcessing
	statusMusic
	statusDone
	statusPostGen
	statusPatchMenu
	statusHostingerList
	statusManual
	statusAuditReport
	statusBatchMenu
	statusBatchFileSelect
	statusBatchRun
	statusCompress
)

type item struct{ title, desc, id string }

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 3 }
func (d itemDelegate) Spacing() int                            { return 1 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	if index == m.Index() {
		fmt.Fprintf(w, " [+] %s\n     %s", styles.TxtRed.Bold(true).Render(strings.ToUpper(i.title)), styles.TxtOrange.Render(i.desc))
	} else {
		fmt.Fprintf(w, "    %s\n     %s", styles.TxtGray.Render(i.title), styles.TxtGray.Faint(true).Render(i.desc))
	}
}

type model struct {
	state     status
	list      list.Model
	textInput textinput.Model
	spinner   spinner.Model
	viewport  viewport.Model

	inputs         []string
	inputStep      int
	opMode         string
	selectedTarget string
	patchField     string
	lastGenFolder  string

	// Batch Processing State
	batchQueue    []domain.BatchJob
	batchProgress int
	batchLog      []string
	batchFile     string

	// Dev Mode State
	devMode       bool
	driverModeStr string

	// Diagnostics & Visuals
	waveHistory     []int
	globalTick      int
	bootProgress    int
	diagRPM         int
	diagTemp        int
	statusMsg       string
	driverConnected bool
	blinkState      bool

	// Real Diagnostics
	startTime time.Time
	ping      int

	activePlayer *exec.Cmd
	musicPlaying bool
	currentSong  string

	width, height int
	auditResults  []string

	// Compression UI
	compressionFolder    string
	compressionEntries   []compressionEntry
	compressionFiles     []compressionFile
	compressionProcessed int
	compressionSaved     int64
}

type compressionFile struct {
	Name string
	Path string
	Size int64
}

type compressionEntry struct {
	Name      string
	Status    string
	Progress  float64
	Original  int64
	Optimized int64
	Saved     int64
	Message   string
}

func Run() error {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func initialModel() model {
	l := list.New([]list.Item{}, itemDelegate{}, 0, 0)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(false)

	ti := textinput.New()
	ti.Cursor.Style = styles.TxtRed
	ti.Prompt = " > "
	ti.TextStyle = styles.TxtBlue
	ti.Focus()

	s := spinner.New()
	s.Spinner = spinner.Line
	s.Style = styles.TxtRed

	vp := viewport.New(0, 0)
	vp.Style = lipgloss.NewStyle().PaddingLeft(2).PaddingRight(2)

	return model{
		state: statusBoot,
		list:  l, textInput: ti, spinner: s, viewport: vp,
		inputs:      make([]string, 0),
		waveHistory: make([]int, 20),
		diagRPM:     5000, diagTemp: 65,
		auditResults:  make([]string, 0),
		batchLog:      make([]string, 0),
		batchFile:     "jobs.json",
		devMode:       false,
		driverModeStr: "INIT",
		startTime:     time.Now(),
		ping:          35,
	}
}

func (m model) Init() tea.Cmd {
	m.setManualText()
	return tea.Batch(
		rebootDriver(false),
		tea.Tick(time.Millisecond*50, func(t time.Time) tea.Msg { return tickMsg(t) }),
		m.spinner.Tick,
	)
}

func (m *model) updateMainMenu() {
	devStatus := "OFF (Silent)"
	if m.devMode {
		devStatus = "ON (Visible + Debugger)"
	}

	items := []list.Item{
		item{title: "PROJECT: AWAKENING", desc: "Genesis Mode (Single Site)", id: "GENESIS"},
		item{title: "BATCH PROCESSING", desc: "The Lazy Degen Protocol (Multi-Site)", id: "BATCH"},
		item{title: "PSYCHIC INJECTION", desc: "Smart Patch (Edit Site)", id: "PATCH"},
		item{title: "IMAGE COMPRESSION", desc: "Optimize /images for speed", id: "COMPRESS"},
		item{title: "SATELLITE UPLINK", desc: "Hostinger Warp", id: "HOSTINGER"},
		item{title: "VISUAL CONFIRMATION", desc: "Preview Site", id: "PREVIEW"},
		item{title: "AUDIT: LINK HEALTH", desc: "Check all external dependencies", id: "AUDIT"},
		item{title: fmt.Sprintf("DEV MODE: %s", devStatus), desc: "Toggle Browser Visibility", id: "TOGGLE_DEV"},
		item{title: "JUKEBOX (SHOJI)", desc: "Audio Deck (.WAV ONLY)", id: "MUSIC"},
		item{title: "NETRUNNER'S CODEX", desc: "System Manual", id: "MANUAL"},
		item{title: "TERMINATE", desc: "Exit System", id: "EXIT"},
	}
	m.list.SetItems(items)
}
