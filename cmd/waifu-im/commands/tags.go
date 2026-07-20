package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"waifuIM/internal/client"
	"waifuIM/internal/models"

	"github.com/spf13/cobra"
)

func getTags(ctx context.Context, c client.WaifuIMClient, query url.Values) (string, error) {
	res, err := c.GET(ctx, "/tags", query)
	if err != nil {
		return "", err
	}

	var respModel models.TagsResponse
	if err := json.Unmarshal(res, &respModel); err != nil {
		return "", fmt.Errorf("failed to unmarshal albums response: %w", err)
	}

	return prettyPrint(respModel)
}

func buildTagsParams(name, page, pageSize string) url.Values {
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

func NewTagsCMD() *cobra.Command {
	var name, page, pageSize string

	cmd := &cobra.Command{
		Use:   "tags",
		Short: "List tags",
		Long:  `List available tags based on a specific search criteria`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c := client.New()
			ctx := getContext()
			params := buildTagsParams(name, page, pageSize)
			res, err := getTags(ctx, c, params)
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
