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

func getRandomImage(ctx context.Context, c client.WaifuIMClient, query url.Values) (string, error) {
	res, err := c.GET(ctx, "/images", query)
	if err != nil {
		return "", err
	}

	var respModel models.RandomImageResponse
	if err := json.Unmarshal(res, &respModel); err != nil {
		return "", fmt.Errorf("failed to unmarshal albums response: %w", err)
	}

	return prettyPrint(respModel)
}

func buildRandomParams(page, pageSize, includeTag, excludeTag, isNSFW, includeArtists string) url.Values {
	queryParams := url.Values{}
	if page != "" {
		queryParams.Set("Page", page)
	}
	if pageSize != "" {
		queryParams.Set("PageSize", pageSize)
	}
	if includeTag != "" {
		queryParams.Set("IncludedTags", includeTag)
	}
	if excludeTag != "" {
		queryParams.Set("ExcludedTags", excludeTag)
	}
	if isNSFW != "" {
		queryParams.Set("IsNsfw", isNSFW)
	}
	if includeArtists != "" {
		queryParams.Set("IncludedArtists", includeArtists)
	}
	return queryParams
}

func NewRandomCMD() *cobra.Command {
	var page, pageSize, includeTag, excludeTag, includeArtists, isNSFW string

	cmd := &cobra.Command{
		Use:   "random",
		Short: "Generate random image",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := client.New()

			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer stop()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()

			params := buildRandomParams(page, pageSize, includeTag, excludeTag, isNSFW, includeArtists)
			res, err := getRandomImage(ctx, c, params)
			if err != nil {
				return err
			}
			fmt.Println(res)
			return nil
		},
	}

	cmd.Flags().StringVar(&page, "page", "", "page number")
	cmd.Flags().StringVar(&pageSize, "pageSize", "", "number of images per page")
	cmd.Flags().StringVar(&includeTag, "include", "", "include images matching tags (AND logic)")
	cmd.Flags().StringVar(&excludeTag, "exclude", "", "exclude images matching tags (OR logic)")
	cmd.Flags().StringVar(&isNSFW, "isNSFW", "False", "False, True, or All")
	cmd.Flags().StringVar(&includeArtists, "artist", "", "filter by artist")

	return cmd
}
