// Package filepicker provides a file picker component for Bubble Tea
// applications.
package filepicker

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// Humanize functions to avoid external dependency
func humanizeBytes(s uint64) string {
	sizes := []string{"B", "KB", "MB", "GB", "TB"}
	if s < 10 {
		return fmt.Sprintf("%dB", s)
	}
	e := 0
	v := float64(s)
	for v >= 1024 && e < len(sizes)-1 {
		v /= 1024
		e++
	}
	return fmt.Sprintf("%.1f%s", v, sizes[e])
}

// Minimal key binding types
type keyBinding struct {
	keys []string
	help string
}

func (k keyBinding) Keys() []string { return k.keys }
func (k keyBinding) Help() string   { return k.help }

func matches(msg tea.KeyPressMsg, k keyBinding) bool {
	for _, key := range k.keys {
		if msg.String() == key {
			return true
		}
	}
	return false
}

type KeyMap struct {
	GoToTop  keyBinding
	GoToLast keyBinding
	Down     keyBinding
	Up       keyBinding
	PageUp   keyBinding
	PageDown keyBinding
	Back     keyBinding
	Open     keyBinding
	Select   keyBinding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		GoToTop:  keyBinding{keys: []string{"g"}, help: "first"},
		GoToLast: keyBinding{keys: []string{"G"}, help: "last"},
		Down:     keyBinding{keys: []string{"j", "down"}, help: "down"},
		Up:       keyBinding{keys: []string{"k", "up"}, help: "up"},
		PageUp:   keyBinding{keys: []string{"K", "pgup"}, help: "page up"},
		PageDown: keyBinding{keys: []string{"J", "pgdown"}, help: "page down"},
		Back:     keyBinding{keys: []string{"h", "backspace", "left", "esc"}, help: "back"},
		Open:     keyBinding{keys: []string{"l", "right", "enter"}, help: "open"},
		Select:   keyBinding{keys: []string{"enter"}, help: "select"},
	}
}

var lastID int64

func nextID() int {
	return int(atomic.AddInt64(&lastID, 1))
}

// New returns a new filepicker model with default styling and key bindings.
func New() Model {
	dir, _ := os.Getwd()
	return Model{
		id:               nextID(),
		CurrentDirectory: dir,
		Cursor:           ">",
		AllowedTypes:     []string{},
		selected:         0,
		ShowPermissions:  true,
		ShowSize:         true,
		ShowHidden:       false,
		DirAllowed:       false,
		FileAllowed:      true,
		AutoHeight:       true,
		height:           0,
		maxIdx:           0,
		minIdx:           0,
		selectedStack:    newStack(),
		minStack:         newStack(),
		maxStack:         newStack(),
		KeyMap:           DefaultKeyMap(),
		Styles:           DefaultStyles(),
	}
}

type errorMsg struct {
	err error
}

type readDirMsg struct {
	id      int
	entries []os.DirEntry
}

const (
	marginBottom  = 5
	fileSizeWidth = 7
	paddingLeft   = 2
)

// Styles defines the possible customizations for styles in the file picker.
type Styles struct {
	DisabledCursor   lipgloss.Style
	Cursor           lipgloss.Style
	Symlink          lipgloss.Style
	Directory        lipgloss.Style
	File             lipgloss.Style
	DisabledFile     lipgloss.Style
	Permission       lipgloss.Style
	Selected         lipgloss.Style
	DisabledSelected lipgloss.Style
	FileSize         lipgloss.Style
	EmptyDirectory   lipgloss.Style
}

// DefaultStyles defines the default styling for the file picker.
func DefaultStyles() Styles {
	return Styles{
		DisabledCursor:   lipgloss.NewStyle().Foreground(lipgloss.Color("247")),
		Cursor:           lipgloss.NewStyle().Foreground(lipgloss.Color("212")),
		Symlink:          lipgloss.NewStyle().Foreground(lipgloss.Color("36")),
		Directory:        lipgloss.NewStyle().Foreground(lipgloss.Color("99")),
		File:             lipgloss.NewStyle(),
		DisabledFile:     lipgloss.NewStyle().Foreground(lipgloss.Color("243")),
		DisabledSelected: lipgloss.NewStyle().Foreground(lipgloss.Color("247")),
		Permission:       lipgloss.NewStyle().Foreground(lipgloss.Color("244")),
		Selected:         lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true),
		FileSize:         lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Width(fileSizeWidth).Align(lipgloss.Right),
		EmptyDirectory:   lipgloss.NewStyle().Foreground(lipgloss.Color("240")).PaddingLeft(paddingLeft).SetString("Bummer. No Files Found."),
	}
}

// Model represents a file picker.
type Model struct {
	id int

	// Path is the path which the user has selected with the file picker.
	Path string

	// CurrentDirectory is the directory that the user is currently in.
	CurrentDirectory string

	// AllowedTypes specifies which file types the user may select.
	// If empty the user may select any file.
	AllowedTypes []string

	KeyMap          KeyMap
	files           []os.DirEntry
	ShowPermissions bool
	ShowSize        bool
	ShowHidden      bool
	DirAllowed      bool
	FileAllowed     bool

	FileSelected  string
	selected      int
	selectedStack stack

	minIdx   int
	maxIdx   int
	maxStack stack
	minStack stack

	height     int
	AutoHeight bool

	Cursor string
	Styles Styles
}

type stack struct {
	Push   func(int)
	Pop    func() int
	Length func() int
}

func newStack() stack {
	slice := make([]int, 0)
	return stack{
		Push: func(i int) {
			slice = append(slice, i)
		},
		Pop: func() int {
			res := slice[len(slice)-1]
			slice = slice[:len(slice)-1]
			return res
		},
		Length: func() int {
			return len(slice)
		},
	}
}

func (m *Model) pushView(selected, minimum, maximum int) {
	m.selectedStack.Push(selected)
	m.minStack.Push(minimum)
	m.maxStack.Push(maximum)
}

func (m *Model) popView() (int, int, int) {
	return m.selectedStack.Pop(), m.minStack.Pop(), m.maxStack.Pop()
}

func (m Model) readDir(path string, showHidden bool) tea.Cmd {
	return func() tea.Msg {
		dirEntries, err := os.ReadDir(path)
		if err != nil {
			return errorMsg{err}
		}

		sort.Slice(dirEntries, func(i, j int) bool {
			if dirEntries[i].IsDir() == dirEntries[j].IsDir() {
				return dirEntries[i].Name() < dirEntries[j].Name()
			}
			return dirEntries[i].IsDir()
		})

		var sanitizedDirEntries []os.DirEntry

		// Add parent directory option if not at root
		parent := filepath.Dir(path)
		if parent != path {
			sanitizedDirEntries = append(sanitizedDirEntries, parentDir{})
		}

		for _, dirEntry := range dirEntries {
			if !showHidden {
				isHidden, _ := IsHidden(dirEntry.Name())
				if isHidden {
					continue
				}
			}
			sanitizedDirEntries = append(sanitizedDirEntries, dirEntry)
		}
		return readDirMsg{id: m.id, entries: sanitizedDirEntries}
	}
}

func IsHidden(name string) (bool, error) {
	return strings.HasPrefix(name, "."), nil
}

// SetHeight sets the height of the file picker.
func (m *Model) SetHeight(h int) {
	m.height = h
	if m.maxIdx > m.height-1 {
		m.maxIdx = m.minIdx + m.height - 1
	}
}

// Height returns the height of the file picker.
func (m Model) Height() int {
	return m.height
}

// Init initializes the file picker model.
func (m Model) Init() tea.Cmd {
	return m.readDir(m.CurrentDirectory, m.ShowHidden)
}

// Update handles user interactions within the file picker model.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case readDirMsg:
		if msg.id != m.id {
			break
		}
		m.files = msg.entries
		m.maxIdx = max(m.maxIdx, m.Height()-1)
	case tea.WindowSizeMsg:
		if m.AutoHeight {
			m.SetHeight(msg.Height - marginBottom)
		}
		m.maxIdx = m.Height() - 1
	case tea.KeyPressMsg:
		switch {
		case matches(msg, m.KeyMap.GoToTop):
			m.selected = 0
			m.minIdx = 0
			m.maxIdx = m.Height() - 1
		case matches(msg, m.KeyMap.GoToLast):
			m.selected = len(m.files) - 1
			m.minIdx = len(m.files) - m.Height()
			m.maxIdx = len(m.files) - 1
		case matches(msg, m.KeyMap.Down):
			m.selected++
			if m.selected >= len(m.files) {
				m.selected = len(m.files) - 1
			}
			if m.selected > m.maxIdx {
				m.minIdx++
				m.maxIdx++
			}
		case matches(msg, m.KeyMap.Up):
			m.selected--
			if m.selected < 0 {
				m.selected = 0
			}
			if m.selected < m.minIdx {
				m.minIdx--
				m.maxIdx--
			}
		case matches(msg, m.KeyMap.PageDown):
			m.selected += m.Height()
			if m.selected >= len(m.files) {
				m.selected = len(m.files) - 1
			}
			m.minIdx += m.Height()
			m.maxIdx += m.Height()

			if m.maxIdx >= len(m.files) {
				m.maxIdx = len(m.files) - 1
				m.minIdx = m.maxIdx - m.Height()
			}
		case matches(msg, m.KeyMap.PageUp):
			m.selected -= m.Height()
			if m.selected < 0 {
				m.selected = 0
			}
			m.minIdx -= m.Height()
			m.maxIdx -= m.Height()

			if m.minIdx < 0 {
				m.minIdx = 0
				m.maxIdx = m.minIdx + m.Height()
			}
		case matches(msg, m.KeyMap.Back):
			parent := filepath.Dir(m.CurrentDirectory)
			if parent == m.CurrentDirectory {
				break
			}
			m.CurrentDirectory = parent
			if m.selectedStack.Length() > 0 {
				m.selected, m.minIdx, m.maxIdx = m.popView()
			} else {
				m.selected = 0
				m.minIdx = 0
				m.maxIdx = m.Height() - 1
			}
			return m, m.readDir(m.CurrentDirectory, m.ShowHidden)
		case matches(msg, m.KeyMap.Open):
			if len(m.files) == 0 {
				break
			}

			f := m.files[m.selected]
			info, err := f.Info()
			if err != nil {
				break
			}
			isSymlink := info.Mode()&os.ModeSymlink != 0
			isDir := f.IsDir()

			if isSymlink {
				symlinkPath, _ := filepath.EvalSymlinks(filepath.Join(m.CurrentDirectory, f.Name()))
				info, err := os.Stat(symlinkPath)
				if err != nil {
					break
				}
				if info.IsDir() {
					isDir = true
				}
			}

			if (!isDir && m.FileAllowed) || (isDir && m.DirAllowed) {
				if matches(msg, m.KeyMap.Select) {
					// Select the current path as the selection
					m.Path = filepath.Join(m.CurrentDirectory, f.Name())
				}
			}

			if !isDir {
				break
			}

			if f.Name() == ".." {
				m.CurrentDirectory = filepath.Dir(m.CurrentDirectory)
				if m.selectedStack.Length() > 0 {
					m.selected, m.minIdx, m.maxIdx = m.popView()
				} else {
					m.selected = 0
					m.minIdx = 0
					m.maxIdx = m.Height() - 1
				}
			} else {
				m.CurrentDirectory = filepath.Join(m.CurrentDirectory, f.Name())
				m.pushView(m.selected, m.minIdx, m.maxIdx)
				m.selected = 0
				m.minIdx = 0
				m.maxIdx = m.Height() - 1
			}
			return m, m.readDir(m.CurrentDirectory, m.ShowHidden)
		}
	}
	return m, nil
}

// View returns the view of the file picker.
func (m Model) View() string {
	var s strings.Builder

	s.WriteString(m.Styles.Directory.Render(" Current Directory: "+m.CurrentDirectory) + "\n\n")

	if len(m.files) == 0 {
		s.WriteString(m.Styles.EmptyDirectory.String())
		// Pad remaining height
		for i := lipgloss.Height(s.String()); i <= m.Height(); i++ {
			s.WriteRune('\n')
		}
		return s.String()
	}

	for i, f := range m.files {
		if i < m.minIdx || i > m.maxIdx {
			continue
		}

		var symlinkPath string
		info, err := f.Info()
		if err != nil {
			continue
		}
		isSymlink := info.Mode()&os.ModeSymlink != 0
		size := strings.Replace(humanizeBytes(uint64(info.Size())), " ", "", 1)
		name := f.Name()

		if isSymlink {
			symlinkPath, _ = filepath.EvalSymlinks(filepath.Join(m.CurrentDirectory, name))
		}

		disabled := !m.canSelect(name) && !f.IsDir()

		if m.selected == i { //nolint:nestif
			selected := ""
			if m.ShowPermissions {
				selected += " " + info.Mode().String()
			}
			if m.ShowSize {
				selected += fmt.Sprintf("%"+strconv.Itoa(m.Styles.FileSize.GetWidth())+"s", size)
			}
			selected += " " + name
			if isSymlink {
				selected += " → " + symlinkPath
			}
			if disabled {
				s.WriteString(m.Styles.DisabledCursor.Render(m.Cursor) + m.Styles.DisabledSelected.Render(selected))
			} else {
				s.WriteString(m.Styles.Cursor.Render(m.Cursor) + m.Styles.Selected.Render(selected))
			}
			s.WriteRune('\n')
			continue
		}

		style := m.Styles.File
		if f.IsDir() {
			style = m.Styles.Directory
		} else if isSymlink {
			style = m.Styles.Symlink
		} else if disabled {
			style = m.Styles.DisabledFile
		}

		fileName := style.Render(name)
		s.WriteString(m.Styles.Cursor.Render(" "))
		if isSymlink {
			fileName += " → " + symlinkPath
		}
		if m.ShowPermissions {
			s.WriteString(" " + m.Styles.Permission.Render(info.Mode().String()))
		}
		if m.ShowSize {
			s.WriteString(m.Styles.FileSize.Render(size))
		}
		s.WriteString(" " + fileName)
		s.WriteRune('\n')
	}

	for i := lipgloss.Height(s.String()); i <= m.Height(); i++ {
		s.WriteRune('\n')
	}

	return s.String()
}

// DidSelectFile returns whether a user has selected a file (on this msg).
func (m Model) DidSelectFile(msg tea.Msg) (bool, string) {
	didSelect, path := m.didSelectFile(msg)
	if didSelect && m.canSelect(path) {
		return true, path
	}
	return false, ""
}

// DidSelectDisabledFile returns whether a user tried to select a disabled file
// (on this msg). This is necessary only if you would like to warn the user that
// they tried to select a disabled file.
func (m Model) DidSelectDisabledFile(msg tea.Msg) (bool, string) {
	didSelect, path := m.didSelectFile(msg)
	if didSelect && !m.canSelect(path) {
		return true, path
	}
	return false, ""
}

func (m Model) didSelectFile(msg tea.Msg) (bool, string) {
	if len(m.files) == 0 {
		return false, ""
	}
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		// If the msg was not a selection, returning early
		f := m.files[m.selected]
		info, err := f.Info()
		if err != nil {
			return false, ""
		}
		isSymlink := info.Mode()&os.ModeSymlink != 0
		isDir := f.IsDir()

		if isSymlink {
			symlinkPath, _ := filepath.EvalSymlinks(filepath.Join(m.CurrentDirectory, f.Name()))
			info, err := os.Stat(symlinkPath)
			if err != nil {
				return false, ""
			}
			if info.IsDir() {
				isDir = true
			}
		}

		if matches(msg, m.KeyMap.Select) {
			if (!isDir && m.FileAllowed) || (isDir && m.DirAllowed) {
				return true, filepath.Join(m.CurrentDirectory, f.Name())
			}
		}

	default:
		return false, ""
	}
	return false, ""
}

func (m Model) canSelect(file string) bool {
	if len(m.AllowedTypes) <= 0 {
		return true
	}

	for _, ext := range m.AllowedTypes {
		if strings.HasSuffix(file, ext) {
			return true
		}
	}
	return false
}

// HighlightedPath returns the path of the currently highlighted file or directory.
func (m Model) HighlightedPath() string {
	if len(m.files) == 0 || m.selected < 0 || m.selected >= len(m.files) {
		return ""
	}
	return filepath.Join(m.CurrentDirectory, m.files[m.selected].Name())
}

type parentDir struct{}

func (p parentDir) Name() string               { return ".." }
func (p parentDir) IsDir() bool                { return true }
func (p parentDir) Type() os.FileMode          { return os.ModeDir }
func (p parentDir) Info() (os.FileInfo, error) { return parentFileInfo{}, nil }

type parentFileInfo struct{}

func (p parentFileInfo) Name() string       { return ".." }
func (p parentFileInfo) Size() int64        { return 0 }
func (p parentFileInfo) Mode() os.FileMode  { return os.ModeDir }
func (p parentFileInfo) ModTime() time.Time { return time.Time{} }
func (p parentFileInfo) IsDir() bool        { return true }
func (p parentFileInfo) Sys() interface{}   { return nil }
