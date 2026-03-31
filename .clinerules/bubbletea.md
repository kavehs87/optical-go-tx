# Bubble Tea Framework - Developer Guidelines & Patterns

## Quick Navigation
- **API Documentation**: [Charm.land Bubble Tea v2 Docs](https://charm.land/bubbletea/v2)
- **Examples Repository**: [Local Bubble Tea Examples](file:///Users/kaveh/Documents/Projects/optical-transfer-cli/bubbletea-repository/examples)
- **Bubbles Component Library**: [Local Bubbles Repository](file:///Users/kaveh/Documents/Projects/optical-transfer-cli/bubbles-repository)
- **Lip Gloss Styling**: [Local Lip Gloss Repository](file:///Users/kaveh/Documents/Projects/optical-transfer-cli/libgloss-repository)

---

## Core Architecture (MVU Pattern)

Bubble Tea is based on The Elm Architecture with **Model-View-Update (MVU)**:

### Model
```go
type Model interface {
    Init() tea.Cmd
    Update(msg tea.Msg) (Model, tea.Cmd)
    View() tea.View  // v2 returns tea.View struct
}
```

### Commands (I/O Operations)
```go
type Cmd func() Msg

// Example command pattern
func loadData() tea.Cmd {
    return func() Msg {
        data, err := performOperation()
        if err != nil {
            return errMsg{err: err}
        }
        return dataMsg{data: data}
    }
}
```

### Messages (Events)
```go
type Msg interface{}

// Common messages
- tea.QuitMsg
- tea.KeyPressMsg  
- tea.WindowSizeMsg
- Custom I/O messages
```

---

## Quick API Reference

### Key Types & Methods

| Type | Documentation Link | Common Use |
|------|-------------------|------------|
| **tea.Model** | [Model Interface](https://charm.land/bubbletea/v2/docs#Model) | Application state wrapper |
| **tea.Cmd** | [Cmd Type](https://charm.land/bubbletea/v2/docs#Cmd) | Asynchronous operations |
| **tea.Msg** | [Msg Interface](https://charm.land/bubbletea/v2/docs#Msg) | Communication events |
| **tea.View** | [View Struct](https://charm.land/bubbletea/v2/docs#View) | v2 declarative rendering |
| **tea.Quit** | [tea.Quit](https://charm.land/bubbletea/v2/docs#Quit) | Exit program |

### View Method (v2 Pattern)

```go
func (m model) View() tea.View {
    var v tea.View
    v.AltScreen = true
    v.Content = renderUI()
    return v
}
```

### Key Message Handling

```go
func (m model) Update(msg tea.Msg) (Model, Cmd) {
    switch msg := msg.(type) {
    case tea.KeyPressMsg:
        if msg.String() == "q" || msg.String() == "ctrl+c" {
            return m, tea.Quit
        }
        // Handle navigation, actions, etc.
    case tea.WindowSizeMsg:
        // Handle terminal resize
    default:
        // Custom message handling
    }
    return m, nil
}
```

### Initialization

```go
func (m model) Init() tea.Cmd {
    // Return initial command or nil
    return nil
}

func main() {
    p := tea.NewProgram(m)
    if _, err := p.Run(); err != nil {
        fmt.Printf("Error: %v", err)
    }
}
```

---

## Common Patterns

### 1. Async I/O with Commands
```go
func fetchData(url string) tea.Cmd {
    return func() tea.Msg {
        resp, err := http.Get(url)
        if err != nil {
            return errMsg{err: err}
        }
        return dataMsg{data: resp}
    }
}

// In Update
case dataMsg:
    m.data = msg.data
    return m, nil
```

### 2. Error Handling as Message
```go
type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }

// In Update
case errMsg:
    m.err = msg.err.Error()
    return m, tea.Quit
```

### 3. Key Press Navigation
```go
switch msg.Text {
    case "j", "\n":
        m.cursor++
    case "k":
        m.cursor--
    case " ":
        m.selected[m.cursor] = struct{}{}
    default:
        // Handle other keys
}
```

### 4. Full Screen (AltScreen)
```go
func (m model) View() tea.View {
    var v tea.View
    v.AltScreen = true
    v.Content = "Your UI here"
    return v
}
```

---

## Debugging Resources

### Logging to File
```go
if os.Getenv("DEBUG") != "" {
    f, _ := tea.LogToFile("debug.log", "debug")
    defer f.Close()
}
```

### Headless Delve Debugging
```bash
# Terminal 1
dlv debug --headless --api-version=2 --listen=127.0.0.1:43000 .

# Terminal 2  
dlv connect 127.0.0.1:43000
```

---

## Local Repository
All related repositories are cloned locally:

- **Bubble Tea**: `/Users/kaveh/Documents/Projects/optical-transfer-cli/bubbletea-repository`
- **Bubbles Components**: `/Users/kaveh/Documents/Projects/optical-transfer-cli/bubbles-repository`
- **Lip Gloss**: `/Users/kaveh/Documents/Projects/optical-transfer-cli/libgloss-repository`

Access package and examples:
```bash
# Get package
go get github.com/charmbracelet/bubbletea@latest

# View examples
open /Users/kaveh/Documents/Projects/optical-transfer-cli/bubbletea-repository/examples

# View Bubbles components
open /Users/kaveh/Documents/Projects/optical-transfer-cli/bubbles-repository

# View Lip Gloss styling
open /Users/kaveh/Documents/Projects/optical-transfer-cli/libgloss-repository
```

---

## Project Structure Template
```
optical-transfer-cli/
├── go.mod              # Dependencies
├── main.go             # Entry point (tea.NewProgram)
├── models.go           # Model definition & methods
├── update.go           # Update implementations
├── view.go             # View implementations  
├── commands.go         # Cmd definitions
└── types.go            # Custom Msg and Cmd types
```

---

## For Specific API Questions
Copy-paste these links for detailed API documentation:
- **Full API Reference**: [https://charm.land/bubbletea/v2/docs](https://charm.land/bubbletea/v2/docs)
- **Package Types**: [https://pkg.go.dev/github.com/charmbracelet/bubbletea/v2](https://pkg.go.dev/github.com/charmbracelet/bubbletea/v2)

---

*Updated for Bubble Tea v2 with quick API access links and local repository reference. Based on official Charm documentation.*