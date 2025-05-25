package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/fogleman/gg"
)

const (
	owner   = "ptenchikDota"
	repo    = "banner"
	path    = "banner.png"
	message = "update banner"
)

type APIResponse struct {
	Data struct {
		Avatar       string  `json:"avatar"`
		LastMatch    string  `json:"lastMatch"`
		Nickname     string  `json:"nickname"`
		OverallRank  int     `json:"overallRank"`
		Percentile   float64 `json:"percentile"`
		Rating       float64 `json:"rating"`
		Region       string  `json:"region"`
		RegionalRank int     `json:"regionalRank"`
		SteamID      int64   `json:"steamId"`
		WinLoss      struct {
			Losses  int     `json:"losses"`
			Total   int     `json:"total"`
			Winrate float64 `json:"winrate"`
			Wins    int     `json:"wins"`
		} `json:"winLoss"`
	} `json:"data"`
}

func downloadImage(url string) (image.Image, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	img, _, err := image.Decode(resp.Body)
	return img, err
}

func loadLocalImage(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, err := png.Decode(f)
	return img, err
}

func drawRoundedAvatar(dc *gg.Context, avatar image.Image, x, y, size float64) {
	dc.DrawCircle(x+size/2, y+size/2, size/2)
	dc.Clip()
	dc.DrawImageAnchored(avatar, int(x+size/2), int(y+size/2), 0.5, 0.5)
	dc.ResetClip()
}

func drawCenteredText(dc *gg.Context, text string, centerX, y float64) {
	dc.DrawStringAnchored(text, centerX, y, 0.5, 0.5)
}

func main() {
	token := os.Getenv("GH_TOKEN")
	if token == "" {
		log.Fatal("❌ GH_TOKEN is not set")
	}

	// Получаем данные с API
	url := "https://windrun.io/api/players/901021922"
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	var result APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Fatal(err)
	}

	// Загружаем изображения
	avatar, err := downloadImage(result.Data.Avatar)
	if err != nil {
		log.Fatal(err)
	}
	badge, err := loadLocalImage("./profile.png")
	if err != nil {
		log.Fatal(err)
	}

	// Подготовка канваса
	const width = 320
	const height = 200
	dc := gg.NewContext(width, height)
	dc.SetRGB(0.07, 0.07, 0.09)
	dc.Clear()

	// Параметры
	leftMargin := 15.0
	avatarSize := 150.0
	centerX := width / 2
	badgeX := float64(width - 158)
	badgeY := 5.0

	// Аватар
	drawRoundedAvatar(dc, avatar, leftMargin, 35, avatarSize)

	// Ник (крупно)
	if err := dc.LoadFontFace("./fonts/Roboto-Light.ttf", 18); err != nil {
		log.Fatal(err)
	}
	dc.SetRGB(1, 1, 1)
	drawCenteredText(dc, "Ptenchik", leftMargin+71, 18)

	// Регион
	if err := dc.LoadFontFace("./fonts/Roboto-Regular.ttf", 14); err != nil {
		log.Fatal(err)
	}
	dc.SetRGB(0.7, 0.7, 0.7)
	region := fmt.Sprintf("REGION %s", strings.ToUpper(result.Data.Region))
	drawCenteredText(dc, region, leftMargin+228, 95)

	// MATCHES
	if err := dc.LoadFontFace("./fonts/Roboto-Regular.ttf", 14); err != nil {
		log.Fatal(err)
	}
	dc.SetRGB(0.6, 0.6, 0.6)
	drawCenteredText(dc, "MATCHES", float64(centerX+82), 130)

	// Победы / Поражения
	if err := dc.LoadFontFace("./fonts/Roboto-Bold.ttf", 18); err != nil {
		log.Fatal(err)
	}
	dc.SetRGB(0.3, 1, 0.3)
	drawCenteredText(dc, strconv.Itoa(result.Data.WinLoss.Wins), float64(centerX+52), 150)

	dc.SetRGB(1, 0.3, 0.3)
	drawCenteredText(dc, strconv.Itoa(result.Data.WinLoss.Losses), float64(centerX+114), 150)

	// Winrate
	if err := dc.LoadFontFace("./fonts/Roboto-Regular.ttf", 14); err != nil {
		log.Fatal(err)
	}
	dc.SetRGB(0.8, 0.8, 0.8)
	winrateStr := fmt.Sprintf("%.2f%%", result.Data.WinLoss.Winrate*100)
	drawCenteredText(dc, winrateStr, float64(centerX+86), 172)

	// Минус между W и L
	if err := dc.LoadFontFace("./fonts/Roboto-Regular.ttf", 24); err != nil {
		log.Fatal(err)
	}
	dc.SetRGB(0.8, 0.8, 0.8)
	drawCenteredText(dc, "-", float64(centerX+84), 150)

	// Значок
	dc.DrawImageAnchored(badge, int(badgeX), int(badgeY), 0, 0)

	// Ранг
	if err := dc.LoadFontFace("./fonts/Roboto-SemiBold.ttf", 13); err != nil {
		log.Fatal(err)
	}

	dc.SetRGB(0.5, 0, 0.6)
	rankStr := fmt.Sprintf("#%d", result.Data.RegionalRank)
	drawCenteredText(dc, rankStr, badgeX+78, badgeY+20)

	// Рейтинг (большой)
	if err := dc.LoadFontFace("./fonts/Roboto-SemiBold.ttf", 28); err != nil {
		log.Fatal(err)
	}
	dc.SetRGB(1, 1, 1)
	ratingStr := fmt.Sprintf("%.0f", result.Data.Rating)
	drawCenteredText(dc, ratingStr, badgeX+80, badgeY+49)

	outFile, err := os.Create("banner.png")
	if err != nil {
		log.Fatal(err)
	}
	defer outFile.Close()
	png.Encode(outFile, dc.Image())

	fmt.Println("✅ Картинка создана: banner.png")

	// Загружаем на GitHub
	imageBytes, err := os.ReadFile("banner.png")
	if err != nil {
		log.Fatal(err)
	}

	if err := uploadToGitHubImage(token, owner, repo, path, message, imageBytes); err != nil {
		log.Fatal("❌ Ошибка загрузки на GitHub:", err)
	}
}

func uploadToGitHubImage(token, owner, repo, path, message string, imageData []byte) error {
	// Получаем текущий sha
	sha, err := getGitHubFileSHA(token, owner, repo, path)
	if err != nil {
		// только логируем — если файла нет, это нормально (будет create)
		fmt.Println("ℹ️ sha не найден (будет создан новый файл):", err)
	}

	// Кодируем изображение в base64
	content := base64.StdEncoding.EncodeToString(imageData)

	// Формируем тело запроса
	payload := map[string]interface{}{
		"message": message,
		"content": content,
	}
	if sha != "" {
		payload["sha"] = sha
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", owner, repo, path)
	req, err := http.NewRequest("PUT", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		b, _ := io.ReadAll(res.Body)
		return fmt.Errorf("GitHub upload failed: %s\n%s", res.Status, string(b))
	}

	fmt.Println("✅ Файл загружен в GitHub:", url)
	return nil
}

func getGitHubFileSHA(token, owner, repo, path string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", owner, repo, path)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode == 404 {
		return "", fmt.Errorf("file does not exist yet")
	}
	if res.StatusCode >= 400 {
		b, _ := io.ReadAll(res.Body)
		return "", fmt.Errorf("GitHub GET failed: %s\n%s", res.Status, string(b))
	}

	var result struct {
		SHA string `json:"sha"`
	}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.SHA, nil
}
