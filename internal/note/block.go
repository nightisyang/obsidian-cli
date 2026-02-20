package note

import (
	"strings"
	"time"

	"github.com/nightisyang/obsidian-cli/internal/errs"
)

func GetBlock(vaultRoot, path, blockID string) (Block, error) {
	n, err := Read(vaultRoot, path)
	if err != nil {
		return Block{}, err
	}
	block, err := blockFromBody(n.Body, n.Path, blockID)
	if err != nil {
		return Block{}, err
	}
	return block, nil
}

func SetBlock(vaultRoot, path, blockID, content string) (Block, error) {
	if strings.Contains(content, "\n") {
		return Block{}, errs.New(errs.ExitValidation, "content must be a single line")
	}

	n, err := Read(vaultRoot, path)
	if err != nil {
		return Block{}, err
	}

	lines := strings.Split(n.Body, "\n")
	index, style, err := findBlockLine(lines, blockID)
	if err != nil {
		return Block{}, err
	}

	switch style {
	case blockStyleAnchorOnly:
		if index == 0 {
			return Block{}, errs.New(errs.ExitValidation, "anchor-only block has no content line")
		}
		lines[index-1] = content
	case blockStyleInline:
		lines[index] = content + " ^" + blockID
	default:
		return Block{}, errs.New(errs.ExitGeneric, "unknown block style")
	}

	n.Body = strings.Join(lines, "\n")
	updated, err := Write(vaultRoot, n.Path, n, false, time.Now())
	if err != nil {
		return Block{}, err
	}
	return blockFromBody(updated.Body, updated.Path, blockID)
}

type blockStyle int

const (
	blockStyleInline blockStyle = iota
	blockStyleAnchorOnly
)

func blockFromBody(body, path, blockID string) (Block, error) {
	lines := strings.Split(body, "\n")
	index, style, err := findBlockLine(lines, blockID)
	if err != nil {
		return Block{}, err
	}

	switch style {
	case blockStyleAnchorOnly:
		if index == 0 {
			return Block{}, errs.New(errs.ExitValidation, "anchor-only block has no content line")
		}
		return Block{
			Path:       path,
			BlockID:    blockID,
			Content:    strings.TrimSpace(lines[index-1]),
			Line:       index,
			AnchorLine: index + 1,
		}, nil
	default:
		line := strings.TrimSpace(lines[index])
		anchor := " ^" + blockID
		content := strings.TrimSpace(strings.TrimSuffix(line, anchor))
		return Block{
			Path:       path,
			BlockID:    blockID,
			Content:    content,
			Line:       index + 1,
			AnchorLine: index + 1,
		}, nil
	}
}

func findBlockLine(lines []string, blockID string) (int, blockStyle, error) {
	needle := "^" + strings.TrimSpace(blockID)
	if needle == "^" {
		return -1, blockStyleInline, errs.New(errs.ExitValidation, "block id is required")
	}
	for i, raw := range lines {
		line := strings.TrimSpace(strings.TrimRight(raw, "\r"))
		if line == needle {
			return i, blockStyleAnchorOnly, nil
		}
		if strings.HasSuffix(line, " "+needle) {
			return i, blockStyleInline, nil
		}
	}
	return -1, blockStyleInline, errs.New(errs.ExitNotFound, "block not found")
}
