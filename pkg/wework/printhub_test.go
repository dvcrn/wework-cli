package wework

import (
	"mime"
	"mime/multipart"
	"strings"
	"testing"
)

func TestBuildPrintJobMultipartValidation(t *testing.T) {
	if _, _, err := buildPrintJobMultipart(&AddPrintJobRequest{FileName: "doc.pdf"}); err == nil {
		t.Fatal("expected error when file bytes are missing")
	}
	if _, _, err := buildPrintJobMultipart(&AddPrintJobRequest{FileBytes: []byte("data")}); err == nil {
		t.Fatal("expected error when file name is missing")
	}
}

func TestBuildPrintJobMultipartDefaults(t *testing.T) {
	input := AddPrintJobRequest{
		FileName:  "resume.pdf",
		FileBytes: []byte("%PDF-1.4 fake"),
	}
	body, contentType, err := buildPrintJobMultipart(&input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Defaults are applied on the passed-in request.
	if input.Copies != 1 {
		t.Errorf("Copies = %d, want 1", input.Copies)
	}
	if input.OrientationRequested != "portrait" {
		t.Errorf("OrientationRequested = %q, want portrait", input.OrientationRequested)
	}
	if input.PrintColorMode != "monochrome" {
		t.Errorf("PrintColorMode = %q, want monochrome", input.PrintColorMode)
	}
	if input.Sides != "one-sided" {
		t.Errorf("Sides = %q, want one-sided", input.Sides)
	}
	if input.JobName != "resume.pdf" {
		t.Errorf("JobName = %q, want resume.pdf (falls back to file name)", input.JobName)
	}
	if input.FileContentType != "application/octet-stream" {
		t.Errorf("FileContentType = %q, want application/octet-stream", input.FileContentType)
	}

	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		t.Fatalf("failed to parse content type %q: %v", contentType, err)
	}
	if mediaType != "multipart/form-data" {
		t.Errorf("media type = %q, want multipart/form-data", mediaType)
	}

	reader := multipart.NewReader(body, params["boundary"])
	form, err := reader.ReadForm(1 << 20)
	if err != nil {
		t.Fatalf("failed to read multipart form: %v", err)
	}
	if got := form.Value["jobName"]; len(got) != 1 || got[0] != "resume.pdf" {
		t.Errorf("jobName field = %v, want [resume.pdf]", got)
	}
	if got := form.Value["copies"]; len(got) != 1 || got[0] != "1" {
		t.Errorf("copies field = %v, want [1]", got)
	}
	files := form.File["file"]
	if len(files) != 1 {
		t.Fatalf("expected 1 file part, got %d", len(files))
	}
	if files[0].Filename != "resume.pdf" {
		t.Errorf("file part filename = %q, want resume.pdf", files[0].Filename)
	}
}

func TestBuildPrintJobMultipartHonorsExplicitValues(t *testing.T) {
	input := AddPrintJobRequest{
		FileName:             "poster.pdf",
		FileBytes:            []byte("data"),
		Copies:               3,
		OrientationRequested: "landscape",
		PrintColorMode:       "color",
		Sides:                "two-sided-long-edge",
		JobName:              "Poster",
		FileContentType:      "application/pdf",
	}
	if _, _, err := buildPrintJobMultipart(&input); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if input.Copies != 3 || input.PrintColorMode != "color" || input.JobName != "Poster" {
		t.Errorf("explicit values were overwritten: %+v", input)
	}
	if !strings.Contains(input.Sides, "two-sided") {
		t.Errorf("Sides = %q, want to keep explicit value", input.Sides)
	}
}
