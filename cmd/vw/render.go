package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/AlexGustafsson/vw-manual-exporter/internal/vwapps"
	"github.com/urfave/cli/v3"
)

func RenderAction(ctx context.Context, cmd *cli.Command) error {
	inputDirectory := cmd.String("input-dir")
	outputDirectory := cmd.String("output-dir")
	format := cmd.String("format")

	if format != "html" {
		return fmt.Errorf("unsupported format: %s", format)
	}

	err := os.MkdirAll(outputDirectory, 0777)
	if err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	var topicTree *vwapps.TopicTree
	{
		contents, err := os.ReadFile(filepath.Join(inputDirectory, "topic-tree.json"))
		if err != nil {
			return fmt.Errorf("failed to read topic tree: %w", err)
		}

		err = json.Unmarshal(contents, &topicTree)
		if err != nil {
			return fmt.Errorf("failed to parse topic tree: %w", err)
		}
	}

	stylesheet, err := os.ReadFile(filepath.Join(inputDirectory, "stylesheet.css"))
	if err != nil {
		return fmt.Errorf("failed to read stylesheet: %w", err)
	}

	// Write topics to disk
	nodes := []vwapps.TopicTreeNode{topicTree.RootNode}
	for i := 0; i < len(nodes); i++ {
		node := nodes[i]
		nodes = append(nodes, node.Children...)

		topicID := node.TargetTopicID
		if topicID == "" {
			continue
		}

		content, err := os.ReadFile(filepath.Join(inputDirectory, topicID+".html"))
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("failed to read topic: %w", err)
		}

		html, err := vwapps.RenderHTML(bytes.NewReader(content), string(stylesheet))
		if err != nil {
			return fmt.Errorf("failed to render topic: %w", err)
		}

		err = os.WriteFile(filepath.Join(outputDirectory, topicID+".html"), []byte(html), 0644)
		if err != nil {
			return fmt.Errorf("failed to write topic output: %w", err)
		}
	}

	return nil
}
