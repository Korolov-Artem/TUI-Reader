# GoReader 📖

> A terminal-based EPUB reader written in Go.
> Built with the [Bubble Tea](https://github.com/charmbracelet/bubbletea) framework.

## 🚀 Overview

**GoReader** is a TUI (Terminal User Interface) application that allows you to read EPUB books directly in your terminal. It features a file browser library, chapter navigation, and a responsive reading view that adapts to your terminal window size.

I built this project to master State Management in Go and to explore the Elm Architecture pattern applied to CLI tools.

## ✨ Features

* **📚 Library System:** Scans your current directory for `.epub` files and presents a browsable menu.
* **🧠 Smart State Management:** Uses the Model-View-Update (MVU) pattern to handle application state seamlessly.
* **⚡ Vim-Style Navigation:** Supports `j/k` for menu navigation and standard arrow keys for reading.
* **📖 Responsive Reading:**
    * Automatic word wrapping based on terminal width.
    * Pagination logic that recalculates instantly on window resize.
    * Clean, distraction-free UI with [Lipgloss](https://github.com/charmbracelet/lipgloss) styling.
* **🛠️ Robust Parsing:** Custom implementation for unzipping EPUB containers and parsing XML content.

## 📦 Installation

Ensure you have Go installed (1.19+).

``` bash
# Clone the repository
git clone [https://github.com/yourusername/goreader.git](https://github.com/yourusername/goreader.git)

# Navigate to the directory
cd goreader

# Run the application
go run .
