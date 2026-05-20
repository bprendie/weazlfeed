# WeazlFeed

![WeazlFeed terminal interface](weazlfeed.png)

Burn the algorithmic timeline and bypass the podcast-industrial complex.
WeazlFeed rips RSS, Atom, and straight-up 1993 Gopher payloads directly into
the terminal. It strips the HTML bloat, kills the tracking pixels, and renders
pure signal. Got a podcast? The MP3 gets piped instantly to `mpv` with a
reactive EQ. Point your local GPU at the text to filter out SEO sludge or
interrogate massive articles on the fly.

No browser tabs. No walled gardens. Just the raw feed on the bare metal.

## Current Status: The Sovereign Reader

- **Encrypted Metal:** Local-first RSS and Atom reader backed by a bcrypt-locked
  SQLite vault.
- **1993 Payloads:** Native Gopher reader for dialing `gopher://` menus and
  ripping pure text.
- **The Audio Pipe:** Podcast enclosures drop straight into `mpv`. Playhead
  checkpoints save your spot automatically.
- **No Fake Bars:** When audio is active, an `ffmpeg`-powered Harmonica EQ
  visualizer renders the actual stream dynamics.
- **Ghost Search:** Podcast directory search built directly into the TUI. No
  paid API keys required.
- **Tactical AI:** Local LLM integration for interrogating articles, ripping
  tactical summaries, and bouncing SEO sludge.
- **BBS TUI:** Three-pane terminal layout with focused pane expansion.

## Forge The Binary

Linux/macOS:

```sh
git clone https://github.com/bprendie/weazlfeed.git
cd weazlfeed
./scripts/install.sh
```

Windows PowerShell:

```powershell
git clone https://github.com/bprendie/weazlfeed.git
cd weazlfeed
powershell -ExecutionPolicy Bypass -File .\scripts\install.ps1
```

No wizards. No corporate installers. The install script compiles the core
engine and the CLI utilities, tucks them into the app bin directory, and adds
that directory to your user `PATH`.

Installed binaries:

- `weazlfeed`
- `weazlfeed-setup`
- `weazlfeed-import`
- `weazlfeed-refresh`
- `weazlfeed-podcast-search`
- `weazlfeed-prune`
- `weazlfeed-vault`

Requirements:

- Go 1.25+
- SQLite through `github.com/mattn/go-sqlite3`
- C toolchain for CGO (`go-sqlite3`)
- `mpv`
- `ffmpeg`
- Optional: Ollama or a vLLM/OpenAI-compatible local endpoint

On Windows, `scripts\install.ps1` installs missing dependencies with `winget`:
Git, Go, MSYS2/UCRT64 GCC, FFmpeg, and mpv. It installs WeazlFeed under:

```text
%APPDATA%\WeazlFeed\bin
%APPDATA%\WeazlFeed\config
%APPDATA%\WeazlFeed\vaults
```

If this is the first time those tools were installed, open a new PowerShell
after the installer so Windows picks up the refreshed `PATH`.

If `mpv`, `ffmpeg`, or your local model are missing or offline, the text reader
survives and keeps working. It fails closed.

## Boot Sequence & The Vault

```sh
weazlfeed
```

On first launch, WeazlFeed demands a vault password. Later launches require
that password before the UI even renders. The database is locked down to `0600`
permissions, and payloads are encrypted at rest.

Warning: There is no cloud recovery flow. You forget the password, you lose the
database. Back up your metal.

Force a vault unlock/encryption migration without refreshing feeds:

```sh
weazlfeed-vault
```

Need to reconfigure endpoints?

```sh
weazlfeed-setup
```

## The Command Deck

Slash commands are dead here. The BBS relies on hotkeys. Hit `ctrl+k` in the
TUI for the full matrix.

| Key | Action |
| --- | --- |
| `j` / `k` | Move through the focused pane. |
| `PgUp` / `PgDn` | Page the focused pane. |
| `Home` / `End` | Jump to top/bottom of the focused pane. |
| `tab` | Manually switch focus and expand that pane. |
| `enter` | Open source, read item, dial Gopher target, or play audio. |
| `esc` / `left` | Move back one pane, or kill the active audio process. |
| `space` | Pick/drop sources in the left rail; pause/resume audio elsewhere. |
| `<` / `>` | Seek active audio back 10s / forward 30s. |
| `p` | Open the podcast directory. |
| `a` | Add a feed or Gopher URL to the selected folder. |
| `n` | Create a folder in the selected section. |
| `f` | Mark selected podcast episode as finished. |
| `ctrl+d` | Delete selected source after confirmation. |
| `r` | Refresh the selected source. |
| `R` | Force refresh all sources. |
| `h` | Hide/show items flagged as sludge by the Bouncer. |
| `ctrl+a` | Interrogate the local AI about the active item. |
| `ctrl+t` | Command the local AI to extract a 3-point tactical summary. |
| `ctrl+b` | Open the Bouncer rule desk. |
| `q` / `ctrl+c` | Kill the app. |

## The Signal: Feeds & Gopher

On first run, WeazlFeed drops a starter deck seeded from the config. Your moves
stick: seed refreshes do not overwrite your folder organization.

Organization: hit `space` to pick up a source, move it, and hit `space` to drop
it. Press `a` to drop a new URL. `gopher://` URLs automatically route to the
Gopher engine. Audio feeds route to Podcasts.

Import/export: OPML files are explicitly ignored by Git so your private feed
lists stay strictly local.

```sh
weazlfeed-import feeds.opml
```

Surfing Gopher: directory menus render as nested item lists. Hitting `enter`
dials the target. `esc` walks you back up the stack. It is exactly what the
internet felt like before the banners took over.

Ghost Search Podcasts: bypass Apple's front door entirely. Search the directory
and add a feed directly from the CLI, or hit `p` inside the TUI to browse and
subscribe.

```sh
weazlfeed-podcast-search "darknet diaries"
weazlfeed-podcast-search -add 1 "darknet diaries"
```

## Tactical AI Execution

AI is wired as a local workbench, not a recommendation engine. It only touches
the item you ask it to touch. Long articles are trimmed before they hit the
model so a 20-token-per-second box does not lock up the TUI.

**Interrogation (`ctrl+a`):** Ask a question about the active article. The
answer, prompt, source title, and URL are saved as an encrypted local artifact.
These appear in their own section in Sources. Open one, continue the
questioning with `ctrl+a`, or nuke it with `ctrl+d`.

**Tactical Triage (`ctrl+t`):** Extracts the three most important technical
points from the active article. Results are cached in the encrypted vault, so
reopening the summary is instant.

**The Bouncer (`ctrl+b`):** Opens the Bouncer desk. Add (`n`) or delete (`d`)
local prompts that decide whether new items are sludge, like `Flag SEO filler
with no primary source`. During refresh, newly inserted items are scanned. Hit
`h` to hide the garbage.

During AI work, WeazlFeed spins up the status phrases and token context bar
from the WeazlChat flow so you know the local model is grinding. If the model
is offline, AI commands fail closed instead of blocking the reader.

## The Amp: Audio

Items with audio enclosures are not text. Hitting `enter` triggers a centered
playback window.

WeazlFeed spawns `mpv` directly. No heavy CGO wrappers. It periodically writes
your `playhead_seconds` to SQLite. If you crash out, it saves your spot.

Podcast states:

- `NEW`: Untouched.
- `LISTENING`: Has a saved playhead.
- `FINISHED`: Marked complete. Hit `f` to manually tag it.

If `ffmpeg` is available, the EQ is driven by a real decoded audio meter. No
fake bars. The Harmonica spring renderer gives the raw frequency signal
physical weight.

## Overrides & Diagnostics

Paths:

- Linux/macOS config: `~/.config/weazlfeed/config.json`
- Linux/macOS SQLite vault: `~/.local/share/weazlfeed/weazlfeed.sqlite3`
- Windows config: `%APPDATA%\WeazlFeed\config\config.json`
- Windows SQLite vault: `%APPDATA%\WeazlFeed\vaults\weazlfeed.sqlite3`

Override these at runtime if you are managing multiple identities:

```sh
WEAZLFEED_CONFIG=/path/to/config.json weazlfeed
WEAZLFEED_DATA=/path/to/data-dir weazlfeed
```

PowerShell:

```powershell
$env:WEAZLFEED_CONFIG="C:\path\to\config.json"; weazlfeed
$env:WEAZLFEED_DATA="C:\path\to\vault-dir"; weazlfeed
```

Diagnostics:

- `mpv not found`: Audio dies. Text reader lives.
- `ffmpeg not found`: EQ dies. Audio lives.
- Model endpoint connection refused: AI features are disabled until you spin up
  your provider.
- Forgotten vault password: You are cooked. Replace the database.

Development:

```sh
go test ./...
go build ./cmd/weazlfeed
```
