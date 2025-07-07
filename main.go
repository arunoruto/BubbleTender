package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// --- DATA ---
type Beverage struct {
	Name  string
	Price float64
	Stock int
}

var ourBeverages = []Beverage{
	{Name: "Club-Mate", Price: 1.50, Stock: 24},
	{Name: "Espresso", Price: 1.00, Stock: 50},
	{Name: "Fritz-Kola", Price: 2.00, Stock: 12},
	{Name: "Water", Price: 0.50, Stock: 100},
	{Name: "Beer", Price: 2.50, Stock: 6},
}

func tabBorderWithBottom(left, middle, right string) lipgloss.Border {
	border := lipgloss.RoundedBorder()
	border.BottomLeft = left
	border.Bottom = middle
	border.BottomRight = right
	return border
}

var (
	inactiveTabBorder = tabBorderWithBottom("┴", "─", "┴")
	activeTabBorder   = tabBorderWithBottom("┘", " ", "└")
	docStyle          = lipgloss.NewStyle().Padding(1, 2, 1, 2)
	highlightColor    = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	inactiveTabStyle  = lipgloss.NewStyle().Border(inactiveTabBorder, true).BorderForeground(highlightColor).Padding(0, 1)
	activeTabStyle    = inactiveTabStyle.Border(activeTabBorder, true)
	windowStyle       = lipgloss.NewStyle().BorderForeground(highlightColor).Padding(2, 0).Align(lipgloss.Center).Border(lipgloss.NormalBorder()).UnsetBorderTop()
)

// --- MODEL ---

type model struct {
	beverages     []Beverage
	table         table.Model
	cart          map[int]int
	isCheckingOut bool
	activeTab     int
	width         int
	height        int
}

func initialModel() model {
	columns := []table.Column{
		{Title: "Name", Width: 20},
		{Title: "Price", Width: 10},
		{Title: "Stock", Width: 10},
		{Title: "Qty", Width: 5},
	}
	cart := make(map[int]int)
	rows := []table.Row{}
	for i, beverage := range ourBeverages {
		row := table.Row{
			beverage.Name,
			fmt.Sprintf("€%.2f", beverage.Price),
			fmt.Sprintf("%d", beverage.Stock),
			fmt.Sprintf("%d", cart[i]),
		}
		rows = append(rows, row)
	}
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(7),
	)
	s := table.DefaultStyles()
	s.Header = s.Header.BorderStyle(lipgloss.NormalBorder()).BorderBottom(true)
	s.Selected = s.Selected.Foreground(lipgloss.Color("229")).Background(lipgloss.Color("57")).Bold(false)
	t.SetStyles(s)

	return model{
		beverages:     ourBeverages,
		table:         t,
		cart:          cart,
		isCheckingOut: false,
		activeTab:     0,
	}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c", "q":
			return m, tea.Quit
		}

		switch keypress := msg.String(); keypress {
		case "s":
			m.activeTab = 0 // Shop
			m.isCheckingOut = false
		case "c":
			m.activeTab = 1 // Cart
			m.isCheckingOut = false
		}

		switch m.activeTab {
		case 0: // Shop Tab
			switch msg.String() {
			case "+", "=", "right":
				cursor := m.table.Cursor()
				if m.cart[cursor] < m.beverages[cursor].Stock {
					m.cart[cursor]++
				}
			case "-", "left":
				cursor := m.table.Cursor()
				if m.cart[cursor] > 0 {
					m.cart[cursor]--
				}
			}
			rows := []table.Row{}
			for i, beverage := range m.beverages {
				row := table.Row{
					beverage.Name,
					fmt.Sprintf("€%.2f", beverage.Price),
					fmt.Sprintf("%d", beverage.Stock),
					fmt.Sprintf("%d", m.cart[i]),
				}
				rows = append(rows, row)
			}
			m.table.SetRows(rows)
			m.table, cmd = m.table.Update(msg)

		case 1: // Cart Tab
			if m.isCheckingOut {
				switch msg.String() {
				case "y":
					return m, tea.Quit
				case "n", "esc":
					m.isCheckingOut = false
				}
			} else {
				if msg.String() == "enter" {
					hasItems := false
					for _, qty := range m.cart {
						if qty > 0 {
							hasItems = true
							break
						}
					}
					if hasItems {
						m.isCheckingOut = true
					}
				}
			}
		}
	}

	return m, cmd
}

// --- VIEWS ---

func (m model) View() string {
	var mainContent string
	var helpText string

	// --- 1. Generate the Main Content String ---
	switch m.activeTab {
	case 1: // Cart
		mainContent = m.cartView()
	default: // Shop
		mainContent = m.table.View()
		helpText = "\n\nUse ←/→ to change quantity.\nPress 'c' to view cart, 'q' to quit."
	}

	// Render the content inside its styled window
	renderedContent := windowStyle.Render(mainContent + helpText)

	// --- 2. Measure the Content's Width ---
	contentWidth := lipgloss.Width(renderedContent)

	// --- 3. Render the Tabs to Match the Width ---
	tabs := []string{"Shop [s]", "Cart [c]"}
	renderedTabs := []string{}

	// Create styled tab strings
	for i, t := range tabs {
		var style lipgloss.Style
		isFirst, isLast, isActive := i == 0, i == len(tabs)-1, i == m.activeTab
		if isActive {
			style = activeTabStyle
		} else {
			style = inactiveTabStyle
		}
		border, _, _, _, _ := style.GetBorder()
		if isFirst && isActive {
			border.BottomLeft = "│"
		} else if isFirst && !isActive {
			border.BottomLeft = "├"
		} else if isLast && isActive {
			border.BottomRight = "│"
		} else if isLast && !isActive {
			border.BottomRight = "┤"
		}
		style = style.Border(border)
		renderedTabs = append(renderedTabs, style.Render(t))
	}

	// Calculate the width of the tabs and create a filler
	tabsWidth := lipgloss.Width(renderedTabs[0]) + lipgloss.Width(renderedTabs[1])
	fillerWidth := contentWidth - tabsWidth

	// Create a style for the filler that only has a bottom border
	fillerStyle := lipgloss.NewStyle().
		BorderStyle(inactiveTabBorder).
		BorderBottom(true).
		BorderForeground(highlightColor).
		Width(fillerWidth)

	// Join the tabs and filler
	tabsRow := lipgloss.JoinHorizontal(lipgloss.Bottom, renderedTabs[0], fillerStyle.Render(""), renderedTabs[1])

	// --- 4. Combine and Center ---
	finalView := lipgloss.JoinVertical(lipgloss.Left, tabsRow, renderedContent)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		finalView,
	)
}

func (m model) cartView() string {
	var s strings.Builder
	s.WriteString("Your Current Order:\n\n")

	totalPrice := 0.0
	hasItems := false
	for i, quantity := range m.cart {
		if quantity > 0 {
			hasItems = true
			beverage := m.beverages[i]
			itemPrice := beverage.Price * float64(quantity)
			totalPrice += itemPrice
			s.WriteString(fmt.Sprintf("  %dx %-20s @ €%.2f each = €%.2f\n",
				quantity, beverage.Name, beverage.Price, itemPrice))
		}
	}

	if !hasItems {
		s.WriteString("  Your cart is empty!\n\n\nGo to the 'Shop' tab to add items.")
	} else {
		s.WriteString("\n  -------------------------------------------\n")
		s.WriteString(fmt.Sprintf("  Total: €%.2f\n", totalPrice))
		if m.isCheckingOut {
			s.WriteString("\n\nConfirm purchase? (y/n)\n(Press 'esc' or 'n' to cancel checkout)")
		} else {
			s.WriteString("\n\nPress 'enter' to checkout.")
		}
	}
	return s.String()
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
