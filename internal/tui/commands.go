package tui

import (
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"

	"meme-factory/internal/config"
	"meme-factory/internal/domain"
	"meme-factory/internal/styles"
)

func (m *model) setManualText() {
	manual := `
 NETRUNNER'S CODEX (MANUAL)
 --------------------------

 1. GENESIS MODE
    Constructs full websites. You will be prompted for Ticker, Theme, etc.
    
 2. BATCH MODE (NEW)
    "The Lazy Degen Protocol".
    - Select [CREATE TEMPLATE] to generate 'jobs.json' in 'batch_jobs' folder.
    - Fill it out with your coin ideas.
    - Select [EXECUTE BATCH] to build them all automatically.

 3. DEV MODE (NEW)
    Toggles Chrome Visibility. Useful for debugging AI inputs.
    - ON: Visible Window + DevTools.
    - OFF: Headless/Minimized (Faster).

 4. SMART PATCH
    Edit existing sites. Select a folder, then choose what to inject.

 5. HOSTINGER WARP
    Scans your Hostinger account.
    NOTE: You MUST log in to Hostinger in the minimized Chrome window!

 [ UP / DOWN / MOUSE ] SCROLL     [ ENTER ] RETURN
`
	m.viewport.SetContent(styles.TxtBlue.Render(manual))
}

func getNextBatchFileName() string {
	base := "batch_jobs"
	os.MkdirAll(base, 0755)

	if _, err := os.Stat(filepath.Join(base, "jobs.json")); os.IsNotExist(err) {
		return "jobs.json"
	}

	i := 2
	for {
		name := fmt.Sprintf("jobs(%d).json", i)
		if _, err := os.Stat(filepath.Join(base, name)); os.IsNotExist(err) {
			return name
		}
		i++
	}
}

func createBatchTemplate() tea.Cmd {
	return func() tea.Msg {
		examples := []domain.BatchJob{
			{
				Ticker:   "$EXAMPLE",
				Theme:    "Cyberpunk",
				Platform: "SOLANA",
				Twitter:  "https://x.com/example",
				Telegram: "https://t.me/example",
				Discord:  "None",
				Dex:      "None",
				Assets:   "https://imgur.com/link_to_logos (Or 'None')",
				Context:  "Make it extremely bullish. High energy.",
			},
			{
				Ticker:   "$SECOND_COIN",
				Theme:    "Retro",
				Platform: "BASE",
				Twitter:  "https://x.com/second",
				Telegram: "None",
				Discord:  "None",
				Dex:      "None",
				Assets:   "None",
				Context:  "A calm, safe project.",
			},
		}

		if err := os.MkdirAll("batch_jobs", 0755); err != nil {
			return browserMsg{msg: "ERR: COULD NOT CREATE BATCH DIRECTORY"}
		}

		fileName := getNextBatchFileName()
		fullPath := filepath.Join("batch_jobs", fileName)

		file, _ := json.MarshalIndent(examples, "", "  ")
		err := os.WriteFile(fullPath, file, 0644)
		if err != nil {
			return browserMsg{msg: "ERR: COULD NOT CREATE TEMPLATE"}
		}
		return browserMsg{msg: fmt.Sprintf("MISSION REPORT: SUCCESS: %s CREATED. EDIT IT NOW.", fileName)}
	}
}

func loadBatchFilesList() tea.Cmd {
	return func() tea.Msg {
		pwd, _ := os.Getwd()
		path := filepath.Join(pwd, "batch_jobs")
		os.MkdirAll(path, 0755)
		entries, _ := os.ReadDir(path)
		var files []string
		for _, e := range entries {
			if strings.HasSuffix(e.Name(), ".json") {
				files = append(files, e.Name())
			}
		}
		return folderMsg(files)
	}
}

func loadBatchFile(filename string) tea.Cmd {
	return func() tea.Msg {
		path := filepath.Join("batch_jobs", filename)
		content, err := os.ReadFile(path)
		if err != nil {
			return browserMsg{msg: fmt.Sprintf("ERR: %s NOT FOUND.", filename)}
		}

		var jobs []domain.BatchJob
		if err := json.Unmarshal(content, &jobs); err != nil {
			return browserMsg{msg: "ERR: INVALID JSON FORMAT."}
		}

		return batchLoadMsg(jobs)
	}
}

func openBatchFolder() {
	pwd, _ := os.Getwd()
	path := filepath.Join(pwd, "batch_jobs")
	exec.Command("explorer", path).Start()
}

func checkDriverPort() bool {
	conn, err := net.DialTimeout("tcp", "127.0.0.1:9222", time.Second)
	if err == nil {
		conn.Close()
		return true
	}
	return false
}

func loadFolders() tea.Cmd {
	return func() tea.Msg {
		pwd, _ := os.Getwd()
		path := filepath.Join(pwd, "websites")
		os.MkdirAll(path, 0755)
		entries, _ := os.ReadDir(path)
		var folders []string
		for _, e := range entries {
			if e.IsDir() {
				folders = append(folders, e.Name())
			}
		}
		return folderMsg(folders)
	}
}

func (m *model) loadThemeOptions() {
	items := []list.Item{
		item{title: "CYBERPUNK", desc: "High contrast, neon, glitch", id: "CYBERPUNK"},
		item{title: "RETROWAVE", desc: "Sunsets, grids, 80s vibe", id: "RETROWAVE"},
		item{title: "DEGEN / MEME", desc: "Chaotic, comic sans, vibrant", id: "DEGEN"},
		item{title: "DARK MATTER", desc: "Deep black, subtle purple", id: "DARK"},
		item{title: "CUSTOM", desc: "Type your own style...", id: "CUSTOM"},
	}
	m.list.SetItems(items)
}

func (m *model) loadPlatformOptions() {
	items := []list.Item{
		item{title: "SOLANA", desc: "SPL Standard", id: "SOL"},
		item{title: "PUMP.FUN", desc: "Pump Bonding Curve", id: "PUMP"},
		item{title: "ETHEREUM", desc: "ERC-20 Standard", id: "ETH"},
		item{title: "BSC", desc: "Binance Smart Chain", id: "BSC"},
		item{title: "BASE", desc: "Coinbase L2", id: "BASE"},
		item{title: "MONAD", desc: "Next Gen L1", id: "MONAD"},
	}
	m.list.SetItems(items)
}

func (m *model) loadPatchOptions() {
	items := []list.Item{
		item{title: "Inject CA", desc: "Update CA_PLACEHOLDER", id: "Inject CA"},
		item{title: "Update X", desc: "Update X_LINK_PLACEHOLDER", id: "Update X"},
		item{title: "Update Telegram", desc: "Update TG_LINK_PLACEHOLDER", id: "Update Telegram"},
		item{title: "Update Discord", desc: "Update DISCORD_PLACEHOLDER", id: "Update Discord"},
		item{title: "Update DexScreener", desc: "Update DEX_LINK_PLACEHOLDER", id: "Update DexScreener"},
	}
	m.list.SetItems(items)
}

func loadMusic() tea.Cmd {
	return func() tea.Msg {
		pwd, _ := os.Getwd()
		path := filepath.Join(pwd, "music")
		os.MkdirAll(path, 0755)
		entries, _ := os.ReadDir(path)
		var files []string
		for _, e := range entries {
			if strings.HasSuffix(strings.ToLower(e.Name()), ".wav") {
				files = append(files, e.Name())
			}
		}
		if len(files) == 0 {
			files = append(files, "NO .WAV FILES FOUND")
		}
		return folderMsg(files)
	}
}

func killPlayer(cmd *exec.Cmd) {
	if cmd != nil && cmd.Process != nil {
		cmd.Process.Kill()
	}
}

func playAudio(filename string) *exec.Cmd {
	pwd, _ := os.Getwd()
	fullPath := filepath.Join(pwd, "music", filename)
	safePath := strings.ReplaceAll(fullPath, "'", "''")
	psScript := fmt.Sprintf(`$player = New-Object System.Media.SoundPlayer; $player.SoundLocation = '%s'; $player.PlayLooping(); while($true) { Start-Sleep 1 }`, safePath)
	cmd := exec.Command("powershell", "-c", psScript)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cmd.Start()
	return cmd
}

func openBrowser(folder string) {
	pwd, _ := os.Getwd()
	path := filepath.Join(pwd, "websites", folder, "index.html")
	exec.Command("cmd", "/c", "start", "", path).Start()
}

func openExplorer(folder string) {
	pwd, _ := os.Getwd()
	path := filepath.Join(pwd, "websites", folder)
	exec.Command("explorer", path).Start()
}

func runHostingerScan() tea.Cmd {
	return func() tea.Msg {
		defer func() { recover() }()
		u := launcher.MustResolveURL("127.0.0.1:9222")
		browser := rod.New().ControlURL(u).MustConnect()
		page := browser.MustPage("https://hpanel.hostinger.com/websites")
		page.MustWaitLoad()
		var sites []string
		page.Race().Element(".website-card").Handle(func(e *rod.Element) error {
			els := page.MustElements(".website-card")
			for _, el := range els {
				txt, _ := el.Text()
				sites = append(sites, strings.Split(txt, "\n")[0])
			}
			return nil
		}).Element("body").MustDo()
		if len(sites) == 0 {
			sites = append(sites, "Manual Dashboard (Scan Failed)")
		}
		return hostingerMsg(sites)
	}
}

func runHostingerNav(domain string) tea.Cmd {
	return func() tea.Msg {
		defer func() { recover() }()
		u := launcher.MustResolveURL("127.0.0.1:9222")
		browser := rod.New().ControlURL(u).MustConnect()
		var page *rod.Page
		if domain == "Manual Dashboard (Scan Failed)" || domain == "" {
			page = browser.MustPage("https://hpanel.hostinger.com/")
		} else {
			page = browser.MustPage(fmt.Sprintf("https://hpanel.hostinger.com/hosting/%s", domain))
		}
		page.MustWaitLoad()
		page.MustActivate()
		return browserMsg{msg: "UPLINK SECURE. DASHBOARD ACTIVE."}
	}
}

func executePatch(folder, action, value string) {
	pwd, _ := os.Getwd()
	path := filepath.Join(pwd, "websites", folder, "index.html")
	content, err := os.ReadFile(path)
	if err != nil {
		return
	}
	text := string(content)
	switch action {
	case "Inject CA":
		text = strings.ReplaceAll(text, "CA_PLACEHOLDER", value)
	case "Update X":
		text = strings.ReplaceAll(text, "X_LINK_PLACEHOLDER", value)
	case "Update Telegram":
		text = strings.ReplaceAll(text, "TG_LINK_PLACEHOLDER", value)
	case "Update Discord":
		text = strings.ReplaceAll(text, "DISCORD_PLACEHOLDER", value)
	case "Update DexScreener":
		text = strings.ReplaceAll(text, "DEX_LINK_PLACEHOLDER", value)
	}
	os.WriteFile(path, []byte(text), 0644)
}

func parseAssetPrompts(htmlContent string, fullPath string) {
	startTag := "### 2. THE ASSET VAULT (DYNAMIC LOGOS)"
	endTag := "#### STEP 3: 💻 THE CODE"

	start := strings.Index(htmlContent, startTag)
	if start == -1 {
		return
	}

	contentAfterStart := htmlContent[start:]
	end := strings.Index(contentAfterStart, endTag)

	promptContent := ""
	if end != -1 {
		promptContent = contentAfterStart[:end]
	} else {
		promptContent = contentAfterStart
	}

	cleanedContent := strings.TrimSpace(promptContent)
	promptPath := filepath.Join(fullPath, "asset_prompts.txt")
	os.WriteFile(promptPath, []byte(cleanedContent), 0644)
}

func runAudit(folder string) tea.Cmd {
	return func() tea.Msg {
		pwd, _ := os.Getwd()
		path := filepath.Join(pwd, "websites", folder, "index.html")
		content, err := os.ReadFile(path)
		if err != nil {
			return auditMsg{fmt.Sprintf("FATAL: Could not read index.html: %v", err)}
		}
		text := string(content)

		re := regexp.MustCompile(`(href|src)="([^"]*http[^"]*)"`)
		matches := re.FindAllStringSubmatch(text, -1)

		var urls []string
		for _, match := range matches {
			if len(match) > 2 {
				urls = append(urls, match[2])
			}
		}

		uniqueURLs := make(map[string]struct{})
		for _, url := range urls {
			uniqueURLs[url] = struct{}{}
		}

		results := []string{}
		client := http.Client{Timeout: 8 * time.Second}

		for url := range uniqueURLs {
			if strings.Contains(url, "PLACEHOLDER") {
				line := fmt.Sprintf("ℹ️ INFO     %-60s [PLACEHOLDER]", url)
				results = append(results, styles.TxtBlue.Render(line))
				continue
			}

			if strings.HasPrefix(url, "http") {
				req, err := http.NewRequest("GET", url, nil)
				if err != nil {
					line := fmt.Sprintf("::FAIL:: FAIL     %s - Request Error", url)
					results = append(results, styles.TxtRed.Render(line))
					continue
				}

				resp, err := client.Do(req)
				if err != nil {
					line := fmt.Sprintf("::FAIL:: FAIL     %s - Connection Error", url)
					results = append(results, styles.TxtRed.Render(line))
					continue
				}
				defer resp.Body.Close()

				statusText := fmt.Sprintf("[STATUS %d]", resp.StatusCode)

				if resp.StatusCode >= 200 && resp.StatusCode < 300 {
					line := fmt.Sprintf("::OK:: ONLINE   %-60s %s", url, statusText)
					results = append(results, styles.TxtGreen.Render(line))
				} else if resp.StatusCode >= 300 && resp.StatusCode < 400 {
					line := fmt.Sprintf("::WARN:: WARNING %-60s %s", url, statusText)
					results = append(results, styles.TxtOrange.Render(line))
				} else {
					line := fmt.Sprintf("::FAIL:: FAIL     %-60s %s", url, statusText)
					results = append(results, styles.TxtRed.Render(line))
				}
			}
		}

		if len(results) == 0 {
			results = append(results, styles.TxtBlue.Render("No external HTTP links detected in index.html."))
		}

		return auditMsg(results)
	}
}

// runImageCompress walks websites/<folder>/images and recompresses PNG/JPEG files.
// It keeps dimensions the same, only reducing file size/quality slightly.
func runImageCompress(folder string) tea.Cmd {
	return func() tea.Msg {
		pwd, _ := os.Getwd()
		root := filepath.Join(pwd, "websites", folder, "images")

		if _, err := os.Stat(root); os.IsNotExist(err) {
			return browserMsg{msg: fmt.Sprintf("ERR: IMAGES FOLDER NOT FOUND FOR %s", folder)}
		}

		var files []compressionFile
		filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d == nil {
				return nil
			}
			if d.IsDir() {
				return nil
			}
			lower := strings.ToLower(d.Name())
			if !(strings.HasSuffix(lower, ".jpg") || strings.HasSuffix(lower, ".jpeg") || strings.HasSuffix(lower, ".png")) {
				return nil
			}
			info, err := d.Info()
			if err != nil {
				return nil
			}
			rel, _ := filepath.Rel(root, path)
			files = append(files, compressionFile{
				Name: rel,
				Path: path,
				Size: info.Size(),
			})
			return nil
		})

		if len(files) == 0 {
			return compressionInitMsg{folder: folder, files: []compressionFile{}}
		}
		return compressionInitMsg{folder: folder, files: files}
	}
}

func compressFileCmd(file compressionFile, index int) tea.Cmd {
	return func() tea.Msg {
		original := file.Size
		optimized := original
		var opErr error

		src, err := os.Open(file.Path)
		if err != nil {
			return compressionProgressMsg{index: index, original: original, optimized: optimized, err: err}
		}
		img, format, err := image.Decode(src)
		src.Close()
		if err != nil {
			return compressionProgressMsg{index: index, original: original, optimized: optimized, err: err}
		}

		out, err := os.Create(file.Path)
		if err != nil {
			return compressionProgressMsg{index: index, original: original, optimized: optimized, err: err}
		}
		switch strings.ToLower(format) {
		case "jpeg", "jpg":
			opErr = jpeg.Encode(out, img, &jpeg.Options{Quality: 80})
		case "png":
			enc := png.Encoder{CompressionLevel: png.BestCompression}
			opErr = enc.Encode(out, img)
		default:
			opErr = fmt.Errorf("unsupported format: %s", format)
		}
		out.Close()

		if opErr != nil {
			return compressionProgressMsg{index: index, original: original, optimized: optimized, err: opErr}
		}

		info, err := os.Stat(file.Path)
		if err == nil {
			optimized = info.Size()
		}
		return compressionProgressMsg{index: index, original: original, optimized: optimized, err: nil}
	}
}

func runGenesis(inputs []string) tea.Cmd {
	return func() tea.Msg {
		defer func() { recover() }()
		u := launcher.MustResolveURL("127.0.0.1:9222")
		browser := rod.New().ControlURL(u).MustConnect()
		page := browser.MustPage("https://gemini.google.com/app")
		page.MustWaitLoad()

		time.Sleep(1 * time.Second)

		inputBox, err := page.Element("rich-textarea")
		if err != nil {
			inputBox = page.MustElement("div[contenteditable='true']")
		}
		inputBox.MustClick()
		time.Sleep(500 * time.Millisecond)

		data := fmt.Sprintf("\nDATA: TICKER:%s THEME:%s PLATFORM:%s LINKS:%s %s ASSETS:%s CONTEXT:%s", inputs[0], inputs[1], inputs[2], inputs[3], inputs[4], inputs[8], inputs[9])
		inputBox.MustInput(config.SYSTEM_PROMPT + data)

		time.Sleep(2 * time.Second)

		sendBtn, err := page.Element("button[aria-label*='Send']")
		if err != nil {
			sendBtn = page.MustElement(".send-button")
		}
		sendBtn.MustClick()

		page.MustWaitIdle()

		foundCode := false
		var el *rod.Element
		stableCount := 0
		lastLen := 0

		for i := 0; i < 300; i++ {
			if has, e, _ := page.Has("pre code"); has {
				el = e
				txt, _ := el.Text()
				currLen := len(txt)

				if currLen > 0 && currLen == lastLen {
					stableCount++
					if stableCount >= 3 {
						foundCode = true
						break
					}
				} else {
					stableCount = 0
				}
				lastLen = currLen
			}
			time.Sleep(1 * time.Second)
		}

		if !foundCode {
			return browserMsg{msg: "ERR: GEMINI TIMEOUT (NO CODE GENERATED)"}
		}

		htmlContent := ""
		if el != nil {
			htmlContent, _ = el.Text()
		}

		cleanTicker := strings.ToUpper(strings.ReplaceAll(inputs[0], "$", ""))
		dateStr := time.Now().Format("2006-01-02")
		folderName := fmt.Sprintf("%s_%s", cleanTicker, dateStr)

		pwd, _ := os.Getwd()
		fullPath := filepath.Join(pwd, "websites", folderName)

		if err := os.MkdirAll(fullPath, 0755); err != nil {
			return browserMsg{msg: fmt.Sprintf("FATAL: Cannot create project folder at %s. Error: %v", fullPath, err)}
		}

		assetBase := filepath.Join(fullPath, "images")
		subfolders := []string{"socials", "emojies", "stickers"}
		failedFolders := []string{}
		firstAssetPath := ""

		for i, sf := range subfolders {
			assetPath := filepath.Join(assetBase, sf)
			if err := os.MkdirAll(assetPath, 0755); err != nil {
				failedFolders = append(failedFolders, fmt.Sprintf("%s (Error: %v)", assetPath, err))
			}
			if i == 0 {
				firstAssetPath = assetPath
			}
		}

		if len(failedFolders) > 0 {
			return browserMsg{msg: fmt.Sprintf("CRITICAL WARNING: Asset folder creation failed. Errors: %s", strings.Join(failedFolders, "; "))}
		}

		if _, err := os.Stat(firstAssetPath); os.IsNotExist(err) {
			return browserMsg{msg: fmt.Sprintf("FATAL: Asset folders reported success, but verification failed. Check permissions for %s", firstAssetPath)}
		}

		if htmlContent != "" {
			parseAssetPrompts(htmlContent, fullPath)

			if err := os.WriteFile(filepath.Join(fullPath, "index.html"), []byte(htmlContent), 0644); err != nil {
				return browserMsg{msg: fmt.Sprintf("FATAL: Could not write index.html. Error: %v", err)}
			}
			return browserMsg{msg: "SUBJECT CONSTRUCTED: Asset folders created successfully.", folder: folderName}
		}
		return browserMsg{msg: "ERR: NO CODE DETECTED IN RESPONSE."}
	}
}

func rebootDriver(devMode bool) tea.Cmd {
	return func() tea.Msg {
		exec.Command("taskkill", "/F", "/IM", "chrome.exe", "/T").Run()
		time.Sleep(1 * time.Second)

		chromePath := "C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe"
		if _, err := os.Stat(chromePath); os.IsNotExist(err) {
			chromePath = "C:\\Program Files (x86)\\Google\\Chrome\\Application\\chrome.exe"
		}

		os.MkdirAll("C:\\selenium", 0755)

		flags := []string{
			"--remote-debugging-port=9222",
			"--user-data-dir=C:\\selenium\\ChromeProfile",
			"--no-first-run",
			"--no-default-browser-check",
			"--disable-infobars",
			"--disable-restore-session-state",
			"--password-store=basic",
			"--use-mock-keychain",
			"--hide-crash-restore-bubble",
			"--disable-background-networking",
			"--disable-sync",
			"--disable-translate",
		}

		modeStr := "SILENT"
		if devMode {
			modeStr = "DEBUG"
			flags = append(flags, "--auto-open-devtools-for-tabs")
			flags = append(flags, "--start-maximized")
		} else {
			modeStr = "SILENT"
			flags = append(flags, "--start-minimized")
			flags = append(flags, "--window-position=5000,5000")
		}

		cmd := exec.Command(chromePath, flags...)
		err := cmd.Start()

		if err != nil {
			return driverReadyMsg{success: false, mode: "ERROR"}
		}

		time.Sleep(2 * time.Second)
		return driverReadyMsg{success: true, mode: modeStr}
	}
}

func (m *model) resetMenu() {
	m.list.ResetSelected()
	m.list.SetItems(initialModel().list.Items())
}
