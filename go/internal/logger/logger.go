package logger

import (
	"fmt"
	"log"

	"github.com/fatih/color"
)

type Level int

const (
	milestone Level = iota
	info
	warn
	softError
	hardError
)

func (l Level) color(event string) string {
	switch l {
	case milestone:
		return color.GreenString(event)
	case info:
		return color.CyanString(event)
	case warn:
		return color.YellowString(event)
	case softError:
		return color.MagentaString(event)
	case hardError:
		return color.RedString(event)
	default:
		return event
	}
}

func formatPrefix(prefix string) string {
	if prefix == "" {
		return ""
	}
	return fmt.Sprintf("[%s] ", prefix)
}

func write(prefix string, level Level, event, detail string) {
	prefix = formatPrefix(prefix)
	event = level.color(event)

	if detail == "" {
		log.Print(prefix + event)
		return
	}
	log.Print(prefix + event + " | " + detail)
}

func Milestone(event, detail string) { write("", milestone, event, detail) }
func Info(event, detail string)      { write("", info, event, detail) }
func Warn(event, detail string)      { write("", warn, event, detail) }
func SoftError(event, detail string) { write("", softError, event, detail) }
func HardError(event, detail string) { write("", hardError, event, detail) }

func Player(id uint64) string {
	return fmt.Sprintf("player=%d", id)
}

func PlayerWithAddr(id uint64, addr string) string {
	return fmt.Sprintf("player=%d addr=%s", id, addr)
}

func PlayerPair(playerId0, playerId1 uint64) string {
	return fmt.Sprintf("%s vs %s", Player(playerId0), Player(playerId1))
}

func PairPrefix(playerId0, playerId1 uint64) string {
	return fmt.Sprintf("%d vs %d", playerId0, playerId1)
}

type Logger struct {
	prefix string
}

func ForPair(playerId0, playerId1 uint64) *Logger {
	return &Logger{prefix: PairPrefix(playerId0, playerId1)}
}

func (l *Logger) Milestone(event, detail string) { write(l.prefix, milestone, event, detail) }
func (l *Logger) Info(event, detail string)      { write(l.prefix, info, event, detail) }
func (l *Logger) Warn(event, detail string)      { write(l.prefix, warn, event, detail) }
func (l *Logger) SoftError(event, detail string) { write(l.prefix, softError, event, detail) }
func (l *Logger) HardError(event, detail string) { write(l.prefix, hardError, event, detail) }
