package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
	"time"

	_ "embed"

	"github.com/google/subcommands"
	"github.com/progrium/macdriver/cocoa"
	"github.com/progrium/macdriver/core"
	"github.com/progrium/macdriver/objc"
	"github.com/progrium/macdriver/webkit"
	"github.com/progrium/watcher"
)

const (
	version = "0.2.0"
	docsURL = "http://github.com/progrium/topframe"
)

var (
	//go:embed index.html
	defaultIndex []byte
	//go:embed agent.plist
	agentPlist string
)

func init() {
	runtime.LockOSThread()
}

func main() {
	subcommands.Register(&agentCmd{}, "")
	subcommands.Register(&docsCmd{}, "")
	subcommands.Register(&versionCmd{}, "")
	subcommands.Register(subcommands.HelpCommand(), "")

	flag.Parse()

	status := subcommands.Execute(context.Background())
	os.Exit(int(status))
}

type agentCmd struct {
	plist bool
}

func (*agentCmd) Name() string     { return "agent" }
func (*agentCmd) Synopsis() string { return "fullscreen webview overlay agent" }
func (*agentCmd) Usage() string    { return "Usage: topframe agent [-plist]\n" }
func (c *agentCmd) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&c.plist, "plist", false, "generate agent plist")
}

func (c *agentCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	dir := ensureDir()

	if c.plist {
		generatePlist(dir)
		return subcommands.ExitSuccess
	}

	addr := startServer(dir)
	fw := startWatcher(dir)

	runApp(dir, addr, fw)

	return subcommands.ExitSuccess
}

func ensureDir() (dir string) {
	usr, err := user.Current()
	fatal(err)

	if os.Getenv("TOPFRAME_DIR") != "" {
		dir = os.Getenv("TOPFRAME_DIR")
	} else {
		dir = filepath.Join(usr.HomeDir, ".topframe")
	}

	os.MkdirAll(dir, 0755)

	if _, err := os.Stat(filepath.Join(dir, "index.html")); os.IsNotExist(err) {
		ioutil.WriteFile(filepath.Join(dir, "index.html"), defaultIndex, 0644)
	}

	return dir
}

func generatePlist(dir string) {
	tmpl, err := template.New("plist").Parse(agentPlist)
	fatal(err)

	p, err := exec.LookPath(os.Args[0])
	fatal(err)

	bin, _ := filepath.Abs(p)
	fatal(tmpl.Execute(os.Stdout, struct {
		Dir, Bin string
	}{
		Dir: dir,
		Bin: bin,
	}))
}

func startServer(dir string) *net.TCPAddr {
	srv := http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			dirpath := filepath.Join(dir, r.URL.Path)
			if isExecScript(dirpath) && r.Header.Get("Accept") == "text/event-stream" {
				streamExecScript(w, dirpath, strings.Split(r.URL.RawQuery, "+"))
				return
			}
			http.FileServer(http.Dir(dir)).ServeHTTP(w, r)
		}),
	}

	addr := ":0"
	if os.Getenv("TOPFRAME_ADDR") != "" {
		addr = os.Getenv("TOPFRAME_ADDR")
	}
	ln, err := net.Listen("tcp", addr)
	fatal(err)

	go srv.Serve(ln)

	return ln.Addr().(*net.TCPAddr)
}

func startWatcher(dir string) *watcher.Watcher {
	fw := watcher.New()
	fatal(fw.AddRecursive(dir))

	go fw.Start(450 * time.Millisecond)

	return fw
}

func runApp(dir string, addr *net.TCPAddr, fw *watcher.Watcher) {
	cocoa.TerminateAfterWindowsClose = false

	config := webkit.WKWebViewConfiguration_New()
	config.Preferences().SetValueForKey(core.True, core.String("developerExtrasEnabled"))

	url := core.URL(fmt.Sprintf("http://localhost:%d", addr.Port))
	req := core.NSURLRequest_Init(url)

	app := cocoa.NSApp_WithDidLaunch(func(_ objc.Object) {
		wv := webkit.WKWebView_Init(cocoa.NSScreen_Main().Frame(), config)
		wv.Retain()
		wv.SetOpaque(false)
		wv.SetBackgroundColor(cocoa.NSColor_Clear())
		wv.SetValueForKey(core.False, core.String("drawsBackground"))
		wv.LoadRequest(req)

		win := cocoa.NSWindow_Init(cocoa.NSScreen_Main().Frame(),
			cocoa.NSClosableWindowMask|cocoa.NSBorderlessWindowMask,
			cocoa.NSBackingStoreBuffered, false)
		win.SetContentView(wv)
		win.SetBackgroundColor(cocoa.NSColor_Clear())
		win.SetOpaque(false)
		win.SetTitleVisibility(cocoa.NSWindowTitleHidden)
		win.SetTitlebarAppearsTransparent(true)
		win.SetIgnoresMouseEvents(true)
		win.SetLevel(cocoa.NSMainMenuWindowLevel + 2)
		win.MakeKeyAndOrderFront(win)
		win.SetCollectionBehavior(cocoa.NSWindowCollectionBehaviorCanJoinAllSpaces)
		win.Send("setHasShadow:", false)

		statusBar := cocoa.NSStatusBar_System().StatusItemWithLength(cocoa.NSVariableStatusItemLength)
		statusBar.Retain()
		statusBar.Button().SetTitle("🔲")

		menuInteract := cocoa.NSMenuItem_New()
		menuInteract.Retain()
		menuInteract.SetTitle("Interactive")
		menuInteract.SetAction(objc.Sel("interact:"))
		cocoa.DefaultDelegateClass.AddMethod("interact:", func(_ objc.Object) {
			if win.IgnoresMouseEvents() {
				win.SetLevel(cocoa.NSMainMenuWindowLevel - 1)
				win.SetIgnoresMouseEvents(false)
				menuInteract.SetState(1)
			} else {
				win.SetIgnoresMouseEvents(true)
				win.SetLevel(cocoa.NSMainMenuWindowLevel + 2)
				menuInteract.SetState(0)
			}
		})

		menuEnabled := cocoa.NSMenuItem_New()
		menuEnabled.Retain()
		menuEnabled.SetTitle("Enabled")
		menuEnabled.SetState(1)
		menuEnabled.SetAction(objc.Sel("enabled:"))
		cocoa.DefaultDelegateClass.AddMethod("enabled:", func(_ objc.Object) {
			if win.IsVisible() {
				win.Send("orderOut:", win)
				menuInteract.SetEnabled(false)
				menuEnabled.SetState(0)
			} else {
				win.Send("orderFront:", win)
				menuInteract.SetEnabled(true)
				menuEnabled.SetState(1)
			}
		})

		menuSource := cocoa.NSMenuItem_New()
		menuSource.SetTitle("Show Source...")
		menuSource.SetAction(objc.Sel("source:"))
		cocoa.DefaultDelegateClass.AddMethod("source:", func(_ objc.Object) {
			go func() {
				fatal(exec.Command("open", dir).Run())
			}()
		})

		menuQuit := cocoa.NSMenuItem_New()
		menuQuit.SetTitle("Quit")
		menuQuit.SetAction(objc.Sel("terminate:"))

		menu := cocoa.NSMenu_New()
		menu.SetAutoenablesItems(false)
		menu.AddItem(menuEnabled)
		menu.AddItem(menuInteract)
		menu.AddItem(cocoa.NSMenuItem_Separator())
		menu.AddItem(menuSource)
		menu.AddItem(cocoa.NSMenuItem_Separator())
		menu.AddItem(menuQuit)

		statusBar.SetMenu(menu)

		go func() {
			for {
				select {
				case event := <-fw.Event:
					if event.IsDir() {
						continue
					}
					wv.Reload(nil)
				case <-fw.Closed:
					return
				}
			}
		}()
	})

	log.Printf("topframe %s from progrium.com\n", version)
	app.ActivateIgnoringOtherApps(true)
	app.Run()
}

type docsCmd struct{}

func (*docsCmd) Name() string             { return "docs" }
func (*docsCmd) Synopsis() string         { return "open documentation in browser" }
func (*docsCmd) Usage() string            { return "Usage: topframe docs\n" }
func (*docsCmd) SetFlags(f *flag.FlagSet) {}
func (p *docsCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	fatal(exec.Command("open", docsURL).Run())
	return subcommands.ExitSuccess
}

type versionCmd struct{}

func (*versionCmd) Name() string             { return "version" }
func (*versionCmd) Synopsis() string         { return "show version" }
func (*versionCmd) Usage() string            { return "Usage: topframe version\n" }
func (*versionCmd) SetFlags(f *flag.FlagSet) {}
func (p *versionCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	fmt.Println(version)
	return subcommands.ExitSuccess
}

func streamExecScript(w http.ResponseWriter, dirpath string, args []string) {
	flusher, ok := w.(http.Flusher)
	if !ok || !isExecScript(dirpath) {
		http.Error(w, "script unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	cmd := exec.Command(dirpath, args...)
	cmd.Stderr = os.Stderr
	r, _ := cmd.StdoutPipe()
	scanner := bufio.NewScanner(r)

	finished := make(chan bool)
	go func() {
		for scanner.Scan() {
			_, err := io.WriteString(w, fmt.Sprintf("data: %s\n\n", scanner.Text()))
			if err != nil {
				log.Println("script:", err)
				return
			}
			flusher.Flush()
		}
		if err := scanner.Err(); err != nil {
			log.Println("script:", err)
		}
		finished <- true
	}()

	if err := cmd.Run(); err != nil {
		log.Println(err)
	}
	<-finished
}

func isExecScript(dirpath string) bool {
	fi, err := os.Stat(dirpath)
	if err != nil {
		return false
	}
	return fi.Mode()&0111 != 0
}

func fatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
