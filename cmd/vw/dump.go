package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/AlexGustafsson/vw-manual-exporter/internal/vwapps"
	"github.com/urfave/cli/v3"
)

func DumpAction(ctx context.Context, cmd *cli.Command) error {
	vin := cmd.String("vin")
	language := cmd.String("language")
	outputDirectory := cmd.String("output-dir")

	err := os.MkdirAll(outputDirectory, 0777)
	if err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	client := &vwapps.DigitalManualClient{
		Client:   http.DefaultClient,
		Language: language,
	}

	slog.Debug("Getting topic tree")
	topicTree, err := client.GetTopicTree(ctx, vin)
	if err != nil {
		return fmt.Errorf("failed to fetch topic tree: %w", err)
	}

	// Write topic tree to disk
	{
		content, err := json.Marshal(topicTree)
		if err != nil {
			panic(err)
		}

		err = os.WriteFile(filepath.Join(outputDirectory, "topic-tree.json"), content, 0644)
		if err != nil {
			return fmt.Errorf("failed to write topic tree: %w", err)
		}
	}

	// Try to fetch and write car photos to disk
	{
		views := []vwapps.VehicleView{
			vwapps.VehicleViewSideLeft,
			vwapps.VehicleViewSideRight,
		}

		for _, view := range views {
			slog.Debug("Getting vehicle picture URL", slog.String("view", string(view)))
			url, err := client.GetVehiclePictureURL(ctx, vin, view)
			if err == nil {
				slog.Debug("Downloading vehicle picture", slog.String("view", string(view)))
				err := downloadFile(ctx, url, filepath.Join(outputDirectory, fmt.Sprintf("vehicle-%s.png", view)))
				if err != nil {
					slog.Warn("Failed to download vehicle picture", slog.String("view", string(view)), slog.Any("error", err))
					continue
				}
			} else {
				slog.Warn("Failed to get vehicle picture URL", slog.Any("error", err))
				continue
			}
		}
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

		log := slog.With(slog.String("topicId", topicID))

		outputFile := filepath.Join(outputDirectory, topicID+".html")

		_, err := os.Stat(outputFile)
		if err == nil {
			log.Debug("Skipping already fetched topic")
			continue
		}

		log.Debug("Fetching topic")
		html, err := client.GetTopicHTML(ctx, vin, topicID)
		if err != nil && err != vwapps.ErrTopicNotFound {
			return fmt.Errorf("failed to fetch topic: %w", err)
		} else if err == nil {
			log.Debug("Writing output")
			err = os.WriteFile(outputFile, []byte(html), 0644)
			if err != nil {
				return fmt.Errorf("failed to write topic output: %w", err)
			}
		}
	}

	return nil
}

func downloadFile(ctx context.Context, url string, path string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return fmt.Errorf("unexpected status code: %d", err)
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, res.Body)
	return err
}
