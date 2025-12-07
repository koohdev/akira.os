
# PROJECT: AWAKENING (Meme Factory v3.8)
> **"MAKE MEME COINS GREAT AGAIN"**
> An elite, cyberpunk-styled CLI tool for the autonomous generation, auditing, and patching of Web3 landing pages.

## 📋 Overview

**Project: Awakening** is a terminal-based automation suite designed for "Degens" and Web3 developers. It bypasses standard API limits by using **Selenium-style browser automation (via Go-Rod)** to control a local Chrome instance.

It acts as a bridge between raw user input and **Google Gemini AI**, prompting the AI to write production-ready HTML5/TailwindCSS landing pages. It also includes modules for asset management, link auditing, and server-side navigation.

## 📑 Table of Contents

  * [Overview](https://www.google.com/search?q=%23-overview)
  * [Features](https://www.google.com/search?q=%23-features)
  * [Prerequisites](https://www.google.com/search?q=%23-prerequisites)
  * [Quickstart](https://www.google.com/search?q=%23-quickstart)
  * [Project Index](https://www.google.com/search?q=%23-project-index)
  * [Project Structure](https://www.google.com/search?q=%23-project-structure)
  * [Roadmap](https://www.google.com/search?q=%23-roadmap)
  * [Contribution](https://www.google.com/search?q=%23-contribution)
  * [License](https://www.google.com/search?q=%23-license)

## ✨ Features

  * **Genesis Mode (Site Generator):** Takes a Ticker, Theme, and Platform (Solana, Pump.fun, Base, etc.) and uses Google Gemini to generate a high-conversion "One-Pager" website.
  * **Smart Patch (Injection):** Allows you to edit existing generated sites by injecting new Contract Addresses (CA), Social Links, or DexScreener URLs without touching the code manually.
  * **Audit Module:** Scans generated `index.html` files and verifies the HTTP status of every external link (200 OK vs 404/Dead Links) to ensure launch readiness.
  * **Hostinger Warp:** Automates navigation to the Hostinger HPanel to view active sites (requires login in the automated browser).
  * **Jukebox (Shoji):** Built-in .WAV audio player that runs in the background while you work (uses PowerShell audio drivers).
  * **Chrome Sandbox:** Runs a dedicated, minimized Chrome instance at `C:\selenium` to handle AI prompts and scraping.

## ⚙️ Prerequisites

  * **OS:** Windows 10/11 (Required for file paths, PowerShell audio, and `taskkill` commands).
  * **Go:** Version 1.21 or higher.
  * **Google Chrome:** Must be installed at default locations.
  * **Gemini Account:** You must be logged into Google Gemini in the automated Chrome window for Genesis mode to work.

## 🚀 Quickstart

### 1\. Clone & Install

```bash
git clone https://github.com/yourusername/project-awakening.git
cd project-awakening
go mod tidy
```

### 2\. The Chrome Driver

The app attempts to launch Chrome automatically on port `9222`.

  * **Warning:** The app runs `taskkill` on existing Chrome processes to free up the debugging port. Save your work in your browser before running.

### 3\. Run the System

```bash
go run main.go
```

### 4\. First Run Protocol

1.  The app will launch a **minimized** Chrome window.
2.  Maximize this window once.
3.  **Log in to Google (Gemini)** and **Hostinger** inside this specific window.
4.  Minimize it again and return to the terminal.

## rw Project Index (Menu)

| ID | Module Name | Description |
| :--- | :--- | :--- |
| **GENESIS** | Project: Awakening | The core AI generator. Prompts for Ticker/Theme and writes code. |
| **PATCH** | Psychic Injection | Updates variables (CA, Telegram, X) in existing HTML files. |
| **HOSTINGER** | Satellite Uplink | Scrapes your active Hostinger website list. |
| **PREVIEW** | Visual Confirmation | Opens the generated site in your default browser. |
| **AUDIT** | Link Health | Checks generated sites for broken links/404s. |
| **MUSIC** | Jukebox (Shoji) | Plays .wav files from the `/music` directory. |
| **MANUAL** | Netrunner's Codex | In-app help documentation. |

## 📂 Project Structure

When you run the tool, it creates/uses the following directory structure:

```text
root/
├── main.go             # The Brain (Source Code)
├── music/              # Place .wav files here for the Jukebox
├── websites/           # Output directory for generated sites
│   └── TICKER_DATE/    # Specific project folder
│       ├── index.html  # The AI-generated code
│       ├── images/     # Asset folders (Socials, Stickers)
│       └── ...
└── C:/selenium/        # (External) Chrome Profile storage
```

## 🗺 Roadmap

  * [ ] **Cross-Platform Support:** Refactor pathing and audio drivers to support MacOS and Linux.
  * [ ] **API Integration:** Replace browser automation with direct Gemini API key integration for faster generation.
  * [ ] **Image Gen:** Auto-download assets based on generated prompts.
  * [ ] **IPFS Deploy:** One-click deployment to IPFS or Netlify.

## 🤝 Contribution

This tool is designed for the "Degen" community. Pull requests are welcome.

1.  Fork the Project
2.  Create your Feature Branch (`git checkout -b feature/NewModule`)
3.  Commit your Changes
4.  Push to the Branch
5.  Open a Pull Request

## 📄 License

**© 2025 @GELLOVESUALWAYS**
Distributed under the MIT License. See `LICENSE` for more information.

*Disclaimer: This tool is for educational and creative purposes. The author is not responsible for financial losses associated with meme coins generated by this software.*