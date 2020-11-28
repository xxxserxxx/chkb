package chkb

import (
	"errors"
	"fmt"

	evdev "github.com/gvalkov/golang-evdev"
	log "github.com/sirupsen/logrus"
)

type Mapper struct {
	LayerBook Book
	Layers    []*Layer
	downkeys  map[KeyCode]MapEvent
}

func NewMapper(book Book, initialLayer string) *Mapper {
	kb := &Mapper{
		LayerBook: book,
		Layers:    []*Layer{},
		downkeys:  map[KeyCode]MapEvent{},
	}
	kb.PushLayer(initialLayer)
	return kb
}

type KeyCode uint16

func (keyCode KeyCode) String() string {
	return evdev.KEY[int(keyCode)]
}

type KeyEvent struct {
	Action  KeyActions
	KeyCode KeyCode
}

func (ev KeyEvent) String() string {
	return fmt.Sprintf("%s-%v", evdev.KEY[int(ev.KeyCode)], ev.Action)
}

type MapEvent struct {
	Action  KbActions `yaml:"action,omitempty"`
	KeyCode KeyCode   `yaml:"keyCode,omitempty"`

	LayerName string `yaml:"layerName,omitempty"`
}

func (ev MapEvent) String() string {
	switch ev.Action {
	case KbActionUp, KbActionDown:
		return fmt.Sprintf("%s-%v", evdev.KEY[int(ev.KeyCode)], ev.Action)
	case KbActionPushLayer, KbActionPopLayer:
		return fmt.Sprintf("%v-%s", ev.Action, ev.LayerName)
	default:
		return fmt.Sprintf("%s-%v", evdev.KEY[int(ev.KeyCode)], ev.Action)
	}
}

func (layer *Layer) findMap(event KeyEvent) ([]MapEvent, bool) {
	keymap, ok := layer.KeyMap[event.KeyCode]
	if !ok {
		return nil, false
	}

	copymaps := make([]MapEvent, 0)
	//ActionMap
	if event.Action == KeyActionUp || event.Action == KeyActionDown {
		kmaps, ok := keymap[KeyActionMap]
		if ok {
			for i := range kmaps {
				m := kmaps[i]
				if m.Action == KbActionMap {
					switch event.Action {
					case KeyActionDown:
						m.Action = KbActionDown
					case KeyActionUp:
						m.Action = KbActionUp
					}
				}
				copymaps = append(copymaps, m)
			}
		}
	}

	kmaps, ok := keymap[event.Action]
	if ok {
		for i := range kmaps {
			m := kmaps[i]
			if m.Action == KbActionMap {
				switch event.Action {
				case KeyActionDown:
					m.Action = KbActionDown
				case KeyActionUp:
					m.Action = KbActionUp
				}
			}
			copymaps = append(copymaps, m)
		}
		return copymaps, true
	}
	return copymaps, len(copymaps) > 0
}

func (kb *Mapper) Maps(events []KeyEvent) ([]MapEvent, error) {
	mapped := make([]MapEvent, 0)
	for _, event := range events {
		switch event.Action {
		case KeyActionUp:
			downkey, ok := kb.downkeys[event.KeyCode]
			if ok {
				downkey.Action = KbActionUp
				mapped = append(mapped, downkey)
				delete(kb.downkeys, event.KeyCode)
			}
		}

		maps, err := kb.mapOne(event)
		if err != nil {
			log.
				WithField("event", event).
				WithError(err).
				Debug("Ignored event")
			continue
		}

		for _, m := range maps {
			switch m.Action {
			case KbActionDown:
				kb.downkeys[event.KeyCode] = m
				mapped = append(mapped, m)
			case KbActionUp:
				_, ok := kb.downkeys[m.KeyCode]
				if ok {
					mapped = append(mapped, m)
					delete(kb.downkeys, m.KeyCode)
				}
			default:
				mapped = append(mapped, m)
			}
		}

	}
	return mapped, nil
}

func (kb *Mapper) mapOne(event KeyEvent) ([]MapEvent, error) {
	for i := len(kb.Layers) - 1; i >= 0; i-- {
		kmaps, ok := kb.Layers[i].findMap(event)
		if !ok {
			continue
		}
		log.
			WithField("from", event).
			WithField("to", kmaps).
			Info("Map Key")
		return kmaps, nil
	}
	// No map detected, forward
	switch event.Action {
	case KeyActionUp, KeyActionDown:
		return []MapEvent{
			{
				Action:  KbActions(event.Action),
				KeyCode: event.KeyCode,
			},
		}, nil
	default:
		return nil, errors.New("Ignore event")
	}
}

func (kb *Mapper) Deliver(event MapEvent) (bool, error) {
	switch event.Action {
	case KbActionPushLayer:
		err := kb.PushLayer(event.LayerName)
		return true, err
	case KbActionPopLayer:
		err := kb.PopLayer(event.LayerName)
		return true, err
	default:
		return false, nil
	}
}

func (kb *Mapper) PushLayer(name string) error {
	log.Printf("Push layer %s", name)
	l, ok := kb.LayerBook[name]
	if !ok {
		return errors.New("Layer do not exist")
	}
	kb.Layers = append(kb.Layers, l)
	return nil
}

func (kb *Mapper) PopLayer(name string) error {
	log.Printf("Pop layer")
	if len(kb.Layers) == 1 {
		return errors.New("You cannot pop the last layer")
	}
	for i := range kb.Layers {
		if kb.Layers[i] == kb.LayerBook[name] {
			kb.Layers = append(kb.Layers[:i], kb.Layers[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("Layer %s not found", name)
}
