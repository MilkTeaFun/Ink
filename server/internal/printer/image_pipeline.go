package printer

import (
	"context"
	"fmt"
	"html"
	"strings"
	"time"

	memobirdapi "github.com/ruhuang2001/memobird-go/memobird"
	memobirdrenderer "github.com/ruhuang2001/memobird-go/renderer"
)

type imagePrintPipeline interface {
	PrintJob(ctx context.Context, client *memobirdapi.Client, job Job) (*memobirdapi.PrintResponse, error)
}

// htmlImagePrintPipeline is a temporary adapter.
// Replace this with the future memobird-go native text/image pipeline once it exists.
type htmlImagePrintPipeline struct {
	timeout time.Duration
}

func (p htmlImagePrintPipeline) PrintJob(ctx context.Context, client *memobirdapi.Client, job Job) (*memobirdapi.PrintResponse, error) {
	renderer := memobirdrenderer.New(p.timeout)
	defer renderer.Close()

	responses, err := client.PrintHTMLAsImages(ctx, renderer, renderPrintHTML(job.Title, job.Content))
	if err != nil {
		return nil, err
	}

	return primaryPrintResponse(responses)
}

func renderPrintHTML(title string, content string) string {
	escapedTitle := html.EscapeString(strings.TrimSpace(title))
	escapedContent := strings.ReplaceAll(html.EscapeString(strings.TrimSpace(content)), "\n", "<br>")

	return fmt.Sprintf(
		`<article style="width: 100%%; box-sizing: border-box; padding: 10px 8px 14px; color: #111827; font-family: 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', 'Noto Sans CJK SC', 'WenQuanYi Zen Hei', sans-serif;"><h1 style="font-size: 24px; font-weight: 700; line-height: 1.35; margin: 0 0 14px 0;">%s</h1><div style="font-size: 17px; line-height: 1.75; white-space: normal; word-break: break-word;">%s</div></article>`,
		escapedTitle,
		escapedContent,
	)
}

func primaryPrintResponse(responses []*memobirdapi.PrintResponse) (*memobirdapi.PrintResponse, error) {
	for _, response := range responses {
		if response != nil {
			return response, nil
		}
	}

	return nil, fmt.Errorf("print returned no responses")
}
