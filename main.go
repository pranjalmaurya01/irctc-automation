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

	vision "cloud.google.com/go/vision/apiv1"
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
	chromedp.Flag("enable-automation", true),

	chromedp.WindowSize(1350, 1000),
)

func main() {
	godotenv.Load()

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()
	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()

	chromedp.Run(ctx,
		chromedp.Navigate("https://www.irctc.co.in/nget/train-search"),
	)

	FillLoginForm(ctx)
	fmt.Println("login success")

	FillOriginField(ctx)
	FillDestField(ctx)
	FillJourneyDate(ctx)
	FillJourneyQuota(ctx)

	if err := chromedp.Run(ctx,
		chromedp.KeyEvent(kb.Enter),
		chromedp.Click("button.search_btn.train_Search", chromedp.ByQuery),
	); err != nil {
		fmt.Println(err)
	}

	if err := chromedp.Run(ctx,
		chromedp.Sleep(5*time.Minute),
	); err != nil {
		fmt.Println(err)
	}

}

func detectText() (string, error) {
	ctx := context.Background()

	client, err := vision.NewImageAnnotatorClient(ctx)
	if err != nil {
		return "", err
	}

	f, err := os.Open("temp.png")
	if err != nil {
		return "", err
	}
	defer f.Close()

	image, err := vision.NewImageFromReader(f)
	if err != nil {
		return "", err
	}
	annotations, err := client.DetectTexts(ctx, image, nil, 1)
	if err != nil {
		return "", err
	}

	if len(annotations) > 0 {
		return annotations[0].Description, nil
	}

	return "", nil

}

var captcha_count = 0

func SolveCaptcha(ctx context.Context) {
	captcha_count++
	var ok bool
	var captcha_base64 string
	chromedp.Run(ctx,
		chromedp.WaitVisible("img.captcha-img"),
		chromedp.AttributeValue(`img.captcha-img`, "src", &captcha_base64, &ok),
	)

	// save captcha in png file
	png := strings.Split(captcha_base64, ",")[1]
	imageData, _ := base64.StdEncoding.DecodeString(png)
	os.WriteFile("temp.png", imageData, 0644)

	var output string

	if captcha_count < 2 {
		cmd := exec.Command("gocr", "temp.png")
		out, cmd_err := cmd.CombinedOutput()
		output = string(out)
		if cmd_err != nil {
			log.Fatal("please install `gocr` package")
		}
	} else if captcha_count >= 2 {
		out, err := detectText()
		if err != nil {
			fmt.Println(err)
			SolveCaptcha(ctx)
			return
		}
		output = out
	} else if captcha_count > 4 {
		panic("Unable to solve captcha")
	}

	fmt.Println("captcha", captcha_count, output)

	chromedp.Run(ctx,
		chromedp.WaitVisible("img.captcha-img"),
		chromedp.Focus(`input#captcha`),
		chromedp.SendKeys(`input#captcha`, output, chromedp.ByQuery),
		chromedp.KeyEvent(kb.Enter),
		// chromedp.WaitNotVisible(`div.my-loading.ng-star-inserted`),
	)

	var is_err string
	chromedp.Run(ctx,
		chromedp.InnerHTML(`div.loginError`, &is_err),
	)

	if is_err == "Invalid Captcha...." {
		SolveCaptcha(ctx)
		return
	}
	if is_err == "Invalid User" {
		FillLoginForm(ctx)
	}
}

func FillLoginForm(ctx context.Context) {
	username := os.Getenv("USERNAME")
	password := os.Getenv("PASSWORD")

	chromedp.Run(ctx,
		// LOGIN
		chromedp.WaitVisible("a.loginText"),
		chromedp.Click("a.loginText", chromedp.ByQuery),
		chromedp.WaitReady(`input[formcontrolname="userid"]`),
		chromedp.SetValue(`input[formcontrolname="userid"]`, username, chromedp.ByQuery),
		chromedp.SetValue(`input[formcontrolname="password"]`, password, chromedp.ByQuery),
		chromedp.SetValue(`input[formcontrolname="password"]`, password, chromedp.ByQuery),
	)

	SolveCaptcha(ctx)
}

func FillOriginField(ctx context.Context) {
	chromedp.Run(ctx,
		chromedp.WaitReady("#origin input"),
		chromedp.Sleep(1*time.Second),
		chromedp.SendKeys(`#origin input`, src, chromedp.ByQuery),
		chromedp.Sleep(1*time.Second),
		chromedp.Click(`#pr_id_1_list span`),
	)
}

func FillDestField(ctx context.Context) {
	chromedp.Run(ctx,
		chromedp.WaitReady("#destination input"),
		chromedp.Sleep(1*time.Second),
		chromedp.SendKeys(`#destination input`, dest, chromedp.ByQuery),
		chromedp.Sleep(1*time.Second),
		chromedp.Click(`#pr_id_2_list span`),
	)
}

func FillJourneyDate(ctx context.Context) {
	chromedp.Run(ctx,
		chromedp.WaitReady("#jDate input"),
		chromedp.Click("#jDate input"),
		chromedp.SetValue("#jDate input", ""),
		chromedp.SendKeys("#jDate input", jDate),
		chromedp.KeyEvent(kb.Escape),
		chromedp.Sleep(1*time.Second),
	)
}

func FillJourneyQuota(ctx context.Context) {
	chromedp.Run(ctx,
		chromedp.WaitReady("p-dropdown#journeyQuota div"),
		chromedp.Click(`p-dropdown#journeyQuota div`, chromedp.ByQuery),
		chromedp.Sleep(1*time.Second),
		chromedp.KeyEvent("t"),
	)
}
