package main

import (
	"go_fyne_markdown_upload/ui"
	"go_fyne_markdown_upload/db"
)

func main() {
	// Initialize DB connection
	db.Initialize()
	defer db.CloseDb()
	// Start app
	ui.Start()
}
