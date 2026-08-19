package commands

import (
	"encoding/json"
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dvcrn/wework-cli/pkg/spinner"
	"github.com/dvcrn/wework-cli/pkg/wework"
	"github.com/spf13/cobra"
)

func NewPrintCommand(authenticate func() (*wework.WeWork, error)) *cobra.Command {
	var jobIDs string

	cmd := &cobra.Command{
		Use:     "print",
		Aliases: []string{"printhub", "print-hub", "printqueue", "print-queue"},
		Short:   "Manage WeWork print queue",
		Long:    `View your WeWork print queue or submit new print jobs.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPrintQueue(cmd, authenticate, jobIDs)
		},
	}

	cmd.PersistentFlags().StringVar(&jobIDs, "job-ids", "", "Filter by comma-separated job IDs (default: all jobs)")

	cmd.AddCommand(
		newPrintQueueCommand(authenticate),
		newPrintAddCommand(authenticate),
	)

	return cmd
}

func newPrintQueueCommand(authenticate func() (*wework.WeWork, error)) *cobra.Command {
	var jobIDs string

	cmd := &cobra.Command{
		Use:     "queue",
		Aliases: []string{"list", "ls", "status"},
		Short:   "List jobs in your print queue",
		Long:    `List all pending or recent jobs in your WeWork print queue.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPrintQueue(cmd, authenticate, jobIDs)
		},
	}

	cmd.Flags().StringVar(&jobIDs, "job-ids", "", "Filter by comma-separated job IDs (default: all jobs)")

	return cmd
}

func runPrintQueue(cmd *cobra.Command, authenticate func() (*wework.WeWork, error), jobIDs string) error {
	ww, err := authenticate()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	jsonOut, _ := cmd.Flags().GetBool("json")
	var res *wework.PrintQueueResponse

	if jsonOut {
		r, err := ww.GetPrintQueue(ctx, jobIDs)
		if err != nil {
			return fmt.Errorf("failed to get print queue: %w", err)
		}
		res = r

		b, err := json.MarshalIndent(res, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(b))
		return nil
	}

	if err := spinner.WithContinuousSpinner(func(cs *spinner.ContinuousSpinner) error {
		cs.Update("Fetching print queue…")
		r, err := ww.GetPrintQueue(ctx, jobIDs)
		if err != nil {
			return fmt.Errorf("failed to get print queue: %w", err)
		}
		res = r
		cs.Success("Print queue retrieved")
		return nil
	}); err != nil {
		return err
	}

	if len(res.Content) == 0 {
		fmt.Println("No print jobs in queue.")
		return nil
	}

	fmt.Printf("%-38s%-25s%-12s%-7s%-8s%-12s%-14s%s\n",
		"Job ID", "Name", "Status", "Pages", "Copies", "Color", "Sides", "Created")
	fmt.Println(strings.Repeat("-", 125))

	for _, job := range res.Content {
		name := job.JobName
		if len(name) > 23 {
			name = name[:23]
		}
		createdStr := ""
		if !job.CreatedTime.Time.IsZero() {
			createdStr = job.CreatedTime.Time.Format("2006-01-02 15:04")
		}

		fmt.Printf("%-38s%-25s%-12s%-7d%-8d%-12s%-14s%s\n",
			job.ID,
			name,
			job.Status,
			job.PagesInJob,
			job.Copies,
			job.PrintColorMode,
			job.Sides,
			createdStr)
	}

	return nil
}

func newPrintAddCommand(authenticate func() (*wework.WeWork, error)) *cobra.Command {
	var filePath string
	var jobName string
	var copies int
	var colorMode string
	var orientation string
	var sides string
	var mediaSize string
	var contentType string

	cmd := &cobra.Command{
		Use:     "add [FILE]",
		Aliases: []string{"upload", "submit"},
		Short:   "Upload a document to the print queue",
		Long:    `Upload a PDF or document file to your WeWork print queue.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 && filePath == "" {
				filePath = args[0]
			}

			if filePath == "" {
				return fmt.Errorf("file path is required: provide as an argument or via --file flag")
			}
			filePath = filepath.Clean(filePath)

			// Validate file exists and read content
			fileInfo, err := os.Stat(filePath)
			if err != nil {
				return fmt.Errorf("failed to access file %s: %w", filePath, err)
			}
			if fileInfo.IsDir() {
				return fmt.Errorf("%s is a directory, not a file", filePath)
			}
			if fileInfo.Size() == 0 {
				return fmt.Errorf("file %s is empty", filePath)
			}

			fileBytes, err := os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("failed to read file %s: %w", filePath, err)
			}

			fileName := filepath.Base(filePath)
			if jobName == "" {
				jobName = fileName
			}

			// Detect content type if not explicitly provided
			if contentType == "" {
				ext := strings.ToLower(filepath.Ext(filePath))
				contentType = mime.TypeByExtension(ext)
				if contentType == "" {
					if ext == ".pdf" {
						contentType = "application/pdf"
					} else {
						contentType = "application/octet-stream"
					}
				}
			}

			// Validate flags
			if copies <= 0 {
				copies = 1
			}
			colorMode = strings.ToLower(colorMode)
			if colorMode != "monochrome" && colorMode != "color" && colorMode != "colour" {
				return fmt.Errorf("invalid color mode %q: must be 'monochrome' or 'color'", colorMode)
			}
			if colorMode == "colour" {
				colorMode = "color"
			}

			orientation = strings.ToLower(orientation)
			if orientation != "portrait" && orientation != "landscape" {
				return fmt.Errorf("invalid orientation %q: must be 'portrait' or 'landscape'", orientation)
			}

			sides = strings.ToLower(sides)
			switch sides {
			case "one-sided", "1-sided", "single":
				sides = "one-sided"
			case "two-sided-long-edge", "2-sided-long", "double-long", "duplex-long":
				sides = "two-sided-long-edge"
			case "two-sided-short-edge", "2-sided-short", "double-short", "duplex-short":
				sides = "two-sided-short-edge"
			default:
				return fmt.Errorf("invalid sides %q: must be 'one-sided', 'two-sided-long-edge', or 'two-sided-short-edge'", sides)
			}

			ww, err := authenticate()
			if err != nil {
				return err
			}

			ctx := cmd.Context()
			req := wework.AddPrintJobRequest{
				Copies:               copies,
				ForceMediaSize:       mediaSize,
				OrientationRequested: orientation,
				PrintColorMode:       colorMode,
				Sides:                sides,
				JobName:              jobName,
				FileName:             fileName,
				FileContentType:      contentType,
				FileBytes:            fileBytes,
			}

			jsonOut, _ := cmd.Flags().GetBool("json")
			var job *wework.PrintJob

			if jsonOut {
				j, err := ww.AddToPrintQueue(ctx, req)
				if err != nil {
					return fmt.Errorf("failed to upload print job: %w", err)
				}
				job = j

				b, err := json.MarshalIndent(job, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal JSON: %w", err)
				}
				fmt.Println(string(b))
				return nil
			}

			if err := spinner.WithContinuousSpinner(func(cs *spinner.ContinuousSpinner) error {
				cs.Update(fmt.Sprintf("Uploading %s (%d bytes) to print queue…", fileName, len(fileBytes)))
				j, err := ww.AddToPrintQueue(ctx, req)
				if err != nil {
					return fmt.Errorf("failed to upload print job: %w", err)
				}
				job = j
				cs.Success("Print job uploaded successfully!")
				return nil
			}); err != nil {
				return err
			}

			fmt.Println()
			fmt.Printf("Job Details:\n")
			fmt.Printf("  Job ID:      %s\n", job.ID)
			fmt.Printf("  Job Name:    %s\n", job.JobName)
			fmt.Printf("  Status:      %s\n", job.Status)
			fmt.Printf("  Pages:       %d\n", job.PagesInJob)
			fmt.Printf("  Copies:      %d\n", job.Copies)
			fmt.Printf("  Color Mode:  %s\n", job.PrintColorMode)
			fmt.Printf("  Sides:       %s\n", job.Sides)
			fmt.Printf("  Orientation: %s\n", job.OrientationRequested)
			if !job.CreatedTime.Time.IsZero() {
				fmt.Printf("  Created:     %s\n", job.CreatedTime.Time.Format(time.RFC1123))
			}
			if !job.ExpiryTime.Time.IsZero() {
				fmt.Printf("  Expires:     %s\n", job.ExpiryTime.Time.Format(time.RFC1123))
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to file to print")
	cmd.Flags().StringVarP(&jobName, "name", "n", "", "Job name (defaults to file name)")
	cmd.Flags().IntVarP(&copies, "copies", "c", 1, "Number of copies")
	cmd.Flags().StringVar(&colorMode, "color", "monochrome", "Print color mode ('monochrome' or 'color')")
	cmd.Flags().StringVar(&orientation, "orientation", "portrait", "Orientation ('portrait' or 'landscape')")
	cmd.Flags().StringVar(&sides, "sides", "one-sided", "Sides ('one-sided', 'two-sided-long-edge', 'two-sided-short-edge')")
	cmd.Flags().StringVar(&mediaSize, "media-size", "null", "Media size (e.g. 'null', 'A4', 'Letter')")
	cmd.Flags().StringVar(&contentType, "content-type", "", "Explicit MIME content type")

	return cmd
}
