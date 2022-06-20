package ui

import (
	"fmt"
	"image/color"
	"log"
	"os"
	"strings"
	"time"

	"go_fyne_markdown_upload/db"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type uiElements struct {
	promptForTitleInput,
	promptForTagInput,
	promptForUpload,
	titleDuplicateWarning,
	fileExtensionWarning,
	fileReadErrorWarning,
	givenTitle,
	selectedFileName,
	tagsTitle, 
	uploadConfirmation *canvas.Text
	titleInput, 
	tagInput *widget.Entry
	logoImg *canvas.Image
	infoDialog dialog.Dialog
	fileExplorer *dialog.FileDialog
	selectFileBtn, submitBtn *widget.Button
	tagContainer *fyne.Container
	uploadProgressBar *widget.ProgressBarInfinite
}

var uploadApp fyne.App
var appWindow fyne.Window
var appUi uiElements
var filePath string
var fileTags []string
var markdownDocument db.NoteFile
var submitInfo string

func handleFileSelect(file fyne.URIReadCloser, err error) {
	if file != nil {
	fileInfo := file.URI()
	filePath = fileInfo.Path()
	fileName := fileInfo.Name()
	fileExtension := fileInfo.Extension()
	fmt.Printf("Filepath: %s\n",filePath)
	fmt.Printf("File extension: %s\n",fileExtension)

	if fileExtension != ".md" {
		appUi.fileExtensionWarning.Show()
		file.Close()
		return
	}

	// Hide error(s) if was shown from earlier file selections / file-reads
	appUi.fileExtensionWarning.Hidden = true
	appUi.fileReadErrorWarning.Hidden = true

	// Set text of element & show
	appUi.selectedFileName.Text = "File: " + fileName
	appUi.selectedFileName.Refresh()
	appUi.selectedFileName.Show()

	// Show prompt and input for giving title to the file to be uploaded
	appUi.promptForTitleInput.Show()
	appUi.titleInput.Show()
	appWindow.Canvas().Focus(appUi.titleInput)

	} else {
		return
	}
}

func handleTitleSubmit(title string) {
	if title == "" {
			appUi.givenTitle.Hidden = true
	} else {
		// Check for duplicate title from collection
		count, err := db.CheckForDuplicateNote(title)
		if err != nil {
			log.Println(err)
			appUi.titleDuplicateWarning.Text = "Error with making query to database."
			appUi.titleDuplicateWarning.Show()
		} else if count > 0 {
			log.Println("Document with given title already exists.")
			// Hide if previously was shown
			appUi.givenTitle.Hidden = true
			appUi.titleInput.SetText("")
			appUi.titleDuplicateWarning.Show()
			time.Sleep(time.Second * 2)
			appUi.titleDuplicateWarning.Hide()
		} else {
			appUi.givenTitle.Text = "Title: "+ title
			markdownDocument.Title = title
			appUi.titleInput.SetText("")
			appUi.titleInput.SetPlaceHolder("Rename the title")
			appUi.givenTitle.Refresh()
			appUi.givenTitle.Show()
			// Show next step for tag input
			appUi.promptForTagInput.Show()
			appUi.tagInput.Show()
			appWindow.Canvas().Focus(appUi.tagInput)
		}
	}
}

func handleTagSubmit(tag string) {
	// Format string input
	tag = strings.ToTitle(tag)
	
	fileTags = append(fileTags, tag)
	if len(fileTags) > 0 {
		appUi.tagsTitle.Show()
		appUi.promptForUpload.Show()
		appUi.submitBtn.Show()
	}
	// Tag element
	tagElement := widget.NewButtonWithIcon(tag, theme.CancelIcon(), nil)
	tagElement.IconPlacement = widget.ButtonIconTrailingText
	// Tag button event handler - Remove tag
	tagElement.OnTapped = func() {
		tagElement.Hide()
		// Remove tag from fileTags:
		var tagIndex int
		for index, item := range fileTags {
			if tag == item {
				tagIndex = index
				break
			}
		}
		fileTags = append(fileTags[:tagIndex], fileTags[tagIndex+1:]...)
		if len(fileTags) < 1 {
			// Hide next step for uploading the file
			appUi.tagsTitle.Hidden = true
			appUi.promptForUpload.Hidden = true
			appUi.submitBtn.Hidden = true
		}
	}
	appUi.tagContainer.Add(tagElement)
	// Reset text input
	appUi.tagInput.SetText("")
}

func openSubmitDialog() {
	// Store input tags to struct field before confirming with submission
	markdownDocument.Tags = fileTags
	// Text to show within dialog element
	submitInfo = "\n"+ "Filepath: "+ filePath + "\n\n" + "Title: " + markdownDocument.Title + "\n\n" + "Tags: " + strings.Join(markdownDocument.Tags, ", ") + "\n"
	
	// Dialog element
	appUi.infoDialog = dialog.NewInformation("Data to be submitted", submitInfo, appWindow)
	appUi.infoDialog.Resize(fyne.NewSize(600,200))
	appUi.infoDialog.SetDismissText("Submit")
	appUi.infoDialog.SetOnClosed(submitData)
	appUi.infoDialog.Show()	
}

func submitData() {
	// Read and save file
	fileBuffer, err := os.ReadFile(filePath)
	if err != nil {
		appUi.fileReadErrorWarning.Text = "Something went wrong with file read:\n" + err.Error()
		appUi.fileReadErrorWarning.Show()
	}
	markdownDocument.File = fileBuffer

	// Render upload view
	setUploadContent()
	appWindow.Content().Refresh()

	uploadFinished := db.UploadNote(markdownDocument)
	if uploadFinished {
		time.Sleep(time.Second * 2)
		appUi.uploadProgressBar.Hidden = true
		appUi.uploadConfirmation.Show()
		// Render initial app view for new file upload
		time.Sleep(time.Second * 2)
		createInitialElements()
		setInitialContent()
		appWindow.Content().Refresh()
		log.Println("Re-rendered initial app view for new file upload")
	} else {
		time.Sleep(time.Second * 2)
		appUi.uploadConfirmation.Text = "Upload failed, try again."
		appUi.uploadConfirmation.Show()
		time.Sleep(time.Second * 2)
		setInitialContent()
		appWindow.Content().Refresh()
	}
}

func createInitialElements() {
	// App logo
	appLogo, appIconErr := fyne.LoadResourceFromURLString("https://ik.imagekit.io/htg3gsxgz/Markdown-upload-app/go-gopher.png?ik-sdk-version=javascript-1.4.3&updatedAt=1655656772973")
	if appIconErr != nil {
		log.Printf("Error with app logo load: %s\n", appIconErr)
	}
	// App logo element
	appUi.logoImg = canvas.NewImageFromResource(appLogo)
	appUi.logoImg.SetMinSize(fyne.Size{Width:200, Height: 150})
	appUi.logoImg.FillMode = canvas.ImageFillContain

	// File explorer dialog
	appUi.fileExplorer = dialog.NewFileOpen(handleFileSelect, appWindow)

	// Select file button
	appUi.selectFileBtn = widget.NewButton("Open file", func() {
		appUi.fileExplorer.Show()
	})

	appUi.selectFileBtn.SetIcon(theme.FileIcon())
	// Filename
	appUi.selectedFileName = canvas.NewText("", color.White)
	appUi.selectedFileName.TextSize = 18
	appUi.selectedFileName.Hidden = true
	// File extension warning - Invalid document type for upload
	appUi.fileExtensionWarning = canvas.NewText("Only markdown files allowed, '.md' extension. Please, try uploading a file again", color.White)
	appUi.fileExtensionWarning.Alignment = fyne.TextAlignCenter
	appUi.fileExtensionWarning.Hidden = true
	// File information read error
	appUi.fileReadErrorWarning = canvas.NewText("", color.White)
	appUi.fileReadErrorWarning.Alignment = fyne.TextAlignCenter
	appUi.fileReadErrorWarning.Hidden = true

	// Title prompt
	appUi.promptForTitleInput = canvas.NewText("Provide title for the file", color.White)
	appUi.promptForTitleInput.Hidden = true
	appUi.promptForTitleInput.Alignment = fyne.TextAlignCenter
	// Title for file
	appUi.givenTitle = canvas.NewText("", color.White)
	appUi.givenTitle.TextSize = 18
	appUi.givenTitle.Hidden = true
	// Title input
	appUi.titleInput = widget.NewEntry()
	appUi.titleInput.Hidden = true
	appUi.titleInput.SetPlaceHolder("Provide title for the file")
	// Title input handler
	appUi.titleInput.OnSubmitted = handleTitleSubmit

	// Title warning - Duplicate found from DB
	appUi.titleDuplicateWarning = canvas.NewText("Duplicate title found from collection. Give a new title for document", color.White)
	appUi.titleDuplicateWarning.Alignment = fyne.TextAlignCenter
	appUi.titleDuplicateWarning.Hidden = true

	// Tag prompt
	appUi.promptForTagInput = canvas.NewText("Provide tags for the document", color.White)
	appUi.promptForTagInput.Alignment = fyne.TextAlignCenter
	appUi.promptForTagInput.Hidden = true
	// Tag input
	appUi.tagInput = widget.NewEntry()
	appUi.tagInput.SetPlaceHolder("Write tag and press enter to insert another tag")
	appUi.tagInput.Hidden = true
	// Handler for tag submit
	appUi.tagInput.OnSubmitted = handleTagSubmit

	// Title text for tags
	appUi.tagsTitle = canvas.NewText("Tags: ", color.White)
	appUi.tagsTitle.TextSize = 18
	appUi.tagsTitle.Hidden = true
	// Tags container, tagsTitle + tags
	appUi.tagContainer = container.NewHBox()
	appUi.tagContainer.Add(appUi.tagsTitle)

	// Submit prompt
	appUi.promptForUpload = canvas.NewText("Submit to database", color.White)
	appUi.promptForUpload.Alignment = fyne.TextAlignCenter
	appUi.promptForUpload.Hidden = true

	// Submit button
	appUi.submitBtn = widget.NewButton("Submit", nil)
	appUi.submitBtn.SetIcon(theme.MailForwardIcon())
	appUi.submitBtn.Hidden = true
	appUi.submitBtn.OnTapped = openSubmitDialog

	// Dialog to show before submission of data
	appUi.infoDialog = dialog.NewInformation("Data to be submitted", submitInfo, appWindow)
	appUi.infoDialog.Resize(fyne.NewSize(600,200))
	appUi.infoDialog.SetDismissText("Submit")
}

func createUploadElements() {
	// Upload related ui content
	appUi.uploadProgressBar = widget.NewProgressBarInfinite()
	appUi.uploadConfirmation = canvas.NewText("Upload completed!", color.White)
	appUi.uploadConfirmation.Alignment = fyne.TextAlignCenter
	appUi.uploadConfirmation.TextSize = 20
	appUi.uploadConfirmation.Hidden = true
}

func setUploadContent() {
	uploadContent := container.New(layout.NewVBoxLayout(), layout.NewSpacer(),appUi.uploadProgressBar, appUi.uploadConfirmation,layout.NewSpacer())

	appWindow.SetContent(uploadContent)
	appWindow.Content().Refresh()
}

func setInitialContent() {
	// Initial app elements
	initialUiElements := []fyne.CanvasObject{
		layout.NewSpacer(),
		appUi.logoImg,
		appUi.selectFileBtn,
		appUi.fileExtensionWarning,
		layout.NewSpacer(),
		appUi.selectedFileName,
		layout.NewSpacer(),
		appUi.promptForTitleInput,
		appUi.titleInput,
		layout.NewSpacer(),
		appUi.titleDuplicateWarning,
		appUi.givenTitle,
		layout.NewSpacer(),
		appUi.promptForTagInput, 
		appUi.tagInput,
		layout.NewSpacer(),
		appUi.tagContainer,
		layout.NewSpacer(), 
		layout.NewSpacer(), 
		appUi.promptForUpload, 
		appUi.submitBtn, 
		layout.NewSpacer(),
	}

	appWindow.SetContent(container.New(layout.NewVBoxLayout(), initialUiElements...))
	appWindow.Content().Refresh()
}

func init() {
	// App icon
	appIcon, iconLoadErr := fyne.LoadResourceFromURLString("https://ik.imagekit.io/htg3gsxgz/Markdown-upload-app/icon-leaf-mongo.png?ik-sdk-version=javascript-1.4.3&updatedAt=1655656760971")

	// Initialize Fyne app
	uploadApp = app.New()
	uploadApp.Settings().SetTheme(myTheme{})
	appWindow = uploadApp.NewWindow("Upload markdown file")
	// Default window size
	appWindow.Resize(fyne.NewSize(800, 700))
	// Set the window icon if successfully fetched
	if iconLoadErr == nil {
		appWindow.SetIcon(appIcon)
	}

	createInitialElements()
	createUploadElements()
	setInitialContent()
}

func tidyUp() {
	log.Println("Exited the app.")
}

func Start() {
	appWindow.ShowAndRun()
	tidyUp()
}
