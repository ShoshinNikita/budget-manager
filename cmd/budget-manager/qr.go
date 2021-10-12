//nolint:forbidigo
package main

import (
	"fmt"
	"strings"

	qrcode "rsc.io/qr"

	"github.com/ShoshinNikita/budget-manager/internal/app"
	"github.com/ShoshinNikita/budget-manager/internal/logger"
	"github.com/ShoshinNikita/budget-manager/internal/web/totp"
)

func printQRForTOTP(cfg app.Config, log logger.Logger, username string) {
	const totpLabel = "Budget Manager"

	if len(cfg.Server.Auth.TOTPAuthSecrets) == 0 {
		log.Errorf("no TOTP secrets")
		return
	}

	secret, ok := cfg.Server.Auth.TOTPAuthSecrets.Get(username)
	if !ok {
		log.Errorf("no secret for %q", username)
		return
	}

	url, err := totp.FormatURL(secret, totpLabel, username)
	if err != nil {
		log.WithError(err).Error("couldn't format TOTP url")
		return
	}

	qr, err := qrcode.Encode(url, qrcode.M)
	if err != nil {
		log.WithError(err).Error("couldn't generate QR code")
		return
	}

	printTitle(username, qr.Size, padding{
		top:    "\n",
		bottom: "\n",
		left:   "    ",
	})
	printQR(qr, padding{
		top:    "\n",
		bottom: "\n\n",
		left:   "    ",
	})
}

type padding struct {
	top    string
	bottom string
	left   string
}

func printTitle(username string, frameLength int, p padding) {
	const (
		topLeft     = "┌"
		topRight    = "┐"
		bottomLeft  = "└"
		bottomRight = "┘"
		vertical    = "│"
		horizontal  = "─"
	)

	text := fmt.Sprintf("QR code for %q", username)
	frameLength -= 2

	fmt.Print(p.top)

	fmt.Print(p.left)
	fmt.Print(topLeft)
	fmt.Print(strings.Repeat(horizontal, frameLength))
	fmt.Print(topRight)

	fmt.Print("\n")

	fmt.Print(p.left)
	fmt.Print(vertical)
	fmt.Print(text + strings.Repeat(" ", frameLength-len([]rune(text))))
	fmt.Print(vertical)

	fmt.Print("\n")

	fmt.Print(p.left)
	fmt.Print(bottomLeft)
	fmt.Print(strings.Repeat(horizontal, frameLength))
	fmt.Print(bottomRight)

	fmt.Print(p.bottom)
}

func printQR(qr *qrcode.Code, p padding) {
	const (
		fullBlock       = "█"
		halfTopBlock    = "▀"
		halfBottomBlock = "▄"
	)

	fmt.Print(p.top)
	for y := 0; y < qr.Size; y += 2 {
		fmt.Print(p.left)

		for x := 0; x < qr.Size; x++ {
			top := qr.Black(x, y)
			bottom := qr.Black(x, y+1)

			switch {
			case top && bottom:
				fmt.Print(fullBlock)
			case top:
				fmt.Print(halfTopBlock)
			case bottom:
				fmt.Print(halfBottomBlock)
			default:
				fmt.Print(" ")
			}
		}
		fmt.Print("\n")
	}
	fmt.Print(p.bottom)
}
