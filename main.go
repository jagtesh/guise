package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// --- Constants ---

const (
	AppName    = "guise"
	StoreDir   = "store"
	ConfigFile = "config.json"
)

// --- Styles ---

var (
	// Colors

purple    = lipgloss.Color("#7D56F4")
green     = lipgloss.Color("#04B575")
lightGray = lipgloss.Color("#767676")
white     = lipgloss.Color("#FAFAFA")
bg        = lipgloss.Color("#1A1B26")

	// Styles

	windowStyle = lipgloss.NewStyle().
			Padding(1, 2).
			Background(bg).
			Foreground(white)

	headerStyle = lipgloss.NewStyle().
			Foreground(purple).
			Bold(true).
			MarginBottom(1)

	subHeaderStyle = lipgloss.NewStyle().
			Foreground(lightGray).
			Italic(true).
			MarginBottom(2)

	providerBoxStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(purple).
				Padding(1, 2).
				MarginRight(2).
				Width(40).
				Height(20)

	profileBoxStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lightGray).
				Padding(1, 2).
				Width(60).
				Height(20)

	selectedItemStyle = lipgloss.NewStyle().
				Foreground(purple).
				Bold(true)

	activeItemStyle = lipgloss.NewStyle().
				Foreground(green).
				Bold(true)

	hintStyle = lipgloss.NewStyle().
			Foreground(lightGray).
			MarginTop(1)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true)
)

// --- Data Models ---

type Profile struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type Provider struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	TargetPath      string    `json:"target_path"`
	Profiles        []Profile `json:"profiles"`
	ActiveProfileID string    `json:"active_profile_id"`
}

type Config struct {
	Providers []Provider `json:"providers"`
}

// --- App State ---

type AppState int

const (
	StateProviderList AppState = iota
	StateProfileList
	StateInputName
)

type model struct {
	config     Config
	configPath string
	baseDir    string

	state               AppState
	cursor              int
	selectedProviderIdx int

	// Terminal dimensions
	width  int
	height int

	textInput string
	err       error
}

func main() {
	userHome, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Error getting home directory: %v\n", err)
		os.Exit(1)
	}

	baseDir := filepath.Join(userHome, "."+AppName)
	configPath := filepath.Join(baseDir, ConfigFile)

	if err := os.MkdirAll(filepath.Join(baseDir, StoreDir), 0755); err != nil {
		fmt.Printf("Error creating base directory: %v\n", err)
		os.Exit(1)
	}

	// Load or Initialize Config
	cfg, err := loadConfig(configPath)
	if err != nil {
		cfg = Config{Providers: []Provider{}}
	}

	// Define Defaults
	defaultProviders := []Provider{
		{
			ID:         "openai-codex",
			Name:       "OpenAI Codex",
			TargetPath: filepath.Join(userHome, ".codex"),
			Profiles:   []Profile{},
		},
		{
			ID:         "google-gemini",
			Name:       "Google Gemini",
			TargetPath: filepath.Join(userHome, ".gemini"),
			Profiles:   []Profile{},
		},
		{
			ID:         "anthropic-claude",
			Name:       "Anthropic Claude",
			TargetPath: filepath.Join(userHome, ".claude"),
			Profiles:   []Profile{},
		},
		{
			ID:         "github-copilot",
			Name:       "GitHub Copilot CLI",
			TargetPath: getStandardConfigPath("github-copilot", userHome),
			Profiles:   []Profile{},
		},
	}

	// Merge Defaults (Append if missing)
	dirty := false

	for _, def := range defaultProviders {
		found := false
		for _, existing := range cfg.Providers {
			if existing.ID == def.ID {
				found = true
				break
			}
		}
		if !found {
			cfg.Providers = append(cfg.Providers, def)
			dirty = true
		}
	}

	// Save if we added new defaults or if it was a fresh create
	if dirty || err != nil {
		if saveErr := saveConfig(configPath, cfg); saveErr != nil {
			fmt.Printf("Error saving config: %v\n", saveErr)
			os.Exit(1)
		}
	}

	p := tea.NewProgram(model{
		config     : cfg,
		configPath : configPath,
		baseDir    : baseDir,
		state      : StateProviderList,
	}, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}

// --- Logic ---

func loadConfig(path string) (Config, error) {
	var cfg Config
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	err = json.Unmarshal(data, &cfg)
	return cfg, err
}

func saveConfig(path string, cfg Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		if m.state == StateInputName {
			return updateInputName(m, msg)
		}

		switch msg.String() {
		case "q":
			return m, tea.Quit
		}
	}

	switch m.state {
	case StateProviderList:
		return updateProviderList(m, msg)
	case StateProfileList:
		return updateProfileList(m, msg)
	}
	return m, nil
}

func updateProviderList(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				m.selectedProviderIdx = m.cursor
			}
		case "down", "j":
			if m.cursor < len(m.config.Providers)-1 {
				m.cursor++
				m.selectedProviderIdx = m.cursor
			}
		case "enter", "right", "l":
			m.selectedProviderIdx = m.cursor
			m.cursor = 0
			m.state = StateProfileList
		}
	}
	return m, nil
}

func updateProfileList(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	provider := &m.config.Providers[m.selectedProviderIdx]
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "left", "h":
			m.state = StateProviderList
			m.cursor = m.selectedProviderIdx
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(provider.Profiles)-1 {
				m.cursor++
			}
		case "n":
			m.state = StateInputName
			m.textInput = ""
		case "d":
			if len(provider.Profiles) > 0 {
				m.deleteProfile(m.selectedProviderIdx, m.cursor)
				if m.cursor >= len(provider.Profiles) {
					m.cursor = len(provider.Profiles) - 1
				}
				if m.cursor < 0 {
					m.cursor = 0
				}
				saveConfig(m.configPath, m.config)
			}
		case "enter":
			if len(provider.Profiles) > 0 {
				if err := m.activateProfile(m.selectedProviderIdx, m.cursor); err != nil {
					m.err = err
				} else {
					saveConfig(m.configPath, m.config)
				}
			}
		}
	}
	return m, nil
}

func updateInputName(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if m.textInput != "" {
				if err := m.createNewProfile(m.textInput); err != nil {
					m.err = err
				} else {
					saveConfig(m.configPath, m.config)
				}
				m.state = StateProfileList
				m.cursor = len(m.config.Providers[m.selectedProviderIdx].Profiles) - 1
			}
		case tea.KeyEsc:
			m.state = StateProfileList
		case tea.KeyBackspace, tea.KeyDelete:
			if len(m.textInput) > 0 {
				m.textInput = m.textInput[:len(m.textInput)-1]
			}
		case tea.KeyRunes:
			m.textInput += string(msg.Runes)
		case tea.KeySpace:
			m.textInput += " "
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.width == 0 {
		return "Initializing..."
	}

	header := headerStyle.Render("GUISE // Identity Manager")
	subHeader := subHeaderStyle.Render("Switch contexts seamlessly across your tools.")

	providerList := m.renderProviderList()
	profileList := m.renderProfileList()

	content := lipgloss.JoinHorizontal(lipgloss.Top, providerList, profileList)

	if m.state == StateInputName {
		inputBox := lipgloss.NewStyle().
				Border(lipgloss.DoubleBorder()).
				BorderForeground(purple).
				Padding(1, 3).
				Render(fmt.Sprintf("Enter Name for Profile:\n\n %s_", m.textInput))
		content = lipgloss.JoinVertical(lipgloss.Center, inputBox, "", content)
	}

	if m.err != nil {
		content += "\n\n" + errorStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	help := hintStyle.Render("\n↑/↓: navigate • enter: select/switch • n: new • d: delete • q: quit")
	fullView := lipgloss.JoinVertical(lipgloss.Left, header, subHeader, content, help)

	return windowStyle.Width(m.width).Height(m.height).Render(fullView)
}

func (m model) renderProviderList() string {
	s := ""
	for i, p := range m.config.Providers {
		prefix := "  "
		style := lipgloss.NewStyle().PaddingLeft(2)
		if i == m.cursor && m.state == StateProviderList {
			prefix = "> "
			style = selectedItemStyle
		} else if i == m.selectedProviderIdx {
			prefix = "• "
			style = lipgloss.NewStyle().Foreground(purple)
		}
		s += style.Render(fmt.Sprintf("%s%s", prefix, p.Name)) + "\n"
	}
	borderStyle := providerBoxStyle
	if m.state == StateProviderList {
		borderStyle = providerBoxStyle.Copy().BorderForeground(purple)
	} else {
		borderStyle = providerBoxStyle.Copy().BorderForeground(lightGray)
	}
	return borderStyle.Render("PROVIDERS\n\n" + s)
}

func (m model) renderProfileList() string {
	p := m.config.Providers[m.selectedProviderIdx]

	// Determine Active Name for Header
	activeName := "NONE"
	activeColor := lightGray
	for _, prof := range p.Profiles {
		if prof.ID == p.ActiveProfileID {
			activeName = prof.Name
			activeColor = green
			break
		}
	}

	headerStyled := lipgloss.NewStyle().Foreground(lightGray).Render(fmt.Sprintf("PROFILES: %s  |  ", p.Name)) +
		lipgloss.NewStyle().Foreground(activeColor).Bold(true).Render(fmt.Sprintf("ACTIVE: %s", activeName))

	s := ""
	if len(p.Profiles) == 0 {
		if _, err := os.Stat(p.TargetPath); err == nil {
			s = lipgloss.NewStyle().Foreground(purple).Render("Existing config detected at " + p.TargetPath + "\n\nPress 'n' to save it as a profile.")
		} else {
			s = lipgloss.NewStyle().Foreground(lightGray).Render("No profiles created yet.\nPress 'n' to start fresh.")
		}
	} else {
		for i, prof := range p.Profiles {
// ... existing loop logic ...
			// Layout: [Cursor(2)] [Status(2)] [Name]
			
			// 1. Cursor Column
			cursor := "  "
			if i == m.cursor && m.state == StateProfileList {
				cursor = "❯ "
			}
			cursorStyled := lipgloss.NewStyle().Foreground(purple).Render(cursor)

			// 2. Status Column
			isActive := prof.ID == p.ActiveProfileID
			status := "  "
			statusStyle := lipgloss.NewStyle().Foreground(lightGray) // Default dim
			
			if isActive {
				status = "● "
				statusStyle = statusStyle.Foreground(green)
			}
			statusStyled := statusStyle.Render(status)

			// 3. Name Column
			nameStyle := lipgloss.NewStyle().Foreground(lightGray) // Default color
			
			if i == m.cursor && m.state == StateProfileList {
				// Selected takes priority for text color to indicate focus
				nameStyle = nameStyle.Foreground(purple).Bold(true)
			} else if isActive {
				// Active but not selected
				nameStyle = nameStyle.Foreground(green).Bold(true)
			} else {
				// Inactive and not selected
				nameStyle = nameStyle.Foreground(white)
			}

			// Render Row
			s += fmt.Sprintf("%s%s%s\n", cursorStyled, statusStyled, nameStyle.Render(prof.Name))
		}
	}
	borderStyle := profileBoxStyle
	if m.state == StateProfileList {
		borderStyle = profileBoxStyle.Copy().BorderForeground(purple)
	}
	return borderStyle.Render(headerStyled + "\n\n" + s)
}

// --- Core Operations ---

func (m *model) deleteProfile(providerIdx, profileIdx int) {
	provider := &m.config.Providers[providerIdx]
	profile := provider.Profiles[profileIdx]
	os.RemoveAll(filepath.Join(m.baseDir, StoreDir, provider.ID, profile.ID))
	provider.Profiles = append(provider.Profiles[:profileIdx], provider.Profiles[profileIdx+1:]...)
	if provider.ActiveProfileID == profile.ID {
		provider.ActiveProfileID = ""
	}
}

func (m *model) createNewProfile(name string) error {
	provider := &m.config.Providers[m.selectedProviderIdx]
	newID := fmt.Sprintf("profile_%d", time.Now().UnixNano())
	if provider.ActiveProfileID != "" {
		m.backupProfile(provider, provider.ActiveProfileID)
	}
	newProfile := Profile{ID: newID, Name: name, CreatedAt: time.Now()}
	provider.Profiles = append(provider.Profiles, newProfile)
	provider.ActiveProfileID = newID
	storePath := filepath.Join(m.baseDir, StoreDir, provider.ID, newID)
	if len(provider.Profiles) == 1 {
		if err := copyDir(provider.TargetPath, storePath); err != nil {
			os.MkdirAll(storePath, 0755)
		}
	} else {
		os.RemoveAll(provider.TargetPath)
		os.MkdirAll(provider.TargetPath, 0755)
		os.MkdirAll(storePath, 0755)
	}
	return nil
}

func (m *model) activateProfile(pIdx, profIdx int) error {
	provider := &m.config.Providers[pIdx]
	target := provider.Profiles[profIdx]
	if provider.ActiveProfileID == target.ID {
		return nil
	}
	if provider.ActiveProfileID != "" {
		m.backupProfile(provider, provider.ActiveProfileID)
	}
	src := filepath.Join(m.baseDir, StoreDir, provider.ID, target.ID)
	os.RemoveAll(provider.TargetPath)
	if err := copyDir(src, provider.TargetPath); err != nil {
		return err
	}
	provider.ActiveProfileID = target.ID
	return nil
}

func (m *model) backupProfile(p *Provider, id string) error {
	dst := filepath.Join(m.baseDir, StoreDir, p.ID, id)
	os.RemoveAll(dst)
	if _, err := os.Stat(p.TargetPath); os.IsNotExist(err) {
		return os.MkdirAll(dst, 0755)
	}
	return copyDir(p.TargetPath, dst)
}

func copyDir(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dst, info.Mode()); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, e := range entries {
		s, d := filepath.Join(src, e.Name()), filepath.Join(dst, e.Name())
		if e.IsDir() {
			copyDir(s, d)
		} else {
			copyFile(s, d)
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}
// getStandardConfigPath returns the platform-specific standard config path for a CLI tool.
// Windows: %LOCALAPPDATA%\<tool>
// macOS/Linux: ~/.config/<tool>
func getStandardConfigPath(tool, home string) string {
	if runtime.GOOS == "windows" {
		// Try LocalAppData first for CLIs, as AppData (Roaming) is for roaming profiles
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData != "" {
			return filepath.Join(localAppData, tool)
		}
		// Fallback to AppData
		appData, _ := os.UserConfigDir()
		if appData != "" {
			return filepath.Join(appData, tool)
		}
	}

	// Unix-like (Linux/macOS) standard: ~/.config/<tool>
	// Note: os.UserConfigDir() on macOS returns ~/Library/Application Support, 
	// but most CLIs prefer ~/.config. We'll stick to ~/.config for consistency with CLI tools.
	return filepath.Join(home, ".config", tool)
}
