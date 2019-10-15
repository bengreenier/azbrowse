package views

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/lawrencegripper/azbrowse/internal/pkg/eventing"
	"github.com/lawrencegripper/azbrowse/internal/pkg/handlers"
	"github.com/stuartleeks/colorjson"
	"github.com/stuartleeks/gocui"
)

// ItemWidget is response for showing the text response from the Rest requests
type ItemWidget struct {
	x, y        int
	w, h        int
	hideGuids   bool
	content     string
	contentType handlers.ExpanderResponseType
	view        *gocui.View
	g           *gocui.Gui
}

// NewItemWidget creates a new instance of ItemWidget
func NewItemWidget(x, y, w, h int, hideGuids bool, content string) *ItemWidget {
	return &ItemWidget{x: x, y: y, w: w, h: h, hideGuids: hideGuids, content: content}
}

// Layout draws the widget in the gocui view
func (w *ItemWidget) Layout(g *gocui.Gui) error {
	w.g = g

	v, err := g.SetView("itemWidget", w.x, w.y, w.x+w.w, w.y+w.h)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}
	v.Editable = true
	v.Wrap = true
	w.view = v
	width, height := v.Size()
	handlers.ItemWidgetHeight = height
	handlers.ItemWidgetWidth = width
	v.Clear()

	if w.content == "" {
		return nil
	}

	if w.hideGuids {
		if w.hideGuids {
			w.content = stripSecretVals(w.content)
		}
	}
	switch w.contentType {
	case handlers.ResponseJSON:
		d := json.NewDecoder(strings.NewReader(w.content))
		d.UseNumber()
		var obj interface{}
		err = d.Decode(&obj)
		if err != nil {
			eventing.SendStatusEvent(eventing.StatusEvent{
				InProgress: false,
				Failure:    true,
				Message:    "Failed to display as JSON: " + err.Error(),
				Timeout:    time.Duration(time.Second * 4),
			})
			fmt.Fprint(v, w.content)
			return nil
		}

		f := colorjson.NewFormatter()
		f.Indent = 2
		s, err := f.Marshal(obj)
		if err != nil {
			fmt.Fprint(v, w.content)
		} else {
			fmt.Fprint(v, string(s))
		}
	// TODO - add YAML colorisation here
	default:
		fmt.Fprint(v, w.content)
	}

	return nil
}

// PageDown move the view down a page
func (w *ItemWidget) PageDown() {
	_, maxHeight := w.view.Size()
	x, y := w.view.Origin()
	w.view.SetCursor(0, 0) //nolint: errcheck

	maxY := strings.Count(w.content, "\n")
	y = y + maxHeight
	if y > maxY {
		y = maxY
	}

	err := w.view.SetOrigin(x, y)
	if err != nil {
		eventing.SendStatusEvent(eventing.StatusEvent{
			InProgress: false,
			Failure:    true,
			Message:    "Failed to execute pagedown: " + err.Error(),
			Timeout:    time.Duration(time.Second * 4),
		})
	}
}

// PageUp move the view a page up
func (w *ItemWidget) PageUp() {
	_, maxHeight := w.view.Size()
	x, y := w.view.Origin()
	w.view.SetCursor(0, 0) //nolint: errcheck

	y = y - maxHeight
	// Check we haven't overshot
	if y < 0 {
		y = 0
	}
	err := w.view.SetOrigin(x, y)
	if err != nil {
		eventing.SendStatusEvent(eventing.StatusEvent{
			InProgress: false,
			Failure:    true,
			Message:    "Failed to execute pagedown: " + err.Error(),
			Timeout:    time.Duration(time.Second * 4),
		})
	}
}

// SetContent displays the string in the itemview
func (w *ItemWidget) SetContent(content string, contentType handlers.ExpanderResponseType, title string) {
	w.g.Update(func(g *gocui.Gui) error {
		w.content = content
		w.contentType = contentType
		// Reset the cursor and origin (scroll poisition)
		// so we don't start at the bottom of a new doc
		w.view.SetCursor(0, 0) //nolint: errcheck
		w.view.SetOrigin(0, 0) //nolint: errcheck
		w.view.Title = title
		return nil
	})
}

// GetContent returns the current content
func (w *ItemWidget) GetContent() string {
	return w.content
}

// GetContentType returns the current content type
func (w *ItemWidget) GetContentType() handlers.ExpanderResponseType {
	return w.contentType
}
