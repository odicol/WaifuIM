# waifu-im CLI

A command-line client for the [waifu.im](https://waifu.im) API.

## Requirements

- Go 1.21+
- A waifu.im API key (required for album operations, optional for public endpoints)

## Setup

Copy `.env.example` to `.env` and fill in your API key:

```
API_KEY=your_api_key_here
```

## Commands

### `random` — fetch random images

```
waifu-im-client random [flags]

Flags:
  --page        page number
  --pageSize    number of images per page
  --include     include images matching tags (AND logic)
  --exclude     exclude images matching tags (OR logic)
  --isNSFW      False, True, or All (default: False)
  --artist      filter by artist
```

### `tags` — list available tags

```
waifu-im-client tags [flags]

Flags:
  -n, --name        filter by tag name
  -p, --page        page number
  -s, --page_size   page size
```

### `artists` — list registered artists

```
waifu-im-client artists [flags]

Flags:
  -n, --name        filter by artist name
  -p, --page        page number
  -s, --page_size   page size
```

### `albums` — list albums for a user

```
waifu-im-client albums [flags]

Flags:
  -u, --user        user ID or "me" (default: me)
  -p, --page        page number
  -s, --page_size   page size
```

### `album` — get album details

```
waifu-im-client album [flags]

Flags:
  -u, --usedID    user ID or "me" (default: me)
  -a, --albumID   album ID or "favorites" (default: favorites)
```

### `create-album` — create a new album

```
waifu-im-client create-album [flags]

Flags:
  -n, --name          album name (required)
  -d, --description   album description (required)
  -u, --usedID        user ID or "me" (default: me)
```

### `update-album` — update an existing album

```
waifu-im-client update-album [flags]

Flags:
  -a, --albumID       album ID (required)
  -n, --name          new album name (required)
  -d, --description   new album description (required)
  -u, --usedID        user ID or "me" (default: me)
```

### `delete-album` — delete an album

```
waifu-im-client delete-album [flags]

Flags:
  -a, --albumID   album ID (required)
  -u, --usedID    user ID or "me" (default: me)
```

## Skipped

- `/images` params: `ByteSize`, `GifOnly`, `OrderBy`, `Orientation`, `ManyFiles`, `Demonstration` — similar shape to implemented params, no new patterns
- `/tags` and `/artists` full param set — only name/page/pageSize implemented; remaining params follow the same shape
