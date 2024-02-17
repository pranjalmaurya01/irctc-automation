package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"
	"github.com/joho/godotenv"
)

var currentTime = time.Now()
var tomorrow = currentTime.AddDate(0, 0, 1)

var src = "lko"
var dest = "ndls"
var jDate = tomorrow.Format("02/01/2006")

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

	godotenv.Load()
	username := os.Getenv("USERNAME")
	password := os.Getenv("PASSWORD")

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()
	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()

	var ok bool
	var captcha_base64 string
	chromedp.Run(ctx,
		chromedp.Navigate("https://www.irctc.co.in/nget/train-search"),

		// LOGIN
		chromedp.WaitVisible("a.loginText"),
		chromedp.Click("a.loginText", chromedp.ByQuery),
		chromedp.WaitReady(`input[formcontrolname="userid"]`),
		chromedp.SetValue(`input[formcontrolname="userid"]`, username, chromedp.ByQuery),
		chromedp.SetValue(`input[formcontrolname="password"]`, password, chromedp.ByQuery),
		chromedp.SetValue(`input[formcontrolname="password"]`, password, chromedp.ByQuery),
		chromedp.AttributeValue(`img.captcha-img`, "src", &captcha_base64, &ok),
	)

	png := strings.Split(captcha_base64, ",")[1]
	imageData, _ := base64.StdEncoding.DecodeString(png)
	os.WriteFile("temp.png", imageData, 0644)

	cmd := exec.Command("gocr", "temp.png")
	output, _ := cmd.CombinedOutput()
	capcha_val := string(output)

	// Login
	if err := chromedp.Run(ctx,
		chromedp.SendKeys(`input#captcha`, capcha_val, chromedp.ByQuery),
	); err != nil {
		fmt.Println(err)
	}

	if err := chromedp.Run(ctx,
		// ORIGIN SET
		chromedp.WaitReady("#origin input"),
		chromedp.Sleep(2*time.Second),
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
		chromedp.Sleep(1*time.Second),

		// JourneyQuota Set
		chromedp.Click(`p-dropdown#journeyQuota div`, chromedp.ByQuery),
		chromedp.Sleep(1*time.Second),
		chromedp.KeyEvent("t"),
		chromedp.KeyEvent(kb.Enter),

		// Submit Form
		chromedp.Click("button.search_btn.train_Search", chromedp.ByQuery),

		// Open Search

		chromedp.Sleep(5*time.Minute),
	); err != nil {
		fmt.Println(err)
	}

}
