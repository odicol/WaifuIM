package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"
	"waifuIM/internal/client"
	"waifuIM/internal/models"

	"github.com/spf13/cobra"
)

func getArtists(ctx context.Context, c client.WaifuIMClient, query url.Values) (string, error) {
	res, err := c.GET(ctx, "/artists", query)
	if err != nil {
		return "", err
	}

	var respModel models.ArtistsResponse
	if err := json.Unmarshal(res, &respModel); err != nil {
		return "", fmt.Errorf("failed to unmarshal albums response: %w", err)
	}

	return prettyPrint(respModel)
}

func buildArtistsParams(name, page, pageSize string) url.Values {
	query := url.Values{}
	if name != "" {
		query.Set("Name", name)
	}
	if page != "" {
		query.Set("Page", page)
	}
	if pageSize != "" {
		query.Set("PageSize", pageSize)
	}
	return query
}

func NewArtistsCMD() *cobra.Command {
	var name, page, pageSize string

	cmd := &cobra.Command{
		Use:   "artists",
		Short: "List artists",
		Long:  `List registered artists based on a specific search criteria`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c := client.New()

			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer stop()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()

			params := buildArtistsParams(name, page, pageSize)
			res, err := getArtists(ctx, c, params)
			if err != nil {
				return err
			}
			fmt.Println(res)
			return nil
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Tag name")
	cmd.Flags().StringVarP(&page, "page", "p", "", "Page number")
	cmd.Flags().StringVarP(&pageSize, "page_size", "s", "", "Page size")

	return cmd
}
