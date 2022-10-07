package utils

import (
	"crypto/md5"
	"encoding/hex"

	"github.com/charmbracelet/lipgloss"
)

// HashStringMD5 hashes a string with md5.
func HashStringMD5(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

func TextColor(text string, textColor string) string {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color(textColor))
	return style.Render(text)
}
