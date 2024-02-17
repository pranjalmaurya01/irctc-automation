package main

import (
	"context"
	"log"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"
)

var username = ""
var password = ""

var src = "lko"
var dest = "ndls"
var jDate = "18/02/2024"

var opts = append(chromedp.DefaultExecAllocatorOptions[:],
	chromedp.Flag("headless", false),
	chromedp.Flag("disable-gpu", false),
	chromedp.Flag("enable-automation", false),
	chromedp.Flag("disable-extensions", false),
	// chromedp.Flag("blink-settings", "imagesEnabled=false"),
	chromedp.Flag("enable-automation", true),

	chromedp.WindowSize(1350, 1000),
)

func main() {
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()
	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()

	chromedp.Run(ctx,
		chromedp.Navigate("https://www.irctc.co.in/nget/train-search"),

		// LOGIN
		chromedp.WaitReady("a.loginText"),
		chromedp.Click("a.loginText", chromedp.ByQuery),
		chromedp.WaitReady(`input[formcontrolname="userid"]`),
		chromedp.SetValue(`input[formcontrolname="userid"]`, username, chromedp.ByQuery),
		chromedp.SetValue(`input[formcontrolname="password"]`, password, chromedp.ByQuery),
		chromedp.Focus(`input[formcontrolname="captcha"]`),
		chromedp.WaitNotPresent(`input[formcontrolname="userid"]`),

		// ORIGIN SET
		chromedp.WaitReady("#origin input"),
		chromedp.SendKeys(`#origin input`, src, chromedp.ByQuery),
		chromedp.Sleep(1*time.Second),
		chromedp.KeyEvent(kb.Tab),

		// DEST SET
		chromedp.SendKeys(`#destination input`, dest, chromedp.ByQuery),
		chromedp.Sleep(1*time.Second),
		chromedp.KeyEvent(kb.Tab),

		// JourneyDate Set
		chromedp.KeyEvent(jDate),
		chromedp.KeyEvent(kb.Escape),

		// JourneyQuota Set
		chromedp.Click(`p-dropdown#journeyQuota div`, chromedp.ByQuery),
		chromedp.KeyEvent("T"),
		chromedp.KeyEvent(kb.Enter),

		// Submit Form
		chromedp.Click("button.search_btn.train_Search", chromedp.ByQuery),

		// Open Search

		chromedp.Sleep(5*time.Minute),
	)

}
