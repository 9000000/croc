package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/rivo/uniseg"
	"github.com/schollz/croc/v11/src/storeclient"
	"github.com/schollz/croc/v11/src/termui"
)

func TestStoredCallbacksClearPreviousLine(t *testing.T) {
	var output bytes.Buffer
	callbacks := newStoredCallbacks(&output, false, "Uploading")
	progress := storeclient.Progress{
		FileName:   "LICENSE",
		FileCount:  1,
		TotalBytes: 100,
		TotalSize:  100,
	}
	callbacks.Progress(progress)
	callbacks.Status("Encrypted upload complete")

	got := output.String()
	if !strings.Contains(got, "Uploading LICENSE") || !strings.Contains(got, "100% |") {
		t.Fatalf("stored progress output has no completed bar: %q", got)
	}
	if !strings.Contains(got, "\n\rEncrypted upload complete") {
		t.Fatalf("stored completion status did not follow the bar: %q", got)
	}
}

func TestStoredCallbacksClearUnicodeLineByDisplayWidth(t *testing.T) {
	var output bytes.Buffer
	callbacks := newStoredCallbacks(&output, false, "Uploading")
	callbacks.Status("🎉")
	callbacks.Status("Done")

	want := "\r🎉\r  \rDone"
	if got := output.String(); got != want {
		t.Fatalf("stored Unicode status output = %q; want %q", got, want)
	}
}

func TestStoredCallbacksQuiet(t *testing.T) {
	var output bytes.Buffer
	callbacks := newStoredCallbacks(&output, true, "Uploading")
	callbacks.Status("Preparing")
	callbacks.Progress(storeclient.Progress{
		FileName:   "LICENSE",
		TotalBytes: 50,
		TotalSize:  100,
	})
	if output.Len() != 0 {
		t.Fatalf("quiet stored progress output = %q; want no output", output.String())
	}
}

func TestStoredCallbacksUseRegularCrocPalette(t *testing.T) {
	var output bytes.Buffer
	callbacks := newStyledStoredCallbacks(&output, false, true, "Uploading")
	callbacks.Progress(storeclient.Progress{
		FileName:   "LICENSE",
		FileCount:  1,
		TotalBytes: 50,
		TotalSize:  100,
	})
	got := output.String()
	if !strings.Contains(got, termui.Bold+"LICENSE"+termui.Reset) {
		t.Fatalf("stored progress filename is not bold: %q", got)
	}
	callbacks.Progress(storeclient.Progress{FileName: "LICENSE", FileCount: 1, TotalBytes: 100, TotalSize: 100})
	callbacks.Status("Encrypted upload complete")
	got = output.String()
	if !strings.Contains(got, termui.Green) || !strings.Contains(got, "100% |") {
		t.Fatalf("stored completed bar is not green: %q", got)
	}
	if !strings.Contains(got, termui.Green+"Encrypted upload complete"+termui.Reset) {
		t.Fatalf("stored completion is not green: %q", got)
	}
}

func TestStoredCallbacksMeasureStyledUnicodeByDisplayWidth(t *testing.T) {
	var output bytes.Buffer
	callbacks := newStyledStoredCallbacks(&output, false, true, "Uploading")
	callbacks.Status("Uploading 🎉")
	callbacks.Status("Done")

	wantClear := "\r" + strings.Repeat(" ", uniseg.StringWidth("Uploading 🎉")) + "\r"
	if got := output.String(); !strings.Contains(got, wantClear) {
		t.Fatalf("styled Unicode status did not clear by display width: %q", got)
	}
}

func TestStoredCallbacksUseAggregateMultiFileProgress(t *testing.T) {
	var output bytes.Buffer
	callbacks := newStoredCallbacks(&output, false, "Downloading")
	callbacks.Status("Downloading first.txt")
	callbacks.Progress(storeclient.Progress{
		FileName: "first.txt", FileCount: 2, TotalBytes: 25, TotalSize: 100,
	})
	callbacks.Status("Downloading second.txt")
	callbacks.Progress(storeclient.Progress{
		FileName: "second.txt", FileCount: 2, TotalBytes: 100, TotalSize: 100,
	})

	got := output.String()
	if !strings.Contains(got, "Downloading 2 files") || !strings.Contains(got, "100% |") {
		t.Fatalf("stored multi-file progress is not aggregate: %q", got)
	}
	if strings.Contains(got, "second.txt") {
		t.Fatalf("stored aggregate progress switched to a concurrent filename: %q", got)
	}
}

func TestFormatStoredSendInstructionsUsesRegularCrocPalette(t *testing.T) {
	plain := formatStoredSendInstructions(
		"tomorrow", "https://example.com/#secret", "croc-store-v1.token", "transfer-id", "one verified download", false,
	)
	colored := formatStoredSendInstructions(
		"tomorrow", "https://example.com/#secret", "croc-store-v1.token", "transfer-id", "one verified download", true,
	)
	if termui.Plain(colored) != plain {
		t.Fatalf("colored stored instructions changed text:\n%s", colored)
	}
	for _, secret := range []string{"https://example.com/#secret", "croc-store-v1.token", "transfer-id"} {
		if !strings.Contains(colored, termui.Yellow+secret+termui.Reset) {
			t.Fatalf("stored instructions do not highlight %q: %q", secret, colored)
		}
	}
	if !strings.Contains(colored, termui.Green+"Stored transfer is encrypted") {
		t.Fatalf("stored ready message is not green: %q", colored)
	}
}
