package filer

import (
	"log"
	"os"

	"github.com/anmitsu/goful/widget"
)

type layoutType int

const (
	layoutTile layoutType = iota
	layoutTileTop
	layoutTileBottom
	layoutOneline
	layoutOneColumn
	layoutFullscreen
)

// Workspace represents box storing directories.
type Workspace struct {
	*widget.Window
	Dirs   []*Directory `json:"directories"`
	Layout layoutType   `json:"layout"`
	Title  string       `json:"title"`
	Cursor int          `json:"cursor"`
}

// NewWorkspace returns workspace of specified size.
func NewWorkspace(x, y, width, height int, title string) *Workspace {
	return &Workspace{
		widget.NewWindow(x, y, width, height),
		[]*Directory{},
		layoutTile,
		title,
		0,
	}
}

func (w *Workspace) init4json(x, y, width, height int) {
	w.Window = widget.NewWindow(x, y, width, height)
}

// CreateDir adds the home directory to the head.
func (w *Workspace) CreateDir() {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	dir := NewDirectory(0, 0, 0, 0)
	dir.Chdir(home)
	w.Dirs = append(w.Dirs, nil)
	copy(w.Dirs[1:], w.Dirs[:len(w.Dirs)-1])
	w.Dirs[0] = dir
	w.SetCursor(0)
	w.allocate()
}

// CloseDir closes the focused directory.
func (w *Workspace) CloseDir() {
	if len(w.Dirs) < 2 {
		return
	}
	i := w.Cursor
	w.Dirs = append(w.Dirs[:i], w.Dirs[i+1:]...)
	if w.Cursor >= len(w.Dirs) {
		w.Cursor = len(w.Dirs) - 1
	}
	w.attach()
	w.allocate()
}

// ChdirNeighbor changes focused directory to neighbor directory path.
func (w *Workspace) ChdirNeighbor() {
	w.Dir().Chdir(w.NextDir().Path)
}

func (w *Workspace) visible(visible bool) {
	if visible {
		w.ReloadAll()
	} else {
		for _, d := range w.Dirs {
			d.ClearList()
		}
	}
}

// MoveCursor moves cursor with specified amounts.
func (w *Workspace) MoveCursor(amount int) {
	w.Cursor += amount
	if len(w.Dirs) <= w.Cursor {
		w.Cursor = 0
	} else if w.Cursor < 0 {
		w.Cursor = len(w.Dirs) - 1
	}
	w.attach()
}

// SetCursor sets cursor to specified position.
func (w *Workspace) SetCursor(x int) {
	w.Cursor = x
	if w.Cursor < 0 {
		w.Cursor = 0
	} else if w.Cursor > len(w.Dirs)-1 {
		w.Cursor = len(w.Dirs) - 1
	}
	w.attach()
}

func (w *Workspace) attach() {
	err := os.Chdir(w.Dir().Path)
	if err != nil {
		log.Fatalln(err)
	}
}

// ReloadAll reloads all directories.
func (w *Workspace) ReloadAll() {
	for _, d := range w.Dirs {
		d.reload()
	}
	err := os.Chdir(w.Dir().Path)
	if err != nil {
		log.Fatalln(err)
	}
}

// Dir returns focused directory.
func (w *Workspace) Dir() *Directory {
	return w.Dirs[w.Cursor]
}

// NextDir returns next directory.
func (w *Workspace) NextDir() *Directory {
	return w.Dirs[w.nextIndex()]
}

// PrevDir returns previous directory.
func (w *Workspace) PrevDir() *Directory {
	return w.Dirs[w.prevIndex()]
}

// SwapNextDir swaps focus and next for directories.
func (w *Workspace) SwapNextDir() {
	next := w.nextIndex()
	w.Dirs[w.Cursor], w.Dirs[next] = w.Dirs[next], w.Dirs[w.Cursor]
	w.MoveCursor(1)
	w.allocate()
}

// SwapPrevDir swaps focus and previous for directories.
func (w *Workspace) SwapPrevDir() {
	prev := w.prevIndex()
	w.Dirs[w.Cursor], w.Dirs[prev] = w.Dirs[prev], w.Dirs[w.Cursor]
	w.MoveCursor(-1)
	w.allocate()
}

func (w *Workspace) nextIndex() int {
	i := w.Cursor + 1
	if i >= len(w.Dirs) {
		return 0
	}
	return i
}

func (w *Workspace) prevIndex() int {
	i := w.Cursor - 1
	if i < 0 {
		return len(w.Dirs) - 1
	}
	return i
}

// SetTitle sets workspace title.
func (w *Workspace) SetTitle(title string) {
	w.Title = title
}

// LayoutTile allocates workspace layout to tile.
func (w *Workspace) LayoutTile() {
	w.Layout = layoutTile
	x, y := w.LeftTop()
	dlen := len(w.Dirs)
	if dlen < 2 {
		w.Dirs[0].Resize(x, y, w.Width(), w.Height())
		return
	}
	width := w.Width() / 2
	w.Dirs[0].Resize(x, y, width, w.Height())
	k := dlen - 1
	height := w.Height() / k
	hodd := w.Height() % k
	wodd := w.Width() % 2
	for i, d := range w.Dirs[1 : dlen-1] {
		d.Resize(x+width, y+height*i, width+wodd, height)
	}
	w.Dirs[dlen-1].Resize(x+width, y+height*(k-1), width+wodd, height+hodd)
}

// LayoutTileTop allocates workspace layout to tile top.
func (w *Workspace) LayoutTileTop() {
	w.Layout = layoutTileTop
	dlen := len(w.Dirs)
	x, y := w.LeftTop()
	if dlen < 2 {
		w.Dirs[0].Resize(x, y, w.Width(), w.Height())
	}
	height := w.Height() / 2
	hodd := w.Height() % 2
	w.Dirs[0].Resize(x, y+height, w.Width(), height+hodd)

	k := dlen - 1
	width := w.Width() / k
	for i, d := range w.Dirs[1 : dlen-1] {
		d.Resize(x+width*i, y, width, height)
	}

	wodd := w.Width() % 2
	w.Dirs[dlen-1].Resize(x+width*(k-1), y, width+wodd, height)
}

// LayoutTileBottom allocates workspace layout to tile bottom.
func (w *Workspace) LayoutTileBottom() {
	w.Layout = layoutTileBottom
	dlen := len(w.Dirs)
	x, y := w.LeftTop()
	if dlen < 2 {
		w.Dirs[0].Resize(x, y, w.Width(), w.Height())
		return
	}
	height := w.Height() / 2
	w.Dirs[0].Resize(x, y, w.Width(), height)

	k := dlen - 1
	width := w.Width() / k
	hodd := w.Height() % 2
	for i, d := range w.Dirs[1 : dlen-1] {
		d.Resize(x+width*i, y+height, width, height+hodd)
	}

	wodd := w.Width() % 2
	w.Dirs[dlen-1].Resize(x+width*(k-1), y+height, width+wodd, height+hodd)
}

// LayoutOneline allocates workspace layout to one line.
func (w *Workspace) LayoutOneline() {
	w.Layout = layoutOneline
	x, y := w.LeftTop()
	k := len(w.Dirs)
	width := w.Width() / k
	for i, d := range w.Dirs[:k-1] {
		d.Resize(x+width*i, y, width, w.Height())
	}
	wodd := w.Width() % k
	w.Dirs[k-1].Resize(x+width*(k-1), y, width+wodd, w.Height())
}

// LayoutOnecolumn allocates workspace layout to one column.
func (w *Workspace) LayoutOnecolumn() {
	w.Layout = layoutOneColumn
	x, y := w.LeftTop()
	k := len(w.Dirs)
	height := w.Height() / k
	for i, d := range w.Dirs[:k-1] {
		d.Resize(x, y+height*i, w.Width(), height)
	}
	hodd := w.Height() % k
	w.Dirs[k-1].Resize(x, y+height*(k-1), w.Width(), height+hodd)
}

// LayoutFullscreen allocates workspace layout to full screen.
func (w *Workspace) LayoutFullscreen() {
	w.Layout = layoutFullscreen
	for _, d := range w.Dirs {
		x, y := w.LeftTop()
		d.Resize(x, y, w.Width(), w.Height())
	}
}

func (w *Workspace) allocate() {
	switch w.Layout {
	case layoutTile:
		w.LayoutTile()
	case layoutTileTop:
		w.LayoutTileTop()
	case layoutTileBottom:
		w.LayoutTileBottom()
	case layoutOneline:
		w.LayoutOneline()
	case layoutOneColumn:
		w.LayoutOnecolumn()
	case layoutFullscreen:
		w.LayoutFullscreen()
	}
}

// Resize and layout allocates.
func (w *Workspace) Resize(x, y, width, height int) {
	w.Window.Resize(x, y, width, height)
	w.allocate()
}

// ResizeRelative relative resizes and layout allocates.
func (w *Workspace) ResizeRelative(x, y, width, height int) {
	w.Window.ResizeRelative(x, y, width, height)
	w.allocate()
}

// Draw implements for interface of widget.Box.
func (w *Workspace) Draw() {
	switch w.Layout {
	case layoutFullscreen:
		w.Dir().draw(true)
	default:
		for i, d := range w.Dirs {
			if i != w.Cursor {
				d.draw(false)
			}
		}
		w.Dir().draw(true)
	}
}