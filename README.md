# Tetris

Classic Tetris game built with Go + WebAssembly, playable in any modern browser on both desktop and mobile.

## Play

Open [https://puerkafei.github.io/tetris/](https://puerkafei.github.io/tetris/) in your browser and start playing immediately.

## How to Build

```bash
GOOS=js GOARCH=wasm go build -o tetris.wasm game.go
```

Serve with any HTTP server:

```bash
python3 -m http.server 8080
```

## Controls

### Keyboard (Desktop)

| Key | Action |
|:---:|:-------|
| `←` `→` | Move left/right |
| `↑` | Rotate |
| `↓` | Soft drop |
| `Space` | Hard drop |
| `P` | Pause/Resume |
| `R` | Restart (after Game Over) |

### Touch (Mobile)

| Gesture | Action |
|:-------:|:-------|
| Swipe left/right | Move |
| Swipe up | Rotate |
| Swipe down | Soft drop |
| Double tap | Hard drop |
| Tap | Rotate / Restart |

### Bottom Buttons

`◀` `↻` `▼` `⏬` `▶`

## Tech Stack

- **Language**: Go
- **Runtime**: WebAssembly (WASM)
- **Rendering**: HTML5 Canvas
- **Bridge**: wasm_exec.js (Go official)

## Version

v2026.5.25
