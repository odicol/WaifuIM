package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"
	"waifuIM/internal/client"
	"waifuIM/internal/models"

	"github.com/spf13/cobra"
)

func getAlbums(ctx context.Context, c client.WaifuIMClient, userID string, query url.Values) (string, error) {
	res, err := c.GET(ctx, "/users/"+userID+"/albums", query)
	if err != nil {
		return "", fmt.Errorf("get albums for user \"%s\": %w", userID, err)
	}

	var respModel models.AlbumsResponse
	if err := json.Unmarshal(res, &respModel); err != nil {
		return "", fmt.Errorf("failed to unmarshal albums response: %w", err)
	}
	return prettyPrint(respModel)
}

func getAlbumDetails(ctx context.Context, c client.WaifuIMClient, userID string, albumID string) (string, error) {
	res, err := c.GET(ctx, "/users/"+userID+"/albums/"+albumID, nil)
	if err != nil {
		return "", fmt.Errorf("get album id \"%s\" details for user \"%s\": %w", albumID, userID, err)
	}

	var respModel models.AlbumItem
	if err := json.Unmarshal(res, &respModel); err != nil {
		return "", fmt.Errorf("failed to unmarshal albums response: %w", err)
	}
	return prettyPrint(respModel)
}

func postAlbum(ctx context.Context, c client.WaifuIMClient, userID string, body io.Reader) (string, error) {
	res, err := c.POST(ctx, "/users/"+userID+"/albums", body)
	if err != nil {
		return "", fmt.Errorf("post album for user \"%s\": %w", userID, err)
	}

	var respModel models.AlbumItem
	if err := json.Unmarshal(res, &respModel); err != nil {
		return "", fmt.Errorf("failed to unmarshal albums response: %w", err)
	}
	return prettyPrint(respModel)
}

func patchAlbum(ctx context.Context, c client.WaifuIMClient, userID, albumID string, body io.Reader) (string, error) {
	res, err := c.PATCH(ctx, "/users/"+userID+"/albums/"+albumID, body)
	if err != nil {
		return "", fmt.Errorf("patch album id \"%s\" for user \"%s\": %w", albumID, userID, err)
	}

	var respModel models.AlbumItem
	if err := json.Unmarshal(res, &respModel); err != nil {
		return "", fmt.Errorf("failed to unmarshal albums response: %w", err)
	}
	return prettyPrint(respModel)
}

func deleteAlbum(ctx context.Context, c client.WaifuIMClient, userID, albumID string) (bool, error) {
	success, err := c.DELETE(ctx, "/users/"+userID+"/albums/"+albumID)
	if err != nil {
		return false, fmt.Errorf("delete album id \"%s\" for user \"%s\": %w", albumID, userID, err)
	}
	return success, nil
}

func NewAlbumsCMD() *cobra.Command {
	var page, pageSize, userID string

	cmd := &cobra.Command{
		Use:   "albums",
		Short: "List albums",
		Long:  `List the albums of a specific user (or for the current user if not specified)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c := client.New(client.WithAPIKey(os.Getenv("API_KEY")))

			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer stop()
			ctx, cancel := context.WithTimeout(cmd.Context(), 10*time.Second)
			defer cancel()

			query := url.Values{}
			if page != "" {
				query.Set("Page", page)
			}
			if pageSize != "" {
				query.Set("PageSize", pageSize)
			}

			res, err := getAlbums(ctx, c, userID, query)
			if err != nil {
				return err
			}
			fmt.Println(res)
			return nil
		},
	}

	cmd.Flags().StringVarP(&page, "page", "p", "", "Page number")
	cmd.Flags().StringVarP(&pageSize, "page_size", "s", "", "Page size")
	cmd.Flags().StringVarP(&userID, "user", "u", "me", "The user ID or \"me\" for the current user")

	return cmd
}

func NewAlbumDetailsCMD() *cobra.Command {
	var userID, albumID string

	cmd := &cobra.Command{
		Use:   "album",
		Short: "Get album details",
		Long:  `Get album details by ID`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c := client.New(client.WithAPIKey(os.Getenv("API_KEY")))

			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer stop()
			ctx, cancel := context.WithTimeout(cmd.Context(), 10*time.Second)
			defer cancel()

			res, err := getAlbumDetails(ctx, c, userID, albumID)
			if err != nil {
				return err
			}
			fmt.Println(res)
			return nil
		},
	}

	cmd.Flags().StringVarP(&userID, "usedID", "u", "me", "The user ID or \"me\" for the current user")
	cmd.Flags().StringVarP(&albumID, "albumID", "a", "favorites", "The album ID or \"favorites\" if not specified")

	return cmd
}

func NewCreateAlbumCMD() *cobra.Command {
	var name, description, userID string

	cmd := &cobra.Command{
		Use:   "create-album",
		Short: "Create an album",
		Long:  `Create an album with the name and description for specific user (or for the current user if not specified)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// only 1 retry to avoid creating duplicate objects
			c := client.New(client.WithAPIKey(os.Getenv("API_KEY")), client.WithMaxRetries(1))

			payload := models.AlbumMetadata{
				Name:        name,
				Description: description,
			}

			jsonBytes, err := json.Marshal(payload)
			if err != nil {
				return fmt.Errorf("failed to encode request body: %w", err)
			}

			bodyReader := bytes.NewBuffer(jsonBytes)

			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer stop()
			ctx, cancel := context.WithTimeout(cmd.Context(), 10*time.Second)
			defer cancel()

			res, err := postAlbum(ctx, c, userID, bodyReader)
			if err != nil {
				return err
			}
			fmt.Println(res)
			return nil
		},
	}
	cmd.Flags().StringVarP(&name, "name", "n", "", "Name of the album")
	cmd.Flags().StringVarP(&description, "description", "d", "", "Description of the album")
	cmd.Flags().StringVarP(&userID, "usedID", "u", "me", "The user ID or \"me\" for the current user")

	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("description")

	return cmd
}

func NewUpdateAlbumCMD() *cobra.Command {
	var userID, albumID, name, description string

	cmd := &cobra.Command{
		Use:   "update-album",
		Short: "Update an album",
		Long:  `Update an album with the name and description`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// only 1 retry to avoid multiple updates
			c := client.New(client.WithAPIKey(os.Getenv("API_KEY")), client.WithMaxRetries(1))

			payload := models.AlbumMetadata{
				Name:        name,
				Description: description,
			}

			jsonBytes, err := json.Marshal(payload)
			if err != nil {
				return fmt.Errorf("failed to encode request body: %w", err)
			}

			bodyReader := bytes.NewBuffer(jsonBytes)

			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer stop()
			ctx, cancel := context.WithTimeout(cmd.Context(), 10*time.Second)
			defer cancel()

			res, err := patchAlbum(ctx, c, userID, albumID, bodyReader)
			if err != nil {
				return err
			}
			fmt.Println(res)
			return nil
		},
	}
	cmd.Flags().StringVarP(&name, "name", "n", "", "New name of the album")
	cmd.Flags().StringVarP(&description, "description", "d", "", "New description of the album")
	cmd.Flags().StringVarP(&userID, "usedID", "u", "me", "The user ID or \"me\" for the current user")
	cmd.Flags().StringVarP(&albumID, "albumID", "a", "", "The album IDto be updated")

	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("description")
	_ = cmd.MarkFlagRequired("albumID")

	return cmd
}

func NewDeleteAlbumCMD() *cobra.Command {
	var userID, albumID string

	cmd := &cobra.Command{
		Use:   "delete-album",
		Short: "Delete an album",
		Long:  `Delete an existing album with`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// only 1 retry to avoid multiple deletes
			c := client.New(client.WithAPIKey(os.Getenv("API_KEY")), client.WithMaxRetries(1))

			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer stop()
			ctx, cancel := context.WithTimeout(cmd.Context(), 10*time.Second)
			defer cancel()

			res, err := deleteAlbum(ctx, c, userID, albumID)
			if err != nil {
				return err
			}
			fmt.Println(res)
			return nil
		},
	}
	cmd.Flags().StringVarP(&userID, "usedID", "u", "me", "The user ID or \"me\" for the current user")
	cmd.Flags().StringVarP(&albumID, "albumID", "a", "", "The album IDto be updated")

	_ = cmd.MarkFlagRequired("albumID")

	return cmd
}
