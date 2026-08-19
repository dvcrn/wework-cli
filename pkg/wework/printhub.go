package wework

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"strconv"
	"strings"
)

// PrintQueueResponse is the paginated payload returned by the print-hub queue
// endpoint.
type PrintQueueResponse struct {
	Content []PrintJob `json:"content"`
	Page    struct {
		TotalElements int `json:"totalElements"`
	} `json:"page"`
}

// PrintJob describes a single job in the member's print queue.
type PrintJob struct {
	CreatedTime                    CustomTime `json:"createdTime"`
	LastUpdatedTime                CustomTime `json:"lastUpdatedTime"`
	InternalJobStatus              string     `json:"internalJobStatus"`
	Status                         string     `json:"status"`
	PagesInJob                     int        `json:"pagesInJob"`
	JobName                        string     `json:"jobName"`
	Collated                       bool       `json:"collated"`
	Copies                         int        `json:"copies"`
	ForceMediaSize                 any        `json:"forceMediaSize"`
	MediaSizes                     []string   `json:"mediaSizes"`
	OrientationRequested           string     `json:"orientationRequested"`
	PageDelivery                   any        `json:"pageDelivery"`
	JobURI                         any        `json:"jobUri"`
	PrintColorMode                 string     `json:"printColorMode"`
	PrinterResolution              any        `json:"printerResolution"`
	PrintQuality                   string     `json:"printQuality"`
	Sides                          string     `json:"sides"`
	PrintScaling                   any        `json:"printScaling"`
	JobState                       any        `json:"jobState"`
	ImpressionsCompleted           int        `json:"impressionsCompleted"`
	MediaSheetsCompleted           int        `json:"mediaSheetsCompleted"`
	UserID                         string     `json:"userId"`
	PrinterID                      any        `json:"printerId"`
	IsSubmittedDocumentPreRendered bool       `json:"isSubmittedDocumentPreRendered"`
	SubmittedDocumentMimeType      string     `json:"submittedDocumentMimeType"`
	RenderedDocumentsRequested     int        `json:"renderedDocumentsRequested"`
	RenderedDocuments              any        `json:"renderedDocuments"`
	ExpiryTime                     CustomTime `json:"expiryTime"`
	AirPrintProfileID              any        `json:"airPrintProfileId"`
	UploadDocuments                any        `json:"uploadDocuments"`
	S3Region                       string     `json:"s3Region"`
	QueuedForLaterRelease          bool       `json:"queuedForLaterRelease"`
	NonPDFRenderingSupport         bool       `json:"nonPdfRenderingSupport"`
	UserAccount                    any        `json:"userAccount"`
	ReleaseIntent                  string     `json:"releaseIntent"`
	ReleasePrinterID               any        `json:"releasePrinterId"`
	MediaCol                       any        `json:"mediaCol"`
	PrinterName                    any        `json:"printerName"`
	StratusResourceInfo            any        `json:"stratusResourceInfo"`
	Borderless                     bool       `json:"borderless"`
	MultipleDocumentHandling       any        `json:"multipleDocumentHandling"`
	NumberUp                       any        `json:"numberUp"`
	HPPCVersionMajor               any        `json:"hpPcVersionMajor"`
	HPPCVersionMinor               any        `json:"hpPcVersionMinor"`
	HPPrintQuality                 any        `json:"hpPrintQuality"`
	JobOrigin                      any        `json:"jobOrigin"`
	ID                             string     `json:"id"`
	IsErrorRetryable               bool       `json:"isErrorRetryable"`
}

// AddPrintJobRequest describes a document to upload to the print queue. FileBytes
// and FileName are required; the remaining fields default to a single-copy,
// portrait, monochrome, one-sided job when left empty.
type AddPrintJobRequest struct {
	Copies               int
	ForceMediaSize       string
	OrientationRequested string
	PrintColorMode       string
	Sides                string
	JobName              string
	FileName             string
	FileContentType      string
	FileBytes            []byte
}

// doPrintHubRequest issues a print-hub request with the headers the web app sends.
// It reuses the authenticated base client (which carries the bearer token and
// handles response decompression) and returns an error for any non-2xx status.
func (w *WeWork) doPrintHubRequest(ctx context.Context, method, requestURL string, body io.Reader, contentType string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, requestURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("fe-pg", "/workplaceone/content2/print")
	req.Header.Set("Referer", "https://members.wework.com/workplaceone/content2/print")
	req.Header.Set("Request-Source", "MemberWeb/WorkplaceOne/Prod")
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	resp, err := w.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("request failed with status code: %d: %s", resp.StatusCode, clipBody(buf.Bytes()))
	}

	return resp, nil
}

// GetPrintQueue returns the member's current print queue. jobIDs is a
// comma-separated list of job ids to filter by; an empty value defaults to "0",
// which returns the full queue.
func (w *WeWork) GetPrintQueue(ctx context.Context, jobIDs string) (*PrintQueueResponse, error) {
	jobIDs = strings.TrimSpace(jobIDs)
	if jobIDs == "" {
		jobIDs = "0"
	}

	params := url.Values{}
	params.Add("jobIds", jobIDs)

	requestURL := "https://members.wework.com/workplaceone/api/wework-yardi/print-hub/get-print-queue?" + params.Encode()
	resp, err := w.doPrintHubRequest(ctx, http.MethodGet, requestURL, nil, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result PrintQueueResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode print queue response: %w", err)
	}

	return &result, nil
}

// applyDefaults fills the optional print-job fields with the values the web app
// uses when they are left empty.
func (input *AddPrintJobRequest) applyDefaults() {
	if input.Copies <= 0 {
		input.Copies = 1
	}
	if strings.TrimSpace(input.ForceMediaSize) == "" {
		input.ForceMediaSize = "null"
	}
	if strings.TrimSpace(input.OrientationRequested) == "" {
		input.OrientationRequested = "portrait"
	}
	if strings.TrimSpace(input.PrintColorMode) == "" {
		input.PrintColorMode = "monochrome"
	}
	if strings.TrimSpace(input.Sides) == "" {
		input.Sides = "one-sided"
	}
	if strings.TrimSpace(input.JobName) == "" {
		input.JobName = input.FileName
	}
	if strings.TrimSpace(input.FileContentType) == "" {
		input.FileContentType = "application/octet-stream"
	}
}

// buildPrintJobMultipart validates the request, applies defaults, and encodes the
// multipart body the print-hub add endpoint expects. It returns the encoded body
// together with the content type (including the generated boundary).
func buildPrintJobMultipart(input *AddPrintJobRequest) (*bytes.Buffer, string, error) {
	if len(input.FileBytes) == 0 {
		return nil, "", fmt.Errorf("file bytes are required")
	}
	if strings.TrimSpace(input.FileName) == "" {
		return nil, "", fmt.Errorf("file name is required")
	}
	input.applyDefaults()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	fields := map[string]string{
		"copies":               strconv.Itoa(input.Copies),
		"forceMediaSize":       input.ForceMediaSize,
		"orientationRequested": input.OrientationRequested,
		"printColorMode":       input.PrintColorMode,
		"sides":                input.Sides,
		"jobName":              input.JobName,
	}
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			return nil, "", fmt.Errorf("failed to write multipart field %s: %w", key, err)
		}
	}

	fileHeader := make(textproto.MIMEHeader)
	disposition := mime.FormatMediaType("form-data", map[string]string{
		"name":     "file",
		"filename": input.FileName,
	})
	fileHeader.Set("Content-Disposition", disposition)
	fileHeader.Set("Content-Type", input.FileContentType)
	fileWriter, err := writer.CreatePart(fileHeader)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create multipart file part: %w", err)
	}
	if _, err := fileWriter.Write(input.FileBytes); err != nil {
		return nil, "", fmt.Errorf("failed to write multipart file: %w", err)
	}
	if err := writer.Close(); err != nil {
		return nil, "", fmt.Errorf("failed to finalize multipart request: %w", err)
	}

	return &body, writer.FormDataContentType(), nil
}

// AddToPrintQueue uploads a document to the member's print queue as a multipart
// form, mirroring the web app's request.
func (w *WeWork) AddToPrintQueue(ctx context.Context, input AddPrintJobRequest) (*PrintJob, error) {
	body, contentType, err := buildPrintJobMultipart(&input)
	if err != nil {
		return nil, err
	}

	requestURL := "https://members.wework.com/workplaceone/api/wework-yardi/print-hub/add-to-print-queue"
	resp, err := w.doPrintHubRequest(ctx, http.MethodPost, requestURL, body, contentType)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result PrintJob
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode add print job response: %w", err)
	}

	return &result, nil
}
