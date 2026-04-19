package plugins

import (
	"fmt"
	"net/url"
	"strings"
)

// ValidateBlocks returns the first semantic error found in blocks, or nil if
// every block is well-formed. It is intentionally strict: malformed blocks are
// rejected at ingest time so the dispatcher never has to guess.
func ValidateBlocks(blocks []ContentBlock) error {
	if len(blocks) == 0 {
		return fmt.Errorf("item must contain at least one block")
	}

	for index, block := range blocks {
		if err := validateBlock(block); err != nil {
			return fmt.Errorf("block[%d]: %w", index, err)
		}
	}
	return nil
}

func validateBlock(block ContentBlock) error {
	switch block.Type {
	case BlockHeading:
		if block.Level < 1 || block.Level > 3 {
			return fmt.Errorf("heading.level must be 1..3, got %d", block.Level)
		}
		if strings.TrimSpace(block.Text) == "" {
			return fmt.Errorf("heading.text is required")
		}
		return nil
	case BlockParagraph:
		if strings.TrimSpace(block.Text) == "" {
			return fmt.Errorf("paragraph.text is required")
		}
		return nil
	case BlockImage:
		if err := validateHTTPURL(block.URL); err != nil {
			return fmt.Errorf("image.url: %w", err)
		}
		return nil
	case BlockLink:
		if err := validateHTTPURL(block.URL); err != nil {
			return fmt.Errorf("link.url: %w", err)
		}
		return nil
	case BlockDivider:
		return nil
	default:
		return fmt.Errorf("unsupported block type %q", block.Type)
	}
}

func validateHTTPURL(raw string) error {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return fmt.Errorf("url is required")
	}
	parsed, err := url.ParseRequestURI(trimmed)
	if err != nil {
		return fmt.Errorf("invalid url")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("url must use http or https scheme")
	}
	if parsed.Host == "" {
		return fmt.Errorf("url must include a host")
	}
	return nil
}
