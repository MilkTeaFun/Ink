package printer

import (
	"context"
	_ "embed"
	"fmt"
	"strings"

	memobirdapi "github.com/ruhuang2001/memobird-go/memobird"
	memobirdtextrender "github.com/ruhuang2001/memobird-go/textrender"
)

type imagePrintPipeline interface {
	PrintJob(ctx context.Context, client *memobirdapi.Client, job Job) (*memobirdapi.PrintResponse, error)
}

//go:embed assets/NotoSansSC-Regular.otf
var printerFontData []byte

type textImagePrintPipeline struct{}

func (p textImagePrintPipeline) PrintJob(ctx context.Context, client *memobirdapi.Client, job Job) (*memobirdapi.PrintResponse, error) {
	imageBase64, err := memobirdtextrender.RenderBase64PNG(renderPrintableText(job.Title, job.Content), memobirdtextrender.Options{
		Width:      384,
		Padding:    18,
		FontSize:   22,
		LineHeight: 1.55,
		FontData:   printerFontData,
	})
	if err != nil {
		return nil, err
	}

	return client.PrintImage(ctx, imageBase64)
}

func renderPrintableText(title string, content string) string {
	parts := []string{strings.TrimSpace(title), strings.TrimSpace(content)}
	return strings.Join(parts, "\n\n")
}

func primaryPrintResponse(responses []*memobirdapi.PrintResponse) (*memobirdapi.PrintResponse, error) {
	for _, response := range responses {
		if response != nil {
			return response, nil
		}
	}

	return nil, fmt.Errorf("print returned no responses")
}
