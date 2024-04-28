package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/charmbracelet/lipgloss"
	Browser "github.com/chromedp/chromedp"
)

var mu sync.Mutex

type Nettruyen struct {
	TruyenUrl string `json:"TruyenUrl"`
	SoChapDoc int    `json:"SoChapDoc"`
	TenTruyen string `json:"TenTruyen"`
}

type Configuration struct {
	TrangChuNettruyen string      `json:"TrangChuNettruyen"`
	TenDangNhap       string      `json:"TenDangNhap"`
	MatKhau           string      `json:"MatKhau"`
	Truyen            []Nettruyen `json:"Truyen"`
	DelayTime         int         `json:"DelayTime"`
}
type SharedData struct {
	mu        sync.Mutex
	WebStatus string
}

var sharedData = SharedData{
	WebStatus: "Không Rõ!",
}

func readConfig() Configuration {
	data, err := ioutil.ReadFile("./config.json")
	if err != nil {
		log.Fatalf("[LỖI] Không load được file config: %v", err)
	}

	var config Configuration
	err = json.Unmarshal(data, &config)
	if err != nil {
		log.Fatalf("[LỖI] Config lỗi cú pháp: %v", err)
	}
	return config
}

func setTitle(title string) {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	setConsoleTitleW := kernel32.NewProc("SetConsoleTitleW")
	_, _, _ = setConsoleTitleW.Call(uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(title))))
}

func appendLog(message string) {
	f, err := os.OpenFile("log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logMessage := fmt.Sprintf("%s - %s\n", timestamp, message)
	if _, err = f.WriteString(logMessage); err != nil {
		log.Fatal(err)
	}
	//	fmt.Println(logMessage)
}
func main() {
	setTitle("Auto Tu Tiên Ver 1.0")
	config := readConfig()
	go AutoCheckWebStatus(config.TrangChuNettruyen, 1*time.Second)
	ctx, cancel := TaskSetupBrowser()
	defer cancel()
	var wg sync.WaitGroup
	for idx, comic := range config.Truyen {
		wg.Add(1)
		go func(c Nettruyen, index int) {
			defer wg.Done()
			ctxChild, cancelCtx := Browser.NewContext(ctx, Browser.WithLogf(log.Printf))
			defer cancelCtx()

			if err := runTasks(ctxChild, config, c, index); err != nil {
				appendLog(fmt.Sprintf("[LỖI] Khi đọc %s: %v", c.TruyenUrl, err))
			} else {
				appendLog(fmt.Sprintf("[HOÀN THÀNH] Đã đọc xong %s.", c.TruyenUrl))
			}
		}(comic, idx)
	}
	wg.Wait()
	Browser.Cancel(ctx)
	cancel()
}
func TimBraveBrowser() (string, error) {
	switch runtime.GOOS {
	case "windows":
		return `C:\Program Files\BraveSoftware\Brave-Browser\Application\brave.exe`, nil
	case "linux":
		return "/usr/bin/brave-browser", nil
	case "darwin":
		return "/Applications/Brave Browser.app/Contents/MacOS/Brave Browser", nil
	default:
		return "", errors.New("hệ điều hành không được hỗ trợ")
	}
}
func TaskSetupBrowser() (context.Context, context.CancelFunc) {
	bravePath, err := TimBraveBrowser()
	if err != nil {
		log.Fatalf("Error finding Brave browser: %v", err)
	}
	debuggingPort := "9222"

	time.Sleep(2 * time.Second)

	opts := append(Browser.DefaultExecAllocatorOptions[:],
		Browser.NoDefaultBrowserCheck,
		//	chromedp.WindowSize(500, 500),
		Browser.ExecPath(bravePath),
		Browser.Flag("remote-debugging-port", debuggingPort),
		Browser.Flag("headless", false),
		Browser.Flag("disable-gpu", true),
		Browser.Flag("no-sandbox", true),
		Browser.Flag("disable-setuid-sandbox", true),
		Browser.Flag("disable-logging", true),
		Browser.Flag("disable-login-animations", true),
		Browser.Flag("disable-notifications", true),
		Browser.Flag("lang", "vi_VN"),
		Browser.Flag("start-maximized", true),
	)

	ctx, cancel := Browser.NewExecAllocator(context.Background(), opts...)
	ctx, cancel = Browser.NewContext(ctx, Browser.WithLogf(log.Printf))

	return ctx, cancel
}
func runTasks(ctx context.Context, config Configuration, comic Nettruyen, idx int) error {
	return Browser.Run(ctx,
		TaskDangNhap(config, comic, idx),
		TaskTheoDoi(comic.TruyenUrl, comic, idx),
		TaskDocTruyen(ctx, comic.SoChapDoc, config.DelayTime, comic, idx),
	)
}

func TaskDangNhap(config Configuration, comic Nettruyen, idx int) Browser.Tasks {
	return Browser.Tasks{
		Browser.Navigate(config.TrangChuNettruyen),
		Browser.ActionFunc(func(ctx context.Context) error {
			TextBox(comic.TenTruyen, 0, comic.SoChapDoc, idx, "[INFO] Mở trang chủ nettruyen.")
			appendLog("[INFO] Mở trang chủ nettruyen.")
			return nil
		}),
		Browser.WaitVisible(`#header li.login-link > a`, Browser.ByQuery),
		Browser.Click(`#header li.login-link > a`, Browser.ByQuery),
		Browser.ActionFunc(func(ctx context.Context) error {
			TextBox(comic.TenTruyen, 0, comic.SoChapDoc, idx, "[INFO] Click nút đăng nhập.")
			appendLog("[INFO] Click nút đăng nhập.")
			return nil
		}),
		Browser.WaitVisible(`#ctl00_mainContent_login1_LoginCtrl_UserName`, Browser.ByQuery),
		Browser.SendKeys(`#ctl00_mainContent_login1_LoginCtrl_UserName`, config.TenDangNhap, Browser.ByQuery),
		Browser.SendKeys(`#ctl00_mainContent_login1_LoginCtrl_Password`, config.MatKhau, Browser.ByQuery),
		Browser.Click(`#ctl00_mainContent_login1_LoginCtrl_Login`, Browser.ByQuery),
		Browser.ActionFunc(func(ctx context.Context) error {
			TextBox(comic.TenTruyen, 0, comic.SoChapDoc, idx, "[INFO] Bắt đầu đăng nhập.")
			appendLog("[INFO] Bắt đầu đăng nhập.")
			return nil
		}),
		Browser.WaitNotPresent(`#ctl00_mainContent_login1_LoginCtrl_Login`, Browser.ByQuery),
		Browser.ActionFunc(func(ctx context.Context) error {
			TextBox(comic.TenTruyen, 0, comic.SoChapDoc, idx, "[INFO] Đăng nhập thành công.")
			appendLog("[INFO] Đăng nhập thành công.")
			return nil
		}),
	}
}

func TaskTheoDoi(url string, comic Nettruyen, idx int) Browser.Tasks {
	return Browser.Tasks{
		Browser.Navigate(url),
		Browser.ActionFunc(func(ctx context.Context) error {
			TextBox(comic.TenTruyen, 0, comic.SoChapDoc, idx, "[INFO] Mở trang đọc truyện.")
			appendLog("[INFO] Mở trang đọc truyện.")
			var followButtonExists bool
			err := Browser.Run(ctx, Browser.Evaluate(`!!document.querySelector('i.fa.fa-heart')`, &followButtonExists))
			if err != nil {
				appendLog("[LỖI] Không thấy nút theo dõi.")
				TextBox(comic.TenTruyen, 0, comic.SoChapDoc, idx, "[LỖI] Không thấy nút theo dõi.")
				return err
			}

			if followButtonExists {
				err = Browser.Run(ctx, Browser.Click(`i.fa.fa-heart`, Browser.ByQuery))
				if err != nil {
					appendLog("[LỖI] Lỗi khi click nút theo dõi.")
					TextBox(comic.TenTruyen, 0, comic.SoChapDoc, idx, "[LỖI] Lỗi khi click nút theo dõi.")
					return err
				}
				appendLog("[INFO] Đã click vào nút theo dõi truyện.")
				TextBox(comic.TenTruyen, 0, comic.SoChapDoc, idx, "[INFO] Đã click vào nút theo dõi truyện.")
				return nil
			}
			var unfollowButtonExists bool
			err = Browser.Run(ctx, Browser.Evaluate(`!!document.querySelector('i.fa.fa-times')`, &unfollowButtonExists))
			if err != nil {
				appendLog("[INFO] Không thấy nút bỏ theo dõi.")
				TextBox(comic.TenTruyen, 0, comic.SoChapDoc, idx, "[INFO] Không thấy nút bỏ theo dõi.")
				return err
			}

			if unfollowButtonExists {
				appendLog("[INFO] Truyện này đã theo dõi rồi.")
				TextBox(comic.TenTruyen, 0, comic.SoChapDoc, idx, "[INFO] Truyện này đã theo dõi rồi.")
				return nil
			}
			appendLog("[INFO] Không tìm thấy nút theo dõi.")
			TextBox(comic.TenTruyen, 0, comic.SoChapDoc, idx, "[INFO] Không tìm thấy nút theo dõi.")
			return nil
		}),
	}
}
func NettruyenWebStatus(url string) string {
	client := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil
		},
	}

	resp, err := client.Get(url)
	if err != nil {
		return fmt.Sprintf("[LỖI] Thất bại: %v", err)
	}
	defer resp.Body.Close()
	return resp.Status
}
func AutoCheckWebStatus(url string, interval time.Duration) {
	for {
		status := NettruyenWebStatus(url)
		sharedData.mu.Lock()
		sharedData.WebStatus = status
		sharedData.mu.Unlock()
		time.Sleep(interval)
	}
}
func TextBox(title string, currentChapter, totalChapters, idx int, status string) {
	sharedData.mu.Lock()
	webStatus := sharedData.WebStatus
	sharedData.mu.Unlock()

	mu.Lock()
	defer mu.Unlock()

	style := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		PaddingLeft(2).PaddingRight(2).
		PaddingTop(1).PaddingBottom(1).
		Width(70).
		Align(lipgloss.Center).
		BorderForeground(lipgloss.Color("#6272a4")).
		Foreground(lipgloss.Color("#ffffff")).
		Background(lipgloss.Color("#282a36"))

	content := fmt.Sprintf("%s\nĐang Đọc Chap: %d/%d\nStatus: %s\nWeb Status: %s", title, currentChapter, totalChapters, status, webStatus)
	posY := idx*7 + 1
	fmt.Printf("\033[%d;1H%s\n", posY, style.Render(content))
}

func TaskDocTruyen(ctx context.Context, SoChapDoc, DelayTime int, TruyenDangDoc Nettruyen, idx int) Browser.Tasks {
	return Browser.Tasks{
		Browser.WaitVisible(`div.read-action > a:nth-of-type(1)`, Browser.ByQuery),
		Browser.Click(`div.read-action > a:nth-of-type(1)`, Browser.ByQuery),
		Browser.ActionFunc(func(innerCtx context.Context) error {
			appendLog(fmt.Sprintf("[INFO] Bắt đầu đọc truyện: %s", TruyenDangDoc.TenTruyen))

			for i := 0; i < SoChapDoc; i++ {
				TextBox(TruyenDangDoc.TenTruyen, i+1, SoChapDoc, idx, "[INFO] Đang đọc chap.")
				err := Browser.Run(innerCtx,
					Browser.WaitVisible(`a.next > em`, Browser.ByQuery),
					Browser.Sleep(time.Duration(DelayTime)*time.Millisecond),
					Browser.Click(`a.next > em`, Browser.ByQuery),
				)
				if err != nil {
					appendLog(fmt.Sprintf("[LỖI] Lỗi khi đọc %d của %s: %v", i+1, TruyenDangDoc.TenTruyen, err))
					TextBox(TruyenDangDoc.TenTruyen, i+1, SoChapDoc, idx, "[LỖI] Đã gặp lỗi.")
					return err
				}
			}

			appendLog(fmt.Sprintf("[HOÀN THÀNH] Thành công đọc hết bộ: %s", TruyenDangDoc.TenTruyen))
			TextBox(TruyenDangDoc.TenTruyen, SoChapDoc, SoChapDoc, idx, "[HOÀN THÀNH] Đã đọc hết bộ này rồi.")
			return nil
		}),
	}
}
